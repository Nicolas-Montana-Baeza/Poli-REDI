package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

var (
	ErrResourceNotAllowedByPolicy = errors.New("el recurso no esta permitido por la politica vigente")
	ErrReservationForbidden       = errors.New("no tienes permisos para cancelar esta reserva")
	ErrReservationNotCancellable  = errors.New("la reserva ya no se puede cancelar")
	ErrReservationFinalized       = errors.New("no puedes cancelar una reserva finalizada")
)

type RequestFrequencyError struct {
	NextDate time.Time
}

func (e RequestFrequencyError) Error() string {
	return fmt.Sprintf("ya tienes una solicitud vigente; proxima fecha permitida: %s", e.NextDate.Format("2006-01-02"))
}

const reservationColumns = `
	r.id, r.policy_id, r.user_id, r.resource_id, r.activity_id,
	r.start_time, r.duration_minutes, r.status,
	COALESCE(r.cancellation_reason, ''), r.created_at, r.updated_at,
	COALESCE(a.name, 'Reserva') AS activity_name,
	res.name AS resource_name,
	COALESCE(u.full_name, '') AS user_full_name,
	COALESCE(u.email, '') AS user_email,
	COALESCE(u.rut, '') AS user_rut,

	-- ------------------------------------------------------------
	-- Estado grupal MVP2.
	-- ------------------------------------------------------------
	--
	-- group_capacity_snapshot identifica una reserva grupal y conserva
	-- la capacidad vigente al momento de su creación.
	r.group_capacity_snapshot,

	-- El minimo queda congelado en la reserva. El fallback permite leer
	-- filas historicas creadas antes de PG16_0010.
	COALESCE(
		r.group_minimum_participants_snapshot,
		p.minimum_participants,
		0
	),

	-- Solo las participaciones CONFIRMED forman parte del grupo activo.
	COALESCE((
		SELECT COUNT(*)
		FROM participants rp
		WHERE rp.reservation_id = r.id
		  AND rp.status = 'CONFIRMED'
	), 0)
`

const reservationJoins = `
	FROM reservations r
	INNER JOIN resources res ON res.id = r.resource_id
	INNER JOIN users u ON u.id = r.user_id
	INNER JOIN reservation_policies p ON p.id = r.policy_id
	LEFT JOIN activities a ON a.id = r.activity_id`

type reservationScanner interface {
	Scan(...any) error
}

type reservationQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type reservationExecer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func scanReservation(scanner reservationScanner) (models.Reservation, error) {
	var reservation models.Reservation
	var activityName, resourceName, userFullName, userEmail, userRUT string
	var groupCapacity sql.NullInt64
	var minimumParticipants int
	var participantCount int
	err := scanner.Scan(
		&reservation.ID,
		&reservation.PolicyID,
		&reservation.UserID,
		&reservation.ResourceID,
		&reservation.ActivityID,
		&reservation.StartTime,
		&reservation.DurationMinutes,
		&reservation.Status,
		&reservation.CancellationReason,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
		&activityName,
		&resourceName,
		&userFullName,
		&userEmail,
		&userRUT,
		&groupCapacity,
		&minimumParticipants,
		&participantCount,
	)
	if err != nil {
		return models.Reservation{}, err
	}

	reservation.StartTime = reservation.StartTime.In(businessclock.Location())
	reservation.CreatedAt = reservation.CreatedAt.In(businessclock.Location())
	reservation.UpdatedAt = reservation.UpdatedAt.In(businessclock.Location())
	reservation.Hour = reservation.StartTime.Format("15:04")
	reservation.Title = activityName
	reservation.Type = mapReservationType(reservation.Status)
	reservation.ResourceName = resourceName
	reservation.UserFullName = userFullName
	reservation.UserEmail = userEmail
	reservation.UserRUT = userRUT
	// ------------------------------------------------------------
	// Metadatos de reserva grupal MVP2.
	// ------------------------------------------------------------
	//
	// Las reservas normales mantienen estos campos en sus valores cero,
	// preservando el contrato de MVP1.
	if groupCapacity.Valid {
		capacity := int(groupCapacity.Int64)

		reservation.IsGroupReservation = true
		reservation.Capacity = &capacity
		reservation.MinimumParticipants = minimumParticipants
		reservation.ParticipantCount = participantCount
		reservation.GroupCondition = participantGroupCondition(
			reservation.Status,
			participantCount,
			minimumParticipants,
		)
	}
	return reservation, nil
}

