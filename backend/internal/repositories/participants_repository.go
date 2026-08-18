package repositories

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

var (
	ErrInvalidJoinCode = errors.New(
		"reserva grupal no encontrada",
	)

	ErrParticipantIneligible = errors.New(
		"la cuenta debe estar activa y tener RUT registrado",
	)

	ErrParticipationWindowClosed = errors.New(
		"el periodo para modificar participantes ya termino",
	)
)

// normalizeJoinCode establece una representación canónica del código.
//
// Los códigos que generemos posteriormente serán en mayúsculas.
// Esto permite que "abcd12" y "ABCD12" sean tratados como el mismo código.
func normalizeJoinCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// codeHash genera el SHA-256 almacenado en reservations.join_code_hash.
//
// El código real nunca se persiste en PostgreSQL.
func codeHash(code string) string {
	sum := sha256.Sum256(
		[]byte(normalizeJoinCode(code)),
	)

	return hex.EncodeToString(sum[:])
}

// generateJoinCode crea un código de unión aleatorio y legible para reservas grupales.
//
// Se excluyen caracteres visualmente ambiguos como I, O, 0 y 1 para reducir
// errores al copiar o ingresar el código manualmente.
//
// El código se genera usando crypto/rand y se entrega con el formato:
//
//	ABCDE-FGHIJ
//
// El código original solo se devuelve al usuario al crear la reserva.
// PostgreSQL almacena únicamente su hash SHA-256 mediante codeHash().
func generateJoinCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	const length = 10

	code := make([]byte, length)
	max := big.NewInt(int64(len(alphabet)))

	for i := range code {
		value, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}

		code[i] = alphabet[value.Int64()]
	}

	// Más fácil de leer/copiar:
	// ABCDE-FGHIJ
	return string(code[:5]) + "-" + string(code[5:]), nil
}

// GetReservationProgress devuelve el estado actual de una reserva grupal.
//
// No modifica información.
//
// Permite conocer:
//
//   - cantidad actual de participantes;
//   - mínimo requerido;
//   - capacidad;
//   - estado de la reserva;
//   - condición HEALTHY / AT_RISK / PENDING_MINIMUM;
//   - si el usuario actual participa;
//   - si es owner.
func GetReservationProgress(
	code string,
	userID int,
) (models.ReservationProgress, error) {

	var progress models.ReservationProgress

	if normalizeJoinCode(code) == "" {
		return progress, ErrInvalidJoinCode
	}

	err := database.DB.QueryRowContext(
		context.Background(),
		`
		SELECT
			r.id,
			r.status,

			COUNT(pa.id)
				FILTER (
					WHERE pa.status = 'CONFIRMED'
				),

			p.minimum_participants,
			r.group_capacity_snapshot,

			EXISTS (
				SELECT 1
				FROM participants mine
				WHERE mine.reservation_id = r.id
				  AND mine.user_id = $2
				  AND mine.status = 'CONFIRMED'
			) AS is_member,

			EXISTS (
				SELECT 1
				FROM participants mine
				WHERE mine.reservation_id = r.id
				  AND mine.user_id = $2
				  AND mine.is_owner = true
			) AS is_owner

		FROM reservations r

		INNER JOIN reservation_policies p
			ON p.id = r.policy_id

		LEFT JOIN participants pa
			ON pa.reservation_id = r.id

		WHERE r.join_code_hash = $1
		  AND r.group_capacity_snapshot IS NOT NULL
		  AND r.status IN ('PENDING', 'CONFIRMED')

		GROUP BY
			r.id,
			r.status,
			p.minimum_participants,
			r.group_capacity_snapshot
		`,
		codeHash(code),
		userID,
	).Scan(
		&progress.ReservationID,
		&progress.Status,
		&progress.ParticipantCount,
		&progress.MinimumParticipants,
		&progress.Capacity,
		&progress.IsMember,
		&progress.IsOwner,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return progress, ErrInvalidJoinCode
	}

	if err != nil {
		return progress, err
	}

	progress.GroupCondition =
		participantGroupCondition(
			progress.Status,
			progress.ParticipantCount,
			progress.MinimumParticipants,
		)

	return progress, nil
}

