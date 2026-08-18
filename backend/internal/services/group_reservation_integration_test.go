package services

import (
	"context"
	"database/sql"
	"os"
	"regexp"
	"testing"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

// TestCreateGroupReservationIntegration valida el flujo completo de creación
// de una reserva grupal contra PostgreSQL.
//
// El objetivo es demostrar que una cancha configurada como recurso grupal:
//
//   - crea la reserva inicialmente en PENDING;
//   - genera un join code legible;
//   - almacena solamente el hash del código;
//   - conserva un snapshot de la capacidad;
//   - registra al solicitante como primer participante;
//   - identifica al solicitante como owner;
//   - comienza con groupCondition = PENDING_MINIMUM.
//
// Este test utiliza exclusivamente la base PostgreSQL temporal de integración.
func TestCreateGroupReservationIntegration(t *testing.T) {
	if os.Getenv("POLIREDI_INTEGRATION") != "1" {
		t.Skip("integration test disabled")
	}

	// El flujo grupal que estamos desarrollando no necesita todavía los
	// módulos legacy/full. Esto evita que el test dependa de tablas como
	// workshops mientras reconstruimos MVP2 por bloques.
	t.Setenv("MVP_SCOPE", "mvp1")

	database.Close()

	if err := database.Connect(); err != nil {
		t.Fatalf("connect postgres: %v", err)
	}

	defer database.Close()

	ctx := context.Background()

	// ------------------------------------------------------------
	// Política vigente.
	// ------------------------------------------------------------
	//
	// Obtenemos horario y duración directamente desde la política para
	// evitar que el test dependa de valores hardcodeados.

	var (
		policyID      int
		openingMinute int
		closingMinute int
	)

	err := database.DB.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			opening_minute,
			closing_minute
		FROM reservation_policies
		WHERE is_published = true
		  AND effective_from <= CURRENT_TIMESTAMP
		  AND (
		      effective_to IS NULL
		      OR effective_to > CURRENT_TIMESTAMP
		  )
		ORDER BY effective_from DESC, id DESC
		LIMIT 1
		`,
	).Scan(
		&policyID,
		&openingMinute,
		&closingMinute,
	)

	if err != nil {
		t.Fatalf("load current policy: %v", err)
	}

	var durationMinutes int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT duration_minutes
		FROM reservation_policy_durations
		WHERE policy_id = $1
		ORDER BY duration_minutes ASC
		LIMIT 1
		`,
		policyID,
	).Scan(&durationMinutes)

	if err != nil {
		t.Fatalf("load allowed duration: %v", err)
	}

	if openingMinute+durationMinutes > closingMinute {
		t.Fatalf(
			"invalid test policy: opening=%d duration=%d closing=%d",
			openingMinute,
			durationMinutes,
			closingMinute,
		)
	}

	// ------------------------------------------------------------
	// Recurso grupal.
	// ------------------------------------------------------------

	var (
		resourceID int
		capacity   int
	)

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT id, capacity
		FROM resources
		WHERE name = 'Cancha 1, Centro Deportivo'
		  AND reservation_mode = 'RESERVABLE'
		  AND is_active = true
		`,
	).Scan(
		&resourceID,
		&capacity,
	)

	if err != nil {
		t.Fatalf("load group resource: %v", err)
	}

	// Confirmamos que el recurso pertenece explícitamente al flujo grupal
	// de la política vigente y no depende de un ID hardcodeado en Go.
	var isGroupResource bool

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM reservation_policy_group_resources
			WHERE policy_id = $1
			  AND resource_id = $2
		)
		`,
		policyID,
		resourceID,
	).Scan(&isGroupResource)

	if err != nil {
		t.Fatalf("check group resource: %v", err)
	}

	if !isGroupResource {
		t.Fatal("Cancha 1 is not configured as group resource")
	}

	// ------------------------------------------------------------
	// Actividad.
	// ------------------------------------------------------------

	var activityID int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT id
		FROM activities
		WHERE name = 'Fútbol'
		  AND is_active = true
		LIMIT 1
		`,
	).Scan(&activityID)

	if err != nil {
		t.Fatalf("load activity: %v", err)
	}

	// ------------------------------------------------------------
	// Usuario temporal.
	// ------------------------------------------------------------
	//
	// Se crea un usuario completamente nuevo para que la regla de
	// frecuencia semanal no interfiera con el escenario probado.

	var userID int

	err = database.DB.QueryRowContext(
		ctx,
		`
		INSERT INTO users (
			email,
			full_name,
			rut,
			is_admin,
			is_blocked
		)
		VALUES (
			'mvp2.create.integration@test.local',
			'Integration Group Owner',
			'22333444-5',
			false,
			false
		)
		RETURNING id
		`,
	).Scan(&userID)

	if err != nil {
		t.Fatalf("create integration user: %v", err)
	}

	var reservationID int

	// El cleanup elimina cualquier dato persistido por el test aunque una
	// aserción posterior falle.
	defer func() {
		if reservationID > 0 {
			_, _ = database.DB.ExecContext(
				ctx,
				`
				DELETE FROM reservation_participant_audit
				WHERE reservation_id = $1
				`,
				reservationID,
			)

			_, _ = database.DB.ExecContext(
				ctx,
				`
				DELETE FROM participants
				WHERE reservation_id = $1
				`,
				reservationID,
			)

			_, _ = database.DB.ExecContext(
				ctx,
				`
				DELETE FROM reservations
				WHERE id = $1
				`,
				reservationID,
			)
		}

		_, _ = database.DB.ExecContext(
			ctx,
			`
			DELETE FROM users
			WHERE id = $1
			`,
			userID,
		)
	}()

	// ------------------------------------------------------------
	// Horario válido.
	// ------------------------------------------------------------
	//
	// Se utiliza mañana y el minuto de apertura configurado por la política.

	now := businessclock.Now()

	startTime := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0,
		0,
		0,
		0,
		businessclock.Location(),
	).AddDate(0, 0, 1).Add(
		time.Duration(openingMinute) * time.Minute,
	)

	// ------------------------------------------------------------
	// Creación mediante la capa de servicio.
	// ------------------------------------------------------------

	created, err := CreateReservation(
		models.Reservation{
			UserID:          userID,
			ResourceID:      resourceID,
			ActivityID:      &activityID,
			StartTime:       startTime,
			DurationMinutes: durationMinutes,
		},
	)

	if err != nil {
		t.Fatalf("create group reservation: %v", err)
	}

	reservationID = created.ID

	// ------------------------------------------------------------
	// Estado inicial del grupo.
	// ------------------------------------------------------------

	if created.Status != models.ReservationStatusPending {
		t.Fatalf(
			"expected PENDING, got %s",
			created.Status,
		)
	}

	if !created.IsGroupReservation {
		t.Fatal("expected group reservation")
	}

	if created.ParticipantCount != 1 {
		t.Fatalf(
			"expected 1 participant, got %d",
			created.ParticipantCount,
		)
	}

	if created.MinimumParticipants != 10 {
		t.Fatalf(
			"expected minimum 10, got %d",
			created.MinimumParticipants,
		)
	}

	if created.Capacity == nil {
		t.Fatal("expected capacity snapshot")
	}

	if *created.Capacity != capacity {
		t.Fatalf(
			"expected capacity %d, got %d",
			capacity,
			*created.Capacity,
		)
	}

	if created.GroupCondition != models.GroupConditionPending {
		t.Fatalf(
			"expected PENDING_MINIMUM, got %s",
			created.GroupCondition,
		)
	}

	// ------------------------------------------------------------
	// Join code.
	// ------------------------------------------------------------
	//
	// El código debe usar el formato legible ABCDE-FGHIJ.

	joinCodePattern :=
		regexp.MustCompile(`^[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{5}-[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{5}$`)

	if !joinCodePattern.MatchString(created.JoinCode) {
		t.Fatalf(
			"unexpected join code format: %q",
			created.JoinCode,
		)
	}

	// ------------------------------------------------------------
	// Persistencia segura del join code.
	// ------------------------------------------------------------

	var (
		storedHash       string
		capacitySnapshot int
	)

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT
			join_code_hash,
			group_capacity_snapshot
		FROM reservations
		WHERE id = $1
		`,
		reservationID,
	).Scan(
		&storedHash,
		&capacitySnapshot,
	)

	if err != nil {
		t.Fatalf("read stored group reservation: %v", err)
	}

	// PostgreSQL nunca debe contener el código original.
	if storedHash == created.JoinCode {
		t.Fatal("plaintext join code was stored in database")
	}

	if len(storedHash) != 64 {
		t.Fatalf(
			"expected SHA-256 hash with 64 hex chars, got %d",
			len(storedHash),
		)
	}

	if capacitySnapshot != capacity {
		t.Fatalf(
			"expected stored capacity %d, got %d",
			capacity,
			capacitySnapshot,
		)
	}

	// ------------------------------------------------------------
	// Owner.
	// ------------------------------------------------------------
	//
	// El creador debe existir inmediatamente como primer participante
	// CONFIRMED y ser el único owner del grupo.

	var (
		participantStatus string
		isOwner           bool
		confirmedAt       sql.NullTime
	)

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT
			status,
			is_owner,
			confirmed_at
		FROM participants
		WHERE reservation_id = $1
		  AND user_id = $2
		`,
		reservationID,
		userID,
	).Scan(
		&participantStatus,
		&isOwner,
		&confirmedAt,
	)

	if err != nil {
		t.Fatalf("load owner participant: %v", err)
	}

	if participantStatus != "CONFIRMED" {
		t.Fatalf(
			"expected owner CONFIRMED, got %s",
			participantStatus,
		)
	}

	if !isOwner {
		t.Fatal("reservation creator must be owner")
	}

	if !confirmedAt.Valid {
		t.Fatal("owner must have confirmed_at")
	}

	t.Logf(
		"PASS: group reservation %d created as 1/10 PENDING with join code %s",
		reservationID,
		created.JoinCode,
	)
}