func scanReservationRows(rows *sql.Rows) ([]models.Reservation, error) {
	defer rows.Close()
	reservations := []models.Reservation{}
	for rows.Next() {
		reservation, err := scanReservation(rows)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, reservation)
	}
	return reservations, rows.Err()
}

func GetAllReservations() ([]models.Reservation, error) {
	rows, err := database.DB.QueryContext(context.Background(), `SELECT `+reservationColumns+reservationJoins+` ORDER BY r.start_time ASC`)
	if err != nil {
		return nil, err
	}
	return scanReservationRows(rows)
}

// ExpirePendingGroupReservations cancela solicitudes grupales que llegaron a
// su confirmation deadline sin alcanzar el mínimo congelado en la reserva.
//
// La operación es atómica e idempotente. El filtro por status=PENDING permite
// que PostgreSQL vuelva a evaluar la fila si una incorporación concurrente la
// confirmó mientras este UPDATE esperaba su bloqueo.
func ExpirePendingGroupReservations(now time.Time) (int64, error) {
	return expirePendingGroupReservations(
		context.Background(),
		database.DB,
		now,
	)
}

func expirePendingGroupReservations(
	ctx context.Context,
	execer reservationExecer,
	now time.Time,
) (int64, error) {
	result, err := execer.ExecContext(
		ctx,
		`
		UPDATE reservations reservation
		SET
			status = 'CANCELLED',
			cancellation_reason = $2
		FROM reservation_policies policy
		WHERE policy.id = reservation.policy_id
		  AND reservation.status = 'PENDING'
		  AND reservation.group_capacity_snapshot IS NOT NULL
		  AND reservation.start_time
			  - make_interval(
					mins => policy.confirmation_deadline_minutes
				)
			  <= $1
		  AND (
			SELECT COUNT(*)
			FROM participants participant
			WHERE participant.reservation_id = reservation.id
			  AND participant.status = 'CONFIRMED'
		  ) < COALESCE(
			reservation.group_minimum_participants_snapshot,
			policy.minimum_participants
		  )
		`,
		now,
		models.CancellationReasonMinimumNotMet,
	)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func GetReservationsByUserID(userID int) ([]models.Reservation, error) {
	rows, err := database.DB.QueryContext(context.Background(), `SELECT `+reservationColumns+reservationJoins+` WHERE r.user_id = $1 ORDER BY r.start_time DESC`, userID)
	if err != nil {
		return nil, err
	}
	return scanReservationRows(rows)
}

func GetReservationByID(id int) (models.Reservation, error) {
	return getReservationByID(context.Background(), database.DB, id)
}

func getReservationByID(ctx context.Context, q reservationQueryer, id int) (models.Reservation, error) {
	return scanReservation(q.QueryRowContext(ctx, `SELECT `+reservationColumns+reservationJoins+` WHERE r.id = $1`, id))
}

func GetActiveReservationsForAvailability(from, to time.Time, userID int, includeAllOpenUse bool) ([]models.Reservation, error) {
	rows, err := database.DB.QueryContext(context.Background(), `
		SELECT `+reservationColumns+reservationJoins+`
		WHERE r.status IN ('PENDING', 'CONFIRMED')
		  AND r.start_time < $2
		  AND r.end_time > $1
		  AND ($3::boolean OR r.reservation_mode_snapshot = 'RESERVABLE' OR r.user_id = $4)
		ORDER BY r.start_time ASC`, from, to, includeAllOpenUse, userID)
	if err != nil {
		return nil, err
	}
	return scanReservationRows(rows)
}

func GetAvailabilityBlocks(from, to time.Time) ([]models.AvailabilityBlock, error) {
	rows, err := database.DB.QueryContext(context.Background(), `
		SELECT b.id, b.resource_id, b.block_type, COALESCE(b.reason, ''),
		       b.start_time, b.end_time, res.name
		FROM availability_blocks b
		INNER JOIN resources res ON res.id = b.resource_id
		WHERE b.is_active = true
		  AND b.start_time < $2
		  AND b.end_time > $1
		ORDER BY b.start_time ASC`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	blocks := []models.AvailabilityBlock{}
	for rows.Next() {
		var block models.AvailabilityBlock
		if err := rows.Scan(&block.ID, &block.ResourceID, &block.BlockType, &block.Reason, &block.StartTime, &block.EndTime, &block.ResourceName); err != nil {
			return nil, err
		}
		block.StartTime = block.StartTime.In(businessclock.Location())
		block.EndTime = block.EndTime.In(businessclock.Location())
		blocks = append(blocks, block)
	}
	return blocks, rows.Err()
}

func GetCurrentReservationPolicy() (models.ReservationPolicy, error) {
	return GetCurrentReservationPolicyComplete()
}

func GetLatestConsumingReservation(userID int) (time.Time, int, error) {
	var createdAt time.Time
	var frequencyDays int
	err := database.DB.QueryRowContext(context.Background(), `
		SELECT r.created_at, p.request_frequency_days
		FROM reservations r
		INNER JOIN reservation_policies p ON p.id = r.policy_id
		WHERE r.user_id = $1
		  AND r.status IN ('PENDING', 'CONFIRMED')
		  AND r.reservation_mode_snapshot = 'RESERVABLE'
		ORDER BY ((r.created_at AT TIME ZONE 'America/Santiago')::date + p.request_frequency_days) DESC, r.id DESC
		LIMIT 1`, userID).Scan(&createdAt, &frequencyDays)
	return createdAt.In(businessclock.Location()), frequencyDays, err
}

// AddReservationWithPolicy crea una reserva utilizando una fotografía
// de la política de reservas vigente al momento de la solicitud.
//
// La operación completa se ejecuta dentro de una transacción SERIALIZABLE
// para mantener consistencia frente a solicitudes concurrentes.
//
// El flujo realiza las siguientes responsabilidades:
//
//   - obtiene y bloquea de forma compartida la política vigente;
//   - valida que el recurso esté permitido por esa política;
//   - aplica las reglas dinámicas de horario, duración y ventana reservable;
//   - obtiene el modo y capacidad actual del recurso;
//   - aplica la restricción de frecuencia únicamente a recursos RESERVABLE;
//   - determina si el recurso utiliza el flujo de reserva grupal;
//   - crea una reserva normal o grupal según corresponda;
//   - en reservas grupales genera un join code seguro;
//   - almacena únicamente el hash SHA-256 del join code;
//   - guarda un snapshot de la capacidad del recurso;
//   - registra al solicitante como participante y owner del grupo.
//
// Para evitar condiciones de carrera entre solicitudes simultáneas del mismo
// usuario se utiliza pg_advisory_xact_lock() dentro de la transacción.
//
// Las reservas grupales comienzan en PENDING mientras no alcancen
// minimum_participants. El solicitante cuenta como el primer participante.
//
// El join code original se devuelve únicamente como parte de la respuesta
// de creación y nunca se almacena directamente en PostgreSQL.
func AddReservationWithPolicy(
	reservation models.Reservation,
	validate func(models.ReservationPolicy) error,
) (models.Reservation, error) {

	ctx := context.Background()

	// La creación de una reserva involucra múltiples lecturas y escrituras
	// relacionadas. SERIALIZABLE evita que operaciones concurrentes puedan
	// observar estados incompatibles durante la evaluación de las reglas.
	tx, err := database.DB.BeginTx(
		ctx,
		&sql.TxOptions{
			Isolation: sql.LevelSerializable,
		},
	)
	if err != nil {
		return models.Reservation{}, err
	}

	// Si ocurre cualquier error antes del Commit, la transacción completa
	// se revierte automáticamente.
	defer tx.Rollback()

	// Serializa las solicitudes de reserva del mismo usuario.
	//
	// Esto es especialmente importante para la regla de frecuencia:
	// dos solicitudes concurrentes del mismo usuario no deben poder pasar
	// simultáneamente la validación antes de que una de ellas sea persistida.
	if _, err := tx.ExecContext(
		ctx,
		`SELECT pg_advisory_xact_lock(73002, $1)`,
		reservation.UserID,
	); err != nil {
		return models.Reservation{}, err
	}

	// ------------------------------------------------------------
	// Política vigente.
	// ------------------------------------------------------------
	//
	// La reserva siempre queda asociada a la versión exacta de la política
	// utilizada al momento de su creación. Las modificaciones posteriores
	// de las reglas no alteran retroactivamente esta reserva.

	policy, err := scanPolicy(
		tx.QueryRowContext(
			ctx,
			`
			SELECT `+policyColumns+`
			FROM reservation_policies
			WHERE is_published = true
			  AND effective_from <= CURRENT_TIMESTAMP
			  AND (
			      effective_to IS NULL
			      OR effective_to > CURRENT_TIMESTAMP
			  )
			ORDER BY effective_from DESC, id DESC
			LIMIT 1
			FOR SHARE
			`,
		),
	)

	if err != nil {
		return models.Reservation{}, err
	}

	// Las duraciones permitidas y recursos asociados forman parte de la
	// misma política versionada y deben cargarse dentro de la transacción.
	if err := loadPolicyCollections(
		ctx,
		tx,
		&policy,
	); err != nil {
		return models.Reservation{}, err
	}

	// Un recurso puede existir y estar activo, pero aun así no estar
	// habilitado por la política vigente.
	if !policyAllowsResource(
		policy,
		reservation.ResourceID,
	) {
		return models.Reservation{},
			ErrResourceNotAllowedByPolicy
	}

	// Ejecuta las validaciones de negocio que dependen de la política,
	// como horario, duración y ventana reservable.
	if err := validate(policy); err != nil {
		return models.Reservation{}, err
	}

	// ------------------------------------------------------------
	// Recurso.
	// ------------------------------------------------------------
	//
	// reservation_mode determina el comportamiento de la reserva.
	// capacity se utiliza como límite y snapshot para reservas grupales.

	var (
		mode     string
		capacity sql.NullInt64
	)

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT
			reservation_mode,
			capacity
		FROM resources
		WHERE id = $1
		  AND is_active = true
		`,
		reservation.ResourceID,
	).Scan(
		&mode,
		&capacity,
	)

	if err != nil {
		return models.Reservation{}, err
	}

	// ------------------------------------------------------------
	// Restricción de frecuencia.
	// ------------------------------------------------------------
	//
	// Solo las reservas de recursos RESERVABLE consumen la frecuencia
	// configurada por la política.
	//
	// Los recursos OPEN_USE pueden utilizarse sin bloquear al usuario
	// para futuras solicitudes de cancha.

	if mode == "RESERVABLE" {

		var (
			previousCreatedAt time.Time
			frequencyDays     int
		)

		err := tx.QueryRowContext(
			ctx,
			`
			SELECT
				r.created_at,
				p.request_frequency_days
			FROM reservations r
			INNER JOIN reservation_policies p
				ON p.id = r.policy_id
			WHERE r.user_id = $1
			  AND r.status IN ('PENDING', 'CONFIRMED')
			  AND r.reservation_mode_snapshot = 'RESERVABLE'
			ORDER BY
				(
					(r.created_at AT TIME ZONE 'America/Santiago')::date
					+ p.request_frequency_days
				) DESC,
				r.id DESC
			LIMIT 1
			`,
			reservation.UserID,
		).Scan(
			&previousCreatedAt,
			&frequencyDays,
		)

		if err != nil &&
			!errors.Is(err, sql.ErrNoRows) {
			return models.Reservation{}, err
		}

		if err == nil {

			localCreated :=
				previousCreatedAt.In(
					businessclock.Location(),
				)

			// La frecuencia se calcula por fecha calendario local y no
			// mediante una simple diferencia de horas.
			nextDate := time.Date(
				localCreated.Year(),
				localCreated.Month(),
				localCreated.Day(),
				0,
				0,
				0,
				0,
				businessclock.Location(),
			).AddDate(
				0,
				0,
				frequencyDays,
			)

			now := businessclock.Now()

			today := time.Date(
				now.Year(),
				now.Month(),
				now.Day(),
				0,
				0,
				0,
				0,
				businessclock.Location(),
			)

			if today.Before(nextDate) {
				return models.Reservation{},
					RequestFrequencyError{
						NextDate: nextDate,
					}
			}
		}
	}

	// ------------------------------------------------------------
	// Determinación del flujo grupal.
	// ------------------------------------------------------------
	//
	// No se utilizan IDs de recursos hardcodeados.
	//
	// Qué recursos requieren participantes pertenece a la política
	// versionada mediante reservation_policy_group_resources. De esta
	// forma la configuración puede cambiar sin modificar código Go.

	var groupMinimumParticipants sql.NullInt64

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT minimum_participants
		FROM reservation_policy_group_resources
		WHERE policy_id = $1
		  AND resource_id = $2
		`,
		policy.ID,
		reservation.ResourceID,
	).Scan(&groupMinimumParticipants)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return models.Reservation{}, err
	}

	isGroupReservation := groupMinimumParticipants.Valid

	// ------------------------------------------------------------
	// Reserva no grupal.
	// ------------------------------------------------------------
	//
	// Mantiene el comportamiento tradicional del MVP1 para recursos
	// que no requieren conformar un grupo de participantes.

	if !isGroupReservation {

		var id int

		err = tx.QueryRowContext(
			ctx,
			`
			INSERT INTO reservations (
				policy_id,
				user_id,
				resource_id,
				activity_id,
				start_time,
				duration_minutes,
				status
			)
			VALUES (
				$1,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7
			)
			RETURNING id
			`,
			policy.ID,
			reservation.UserID,
			reservation.ResourceID,
			reservation.ActivityID,
			reservation.StartTime,
			reservation.DurationMinutes,
			reservation.Status,
		).Scan(&id)

		if err != nil {
			return models.Reservation{}, err
		}

		created, err :=
			getReservationByID(
				ctx,
				tx,
				id,
			)

		if err != nil {
			return models.Reservation{}, err
		}

		if err := tx.Commit(); err != nil {
			return models.Reservation{}, err
		}

		return created, nil
	}

	// ------------------------------------------------------------
	// Reserva grupal.
	// ------------------------------------------------------------
	//
	// Una reserva grupal necesita una capacidad válida porque esta
	// determina tanto el máximo de participantes como el snapshot que
	// conservará la reserva aunque la capacidad del recurso cambie después.

	if !capacity.Valid ||
		capacity.Int64 <= 0 {

		return models.Reservation{},
			ErrInvalidGroupConfig
	}

	// El mínimo nunca puede ser cero, negativo ni superar la capacidad
	// disponible del recurso.
	minimumParticipants := int(groupMinimumParticipants.Int64)

	if minimumParticipants <= 0 ||
		minimumParticipants >
			int(capacity.Int64) {

		return models.Reservation{},
			ErrInvalidGroupConfig
	}

	// Se genera un código aleatorio que podrán utilizar otros usuarios
	// para incorporarse al grupo.
	//
	// El código original no se almacena; solo su hash SHA-256.
	joinCode, err := generateJoinCode()
	if err != nil {
		return models.Reservation{}, err
	}

	// Una reserva grupal comienza pendiente mientras debe reunir
	// participantes suficientes para alcanzar minimum_participants.
	initialStatus :=
		models.ReservationStatusPending

	// Se mantiene la lógica genérica por si una futura política define
	// un grupo cuyo mínimo sea un único participante.
	if minimumParticipants <= 1 {
		initialStatus =
			models.ReservationStatusConfirmed
	}

	var reservationID int

	// group_capacity_snapshot conserva la capacidad existente al momento
	// de crear la reserva. Cambios futuros en resources.capacity no alteran
	// el límite histórico de esta reserva.
	err = tx.QueryRowContext(
		ctx,
		`
		INSERT INTO reservations (
			policy_id,
			user_id,
			resource_id,
			activity_id,
			start_time,
			duration_minutes,
			status,
			join_code_hash,
			group_capacity_snapshot,
			group_minimum_participants_snapshot
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10
		)
		RETURNING id
		`,
		policy.ID,
		reservation.UserID,
		reservation.ResourceID,
		reservation.ActivityID,
		reservation.StartTime,
		reservation.DurationMinutes,
		initialStatus,
		codeHash(joinCode),
		int(capacity.Int64),
		minimumParticipants,
	).Scan(&reservationID)

	if err != nil {
		return models.Reservation{}, err
	}

	// ------------------------------------------------------------
	// Owner del grupo.
	// ------------------------------------------------------------
	//
	// El usuario que crea la reserva se incorpora automáticamente como
	// primer participante confirmado y se identifica como owner.
	//
	// El owner no podrá retirarse mediante el flujo normal de participantes;
	// deberá cancelar la reserva completa si desea abandonarla.
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
			'CONFIRMED',
			true,
			CURRENT_TIMESTAMP
		)
		`,
		reservationID,
		reservation.UserID,
	)

	if err != nil {
		return models.Reservation{}, err
	}

	// ------------------------------------------------------------
	// Respuesta de creación.
	// ------------------------------------------------------------

	created, err :=
		getReservationByID(
			ctx,
			tx,
			reservationID,
		)

	if err != nil {
		return models.Reservation{}, err
	}

	capacityValue := int(capacity.Int64)

	// El join code solo se adjunta a esta respuesta.
	// Consultas posteriores no pueden reconstruirlo desde su hash.
	created.JoinCode = joinCode

	created.IsGroupReservation = true
	created.ParticipantCount = 1
	created.MinimumParticipants =
		minimumParticipants
	created.Capacity = &capacityValue

	// La condición grupal es independiente de reservation.status.
	//
	// Ejemplo inicial con mínimo 10:
	//
	//     reservation.status = PENDING
	//     participantes      = 1
	//     groupCondition     = PENDING_MINIMUM
	created.GroupCondition =
		participantGroupCondition(
			initialStatus,
			1,
			minimumParticipants,
		)

	// Ninguna parte de la reserva queda persistida hasta que todas
	// las operaciones anteriores han finalizado correctamente.
	if err := tx.Commit(); err != nil {
		return models.Reservation{}, err
	}

	return created, nil
}

func policyAllowsResource(policy models.ReservationPolicy, resourceID int) bool {
	for _, allowedID := range policy.ResourceIDs {
		if allowedID == resourceID {
			return true
		}
	}
	return false
}

func mapReservationType(status string) string {
	switch status {
	case "PENDING":
		return "pending"
	case "CANCELLED":
		return "cancelled"
	default:
		return "normal"
	}
}

func GetReservationCancellationSnapshot(id int) (int, string, time.Time, int, error) {
	var ownerID, durationMinutes int
	var status string
	var startTime time.Time
	err := database.DB.QueryRowContext(context.Background(), `SELECT user_id, status, start_time, duration_minutes FROM reservations WHERE id = $1`, id).Scan(&ownerID, &status, &startTime, &durationMinutes)
	return ownerID, status, startTime.In(businessclock.Location()), durationMinutes, err
}

func CancelReservationAuthorized(id int, requestedBy models.LocalAuthUser, now time.Time) (models.Reservation, error) {
	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return models.Reservation{}, err
	}
	defer tx.Rollback()

	var ownerID int
	var status string
	var endTime time.Time
	if err := tx.QueryRowContext(ctx, `SELECT user_id, status, end_time FROM reservations WHERE id = $1 FOR UPDATE`, id).Scan(&ownerID, &status, &endTime); err != nil {
		return models.Reservation{}, err
	}
	if !requestedBy.IsAdmin && ownerID != requestedBy.ID {
		return models.Reservation{}, ErrReservationForbidden
	}
	if status != models.ReservationStatusConfirmed && status != models.ReservationStatusPending {
		return models.Reservation{}, ErrReservationNotCancellable
	}
	if !endTime.After(now) {
		return models.Reservation{}, ErrReservationFinalized
	}
	result, err := tx.ExecContext(ctx, `UPDATE reservations SET status = 'CANCELLED' WHERE id = $1 AND status IN ('CONFIRMED', 'PENDING') AND end_time > $2`, id, now)
	if err != nil {
		return models.Reservation{}, err
	}
	if count, _ := result.RowsAffected(); count != 1 {
		return models.Reservation{}, ErrReservationNotCancellable
	}
	cancelled, err := getReservationByID(ctx, tx, id)
	if err != nil {
		return models.Reservation{}, err
	}
	if err := tx.Commit(); err != nil {
		return models.Reservation{}, err
	}
	return cancelled, nil
}

func CancelReservation(id int) (models.Reservation, error) {
	var updatedID int
	err := database.DB.QueryRowContext(context.Background(), `UPDATE reservations SET status = 'CANCELLED' WHERE id = $1 AND status IN ('CONFIRMED', 'PENDING') AND end_time > CURRENT_TIMESTAMP RETURNING id`, id).Scan(&updatedID)
	if err != nil {
		return models.Reservation{}, err
	}
	return GetReservationByID(updatedID)
}

func IsUserAdmin(userID int) (bool, error) {
	var isAdmin bool
	err := database.DB.QueryRowContext(context.Background(), `SELECT is_admin FROM users WHERE id = $1 AND is_blocked = false`, userID).Scan(&isAdmin)
	return isAdmin, err
}