// ChangeParticipation confirma o retira un participante.
//
// La transacción bloquea primero la fila reservations.
//
// Esto serializa las modificaciones de participantes de una misma reserva,
// evitando que dos usuarios puedan sobrepasar simultáneamente la capacidad.
//
// Regla MVP2:
//
//	PENDING + alcanza mínimo
//		-> CONFIRMED
//
//	CONFIRMED + baja del mínimo
//		-> continúa CONFIRMED
//		-> GroupCondition = AT_RISK
//
// No se generan infracciones aquí.
//
// Los retiros realizados dentro de late_withdrawal_minutes se registran
// como LATE_WITHDRAWAL en reservation_participant_audit para que el futuro
// módulo de infracciones pueda procesarlos.
func ChangeParticipation(
	code string,
	userID int,
	confirm bool,
) (models.ReservationProgress, error) {

	ctx := context.Background()

	if normalizeJoinCode(code) == "" {
		return models.ReservationProgress{},
			ErrInvalidJoinCode
	}

	tx, err := database.DB.BeginTx(
		ctx,
		&sql.TxOptions{
			Isolation: sql.LevelSerializable,
		},
	)

	if err != nil {
		return models.ReservationProgress{}, err
	}

	defer tx.Rollback()

	var (
		reservationID                int
		reservationStatus            string
		startTime                    time.Time
		capacity                     int
		minimum                      int
		confirmationDeadlineMinutes  int
		lateWithdrawalMinutes        int
		groupRecoveryDeadlineMinutes int
		dbNow                        time.Time
	)

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT
			r.id,
			r.status,
			r.start_time,
			r.group_capacity_snapshot,
			p.minimum_participants,
			p.confirmation_deadline_minutes,
			p.late_withdrawal_minutes,
			p.group_recovery_deadline_minutes,
			CURRENT_TIMESTAMP

		FROM reservations r

		INNER JOIN reservation_policies p
			ON p.id = r.policy_id

		WHERE r.join_code_hash = $1
		  AND r.group_capacity_snapshot IS NOT NULL
		  AND r.status IN ('PENDING', 'CONFIRMED')

		FOR UPDATE OF r
		`,
		codeHash(code),
	).Scan(
		&reservationID,
		&reservationStatus,
		&startTime,
		&capacity,
		&minimum,
		&confirmationDeadlineMinutes,
		&lateWithdrawalMinutes,
		&groupRecoveryDeadlineMinutes,
		&dbNow,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return models.ReservationProgress{},
			ErrInvalidJoinCode
	}

	if err != nil {
		return models.ReservationProgress{}, err
	}

	// Después del comienzo de la reserva ya no puede modificarse
	// la composición del grupo.
	if !startTime.After(dbNow) {
		return models.ReservationProgress{},
			ErrParticipationWindowClosed
	}

	// Para ingresar al grupo usamos dos ventanas:
	//
	// PENDING:
	//     confirmation_deadline_minutes
	//
	// CONFIRMED:
	//     group_recovery_deadline_minutes
	//
	// Esto permite que un grupo que quedó AT_RISK pueda buscar reemplazos
	// hasta el límite configurado.
	if confirm {

		deadlineMinutes :=
			confirmationDeadlineMinutes

		if reservationStatus ==
			models.ReservationStatusConfirmed {

			deadlineMinutes =
				groupRecoveryDeadlineMinutes
		}

		deadline :=
			startTime.Add(
				-time.Duration(deadlineMinutes) *
					time.Minute,
			)

		if !dbNow.Before(deadline) {
			return models.ReservationProgress{},
				ErrParticipationWindowClosed
		}
	}

	// ---------------------------------------------------------------------
	// Validación del usuario.
	// ---------------------------------------------------------------------

	var (
		rut       string
		isBlocked bool
	)

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT
			COALESCE(rut, ''),
			is_blocked

		FROM users

		WHERE id = $1

		FOR UPDATE
		`,
		userID,
	).Scan(
		&rut,
		&isBlocked,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return models.ReservationProgress{},
			ErrParticipantIneligible
	}

	if err != nil {
		return models.ReservationProgress{}, err
	}

	if isBlocked || strings.TrimSpace(rut) == "" {
		return models.ReservationProgress{},
			ErrParticipantIneligible
	}

	// ---------------------------------------------------------------------
	// Participación previa.
	// ---------------------------------------------------------------------

	var (
		oldParticipantStatus string
		isOwner              bool
	)

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT
			status,
			is_owner

		FROM participants

		WHERE reservation_id = $1
		  AND user_id = $2

		FOR UPDATE
		`,
		reservationID,
		userID,
	).Scan(
		&oldParticipantStatus,
		&isOwner,
	)

	participantExists := err == nil

	if err != nil &&
		!errors.Is(err, sql.ErrNoRows) {

		return models.ReservationProgress{}, err
	}

	// ---------------------------------------------------------------------
	// Cantidad actual confirmada.
	// ---------------------------------------------------------------------

	var confirmedCount int

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT COUNT(*)
		FROM participants
		WHERE reservation_id = $1
		  AND status = 'CONFIRMED'
		`,
		reservationID,
	).Scan(&confirmedCount)

	if err != nil {
		return models.ReservationProgress{}, err
	}

	// ---------------------------------------------------------------------
	// Regla pura de negocio.
	// ---------------------------------------------------------------------

	mutate,
		newParticipantStatus,
		newReservationStatus,
		transitionErr :=
		participantTransition(
			participantExists,
			oldParticipantStatus,
			isOwner,
			confirmedCount,
			capacity,
			minimum,
			reservationStatus,
			confirm,
		)

	if transitionErr != nil {
		return models.ReservationProgress{},
			transitionErr
	}

	// ---------------------------------------------------------------------
	// Operación idempotente / no-op.
	// ---------------------------------------------------------------------

	if !mutate {

		if err := tx.Commit(); err != nil {
			return models.ReservationProgress{}, err
		}

		return models.ReservationProgress{
			ReservationID: reservationID,

			Status: reservationStatus,

			GroupCondition: participantGroupCondition(
				reservationStatus,
				confirmedCount,
				minimum,
			),

			ParticipantCount: confirmedCount,

			MinimumParticipants: minimum,

			Capacity: capacity,

			IsMember: participantExists &&
				oldParticipantStatus ==
					"CONFIRMED",

			IsOwner: participantExists &&
				isOwner,
		}, nil
	}

	// ---------------------------------------------------------------------
	// Modificación de participant.
	// ---------------------------------------------------------------------

	if !participantExists {

		_, err = tx.ExecContext(
			ctx,
			`
			INSERT INTO participants (
				reservation_id,
				user_id,
				status,
				is_owner,
				confirmed_at
			)
			VALUES (
				$1,
				$2,
				$3::varchar,
				false,
				CASE
					WHEN $3::varchar = 'CONFIRMED'
					THEN $4::timestamptz
					ELSE NULL
				END
			)
			`,
			reservationID,
			userID,
			newParticipantStatus,
			dbNow,
		)

	} else {

		_, err = tx.ExecContext(
			ctx,
			`
			UPDATE participants

			SET
				status = $3::varchar,

				confirmed_at =
					CASE
						WHEN $3::varchar = 'CONFIRMED'
						THEN $4::timestamptz
						ELSE confirmed_at
					END

			WHERE reservation_id = $1
			  AND user_id = $2
			`,
			reservationID,
			userID,
			newParticipantStatus,
			dbNow,
		)
	}

	if err != nil {
		return models.ReservationProgress{}, err
	}

	// ---------------------------------------------------------------------
	// Recalculamos cantidad después de la modificación.
	// ---------------------------------------------------------------------

	var newConfirmedCount int

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT COUNT(*)
		FROM participants
		WHERE reservation_id = $1
		  AND status = 'CONFIRMED'
		`,
		reservationID,
	).Scan(&newConfirmedCount)

	if err != nil {
		return models.ReservationProgress{}, err
	}

	// ---------------------------------------------------------------------
	// Actualización del estado de la reserva.
	//
	// La regla B está incorporada en participantTransition:
	//
	// CONFIRMED nunca vuelve automáticamente a PENDING.
	// ---------------------------------------------------------------------

	if newReservationStatus != reservationStatus {

		_, err = tx.ExecContext(
			ctx,
			`
			UPDATE reservations
			SET status = $2
			WHERE id = $1
			`,
			reservationID,
			newReservationStatus,
		)

		if err != nil {
			return models.ReservationProgress{}, err
		}
	}

	// ---------------------------------------------------------------------
	// Auditoría.
	// ---------------------------------------------------------------------

	action := "CONFIRM"

	if !confirm {

		action = "WITHDRAW"

		if lateWithdrawalMinutes > 0 {

			lateBoundary :=
				startTime.Add(
					-time.Duration(
						lateWithdrawalMinutes,
					) * time.Minute,
				)

			if !dbNow.Before(lateBoundary) {
				action = "LATE_WITHDRAWAL"
			}
		}
	}

	var previousParticipantStatus sql.NullString

	if participantExists {
		previousParticipantStatus =
			sql.NullString{
				String: oldParticipantStatus,
				Valid:  true,
			}
	}

	_, err = tx.ExecContext(
		ctx,
		`
		INSERT INTO reservation_participant_audit (
			reservation_id,
			actor_user_id,
			participant_user_id,
			action,
			previous_status,
			new_status,
			previous_reservation_status,
			new_reservation_status
		)
		VALUES (
			$1,
			$2,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7
		)
		`,
		reservationID,
		userID,
		action,
		previousParticipantStatus,
		newParticipantStatus,
		reservationStatus,
		newReservationStatus,
	)

	if err != nil {
		return models.ReservationProgress{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.ReservationProgress{}, err
	}

	return models.ReservationProgress{
		ReservationID: reservationID,

		Status: newReservationStatus,

		GroupCondition: participantGroupCondition(
			newReservationStatus,
			newConfirmedCount,
			minimum,
		),

		ParticipantCount: newConfirmedCount,

		MinimumParticipants: minimum,

		Capacity: capacity,

		IsMember: newParticipantStatus ==
			"CONFIRMED",

		IsOwner: isOwner,
	}, nil
}

// GetReservationParticipants devuelve la composición del grupo.
//
// La autorización de quién puede consultar esta información se realizará
// posteriormente en service/handler.
func GetReservationParticipants(
	reservationID int,
) ([]models.ReservationParticipant, error) {

	rows, err := database.DB.QueryContext(
		context.Background(),
		`
		SELECT
			p.user_id,
			u.full_name,
			u.email,
			COALESCE(u.rut, ''),
			p.is_owner,
			p.status,
			p.confirmed_at,
			p.created_at,
			p.updated_at

		FROM participants p

		INNER JOIN users u
			ON u.id = p.user_id

		WHERE p.reservation_id = $1

		ORDER BY
			p.is_owner DESC,
			p.created_at ASC,
			p.id ASC
		`,
		reservationID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	participants :=
		[]models.ReservationParticipant{}

	for rows.Next() {

		var (
			participant models.ReservationParticipant
			confirmedAt sql.NullTime
		)

		err := rows.Scan(
			&participant.UserID,
			&participant.FullName,
			&participant.Email,
			&participant.RUT,
			&participant.IsOwner,
			&participant.Status,
			&confirmedAt,
			&participant.CreatedAt,
			&participant.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if confirmedAt.Valid {

			value :=
				confirmedAt.Time.In(
					businessclock.Location(),
				)

			participant.ConfirmedAt =
				&value
		}

		participant.CreatedAt =
			participant.CreatedAt.In(
				businessclock.Location(),
			)

		participant.UpdatedAt =
			participant.UpdatedAt.In(
				businessclock.Location(),
			)

		participants =
			append(
				participants,
				participant,
			)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return participants, nil
}
