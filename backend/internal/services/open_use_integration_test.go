package services

import (
	"context"
	"os"
	"testing"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

// TestOpenUseDoesNotConsumeReservationFrequencyIntegration valida que
// OPEN_USE mantenga una semántica distinta de RESERVABLE.
//
// Contrato MVP2:
//
//   - una reserva OPEN_USE se crea como reserva normal;
//   - no genera join code;
//   - no genera participantes;
//   - no utiliza flujo grupal;
//   - no consume la frecuencia configurada para recursos RESERVABLE;
//
// El escenario crea primero un uso OPEN_USE y luego, con el mismo usuario,
// intenta crear inmediatamente una reserva RESERVABLE. La segunda operación
// debe ser válida porque el primer uso no consume la frecuencia semanal.
func TestOpenUseDoesNotConsumeReservationFrequencyIntegration(t *testing.T) {
	if os.Getenv("POLIREDI_INTEGRATION") != "1" {
		t.Skip("integration test disabled")
	}

	// MVP2 incluye OPEN_USE y reservas grupales, pero aún no necesita
	// dependencias legacy del scope FULL.
	t.Setenv("MVP_SCOPE", "mvp2")

	database.Close()

	if err := database.Connect(); err != nil {
		t.Fatalf("connect postgres: %v", err)
	}

	defer database.Close()

	ctx := context.Background()

	// ------------------------------------------------------------
	// Política vigente.
	// ------------------------------------------------------------

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
	// Recursos.
	// ------------------------------------------------------------

	var openUseResourceID int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT id
		FROM resources
		WHERE reservation_mode = 'OPEN_USE'
		  AND is_active = true
		  AND id IN (
		      SELECT resource_id
		      FROM reservation_policy_resources
		      WHERE policy_id = $1
		  )
		ORDER BY id
		LIMIT 1
		`,
		policyID,
	).Scan(&openUseResourceID)

	if err != nil {
		t.Fatalf("load OPEN_USE resource: %v", err)
	}

	var reservableResourceID int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT r.id
		FROM resources r
		WHERE r.reservation_mode = 'RESERVABLE'
		  AND r.is_active = true
		  AND r.id IN (
		      SELECT resource_id
		      FROM reservation_policy_resources
		      WHERE policy_id = $1
		  )
		ORDER BY r.id
		LIMIT 1
		`,
		policyID,
	).Scan(&reservableResourceID)

	if err != nil {
		t.Fatalf("load RESERVABLE resource: %v", err)
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
		WHERE is_active = true
		ORDER BY id
		LIMIT 1
		`,
	).Scan(&activityID)

	if err != nil {
		t.Fatalf("load activity: %v", err)
	}

	// ------------------------------------------------------------
	// Usuario temporal.
	// ------------------------------------------------------------

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
			'mvp2.openuse.integration@test.local',
			'Integration Open Use User',
			'23333444-3',
			false,
			false
		)
		RETURNING id
		`,
	).Scan(&userID)

	if err != nil {
		t.Fatalf("create integration user: %v", err)
	}

	reservationIDs := []int{}

	defer func() {
		for _, reservationID := range reservationIDs {
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
	// Horarios válidos.
	// ------------------------------------------------------------

	now := businessclock.Now()

	baseDate := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0,
		0,
		0,
		0,
		businessclock.Location(),
	).AddDate(0, 0, 1)

	openUseStart := baseDate.Add(
		time.Duration(openingMinute) * time.Minute,
	)

	reservableStart := openUseStart.Add(
		time.Duration(durationMinutes) * time.Minute,
	)

	if reservableStart.Add(
		time.Duration(durationMinutes) * time.Minute,
	).After(
		baseDate.Add(
			time.Duration(closingMinute) * time.Minute,
		),
	) {
		t.Fatal("policy does not provide enough room for integration scenario")
	}

	// ------------------------------------------------------------
	// Primer uso: OPEN_USE.
	// ------------------------------------------------------------

	openUseCreated, err := CreateReservation(
		models.Reservation{
			UserID:          userID,
			ResourceID:      openUseResourceID,
			ActivityID:      &activityID,
			StartTime:       openUseStart,
			DurationMinutes: durationMinutes,
		},
	)

	if err != nil {
		t.Fatalf("create OPEN_USE reservation: %v", err)
	}

	reservationIDs = append(
		reservationIDs,
		openUseCreated.ID,
	)

	if openUseCreated.IsGroupReservation {
		t.Fatal("OPEN_USE must not create a group reservation")
	}

	if openUseCreated.JoinCode != "" {
		t.Fatal("OPEN_USE must not generate a join code")
	}

	if openUseCreated.ParticipantCount != 0 {
		t.Fatalf(
			"OPEN_USE expected 0 participants, got %d",
			openUseCreated.ParticipantCount,
		)
	}

	var participantCount int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT COUNT(*)
		FROM participants
		WHERE reservation_id = $1
		`,
		openUseCreated.ID,
	).Scan(&participantCount)

	if err != nil {
		t.Fatalf("count OPEN_USE participants: %v", err)
	}

	if participantCount != 0 {
		t.Fatalf(
			"OPEN_USE persisted %d participants",
			participantCount,
		)
	}

	// ------------------------------------------------------------
	// Segundo uso: RESERVABLE.
	// ------------------------------------------------------------
	//
	// Esta operación ocurre inmediatamente después de OPEN_USE.
	// Si OPEN_USE consumiera incorrectamente la frecuencia semanal,
	// esta creación fallaría con RequestFrequencyError.

	reservableCreated, err := CreateReservation(
		models.Reservation{
			UserID:          userID,
			ResourceID:      reservableResourceID,
			ActivityID:      &activityID,
			StartTime:       reservableStart,
			DurationMinutes: durationMinutes,
		},
	)

	if err != nil {
		t.Fatalf(
			"RESERVABLE after OPEN_USE should be allowed: %v",
			err,
		)
	}

	reservationIDs = append(
		reservationIDs,
		reservableCreated.ID,
	)

	t.Logf(
		"PASS: OPEN_USE reservation %d did not consume frequency; RESERVABLE %d was created",
		openUseCreated.ID,
		reservableCreated.ID,
	)
}

// TestOpenUseOverlapRulesIntegration valida la semántica de solapamiento
// específica de los recursos OPEN_USE.
//
// Contrato MVP2:
//
//   - usuarios distintos pueden utilizar simultáneamente el mismo recurso;
//   - el mismo usuario no puede registrar dos usos simultáneos;
//   - OPEN_USE no se comporta como un recurso exclusivo RESERVABLE.
//
// Este test protege el contrato también en backend, independientemente de
// las validaciones preventivas que realiza el frontend.
func TestOpenUseOverlapRulesIntegration(t *testing.T) {
	if os.Getenv("POLIREDI_INTEGRATION") != "1" {
		t.Skip("integration test disabled")
	}

	t.Setenv("MVP_SCOPE", "mvp2")

	database.Close()

	if err := database.Connect(); err != nil {
		t.Fatalf("connect postgres: %v", err)
	}

	defer database.Close()

	ctx := context.Background()

	// ------------------------------------------------------------
	// Política y duración vigente.
	// ------------------------------------------------------------

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
		t.Fatal("policy does not provide a valid test interval")
	}

	// ------------------------------------------------------------
	// Recurso OPEN_USE.
	// ------------------------------------------------------------

	var resourceID int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT r.id
		FROM resources r
		INNER JOIN reservation_policy_resources pr
			ON pr.resource_id = r.id
		   AND pr.policy_id = $1
		WHERE r.reservation_mode = 'OPEN_USE'
		  AND r.is_active = true
		ORDER BY r.id
		LIMIT 1
		`,
		policyID,
	).Scan(&resourceID)

	if err != nil {
		t.Fatalf("load OPEN_USE resource: %v", err)
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
		WHERE is_active = true
		ORDER BY id
		LIMIT 1
		`,
	).Scan(&activityID)

	if err != nil {
		t.Fatalf("load activity: %v", err)
	}

	// ------------------------------------------------------------
	// Usuarios temporales.
	// ------------------------------------------------------------

	var userAID int
	var userBID int

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
			'mvp2.openuse.a@test.local',
			'Integration Open Use A',
			'24333444-1',
			false,
			false
		)
		RETURNING id
		`,
	).Scan(&userAID)

	if err != nil {
		t.Fatalf("create user A: %v", err)
	}

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
			'mvp2.openuse.b@test.local',
			'Integration Open Use B',
			'25333444-K',
			false,
			false
		)
		RETURNING id
		`,
	).Scan(&userBID)

	if err != nil {
		t.Fatalf("create user B: %v", err)
	}

	reservationIDs := []int{}

	defer func() {
		for _, reservationID := range reservationIDs {
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
			WHERE id IN ($1, $2)
			`,
			userAID,
			userBID,
		)
	}()

	// ------------------------------------------------------------
	// Intervalo compartido.
	// ------------------------------------------------------------

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
	// Usuario A registra el primer uso.
	// ------------------------------------------------------------

	first, err := CreateReservation(
		models.Reservation{
			UserID:          userAID,
			ResourceID:      resourceID,
			ActivityID:      &activityID,
			StartTime:       startTime,
			DurationMinutes: durationMinutes,
		},
	)

	if err != nil {
		t.Fatalf("create first OPEN_USE reservation: %v", err)
	}

	reservationIDs = append(
		reservationIDs,
		first.ID,
	)

	// ------------------------------------------------------------
	// Usuario B usa el mismo recurso al mismo tiempo.
	// ------------------------------------------------------------
	//
	// Esto debe estar permitido porque OPEN_USE representa uso concurrente
	// y no una ocupación exclusiva del recurso.

	second, err := CreateReservation(
		models.Reservation{
			UserID:          userBID,
			ResourceID:      resourceID,
			ActivityID:      &activityID,
			StartTime:       startTime,
			DurationMinutes: durationMinutes,
		},
	)

	if err != nil {
		t.Fatalf(
			"different user should share OPEN_USE interval: %v",
			err,
		)
	}

	reservationIDs = append(
		reservationIDs,
		second.ID,
	)

	// ------------------------------------------------------------
	// Usuario A intenta duplicar su propio uso simultáneo.
	// ------------------------------------------------------------
	//
	// Este caso debe ser rechazado. Un usuario no puede figurar utilizando
	// dos veces el mismo intervalo aunque OPEN_USE permita concurrencia
	// entre personas distintas.

	duplicate, err := CreateReservation(
		models.Reservation{
			UserID:          userAID,
			ResourceID:      resourceID,
			ActivityID:      &activityID,
			StartTime:       startTime,
			DurationMinutes: durationMinutes,
		},
	)

	if err == nil {
		reservationIDs = append(
			reservationIDs,
			duplicate.ID,
		)

		t.Fatal(
			"same user was allowed to create overlapping OPEN_USE reservations",
		)
	}

	t.Logf(
		"PASS: users %d and %d shared OPEN_USE resource %d; duplicate use by user %d was rejected",
		userAID,
		userBID,
		resourceID,
		userAID,
	)
}

// TestOpenUseRespectsAdministrativeBlocksIntegration valida que OPEN_USE
// permita concurrencia entre usuarios, pero nunca ignore un bloqueo
// administrativo.
//
// Un availability_block representa una indisponibilidad institucional del
// recurso y tiene prioridad sobre la semántica concurrente de OPEN_USE.
func TestOpenUseRespectsAdministrativeBlocksIntegration(t *testing.T) {
	if os.Getenv("POLIREDI_INTEGRATION") != "1" {
		t.Skip("integration test disabled")
	}

	t.Setenv("MVP_SCOPE", "mvp2")

	database.Close()

	if err := database.Connect(); err != nil {
		t.Fatalf("connect postgres: %v", err)
	}

	defer database.Close()

	ctx := context.Background()

	// ------------------------------------------------------------
	// Política vigente.
	// ------------------------------------------------------------

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
		t.Fatal("policy does not provide a valid test interval")
	}

	// ------------------------------------------------------------
	// Recurso OPEN_USE.
	// ------------------------------------------------------------

	var resourceID int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT r.id
		FROM resources r
		INNER JOIN reservation_policy_resources pr
			ON pr.resource_id = r.id
		   AND pr.policy_id = $1
		WHERE r.reservation_mode = 'OPEN_USE'
		  AND r.is_active = true
		ORDER BY r.id
		LIMIT 1
		`,
		policyID,
	).Scan(&resourceID)

	if err != nil {
		t.Fatalf("load OPEN_USE resource: %v", err)
	}

	// ------------------------------------------------------------
	// Actividad y usuario temporal.
	// ------------------------------------------------------------

	var activityID int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT id
		FROM activities
		WHERE is_active = true
		ORDER BY id
		LIMIT 1
		`,
	).Scan(&activityID)

	if err != nil {
		t.Fatalf("load activity: %v", err)
	}

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
			'mvp2.openuse.block@test.local',
			'Integration Open Use Block',
			'26333444-8',
			false,
			false
		)
		RETURNING id
		`,
	).Scan(&userID)

	if err != nil {
		t.Fatalf("create integration user: %v", err)
	}

	// ------------------------------------------------------------
	// Intervalo.
	// ------------------------------------------------------------

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

	endTime := startTime.Add(
		time.Duration(durationMinutes) * time.Minute,
	)

	var blockID int

	// ------------------------------------------------------------
	// Bloqueo institucional.
	// ------------------------------------------------------------

	err = database.DB.QueryRowContext(
		ctx,
		`
		INSERT INTO availability_blocks (
			resource_id,
			block_type,
			reason,
			start_time,
			end_time,
			created_by_user_id,
			is_active
		)
		VALUES (
			$1,
			'MAINTENANCE',
			'Prueba integración OPEN_USE',
			$2,
			$3,
			$4,
			true
		)
		RETURNING id
		`,
		resourceID,
		startTime,
		endTime,
		userID,
	).Scan(&blockID)

	if err != nil {
		t.Fatalf("create availability block: %v", err)
	}

	defer func() {
		_, _ = database.DB.ExecContext(
			ctx,
			`
			DELETE FROM reservations
			WHERE user_id = $1
			`,
			userID,
		)

		_, _ = database.DB.ExecContext(
			ctx,
			`
			DELETE FROM availability_blocks
			WHERE id = $1
			`,
			blockID,
		)

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
	// Intento de uso.
	// ------------------------------------------------------------
	//
	// OPEN_USE permite concurrencia entre usuarios, pero un bloqueo
	// administrativo representa indisponibilidad total del recurso.

	created, err := CreateReservation(
		models.Reservation{
			UserID:          userID,
			ResourceID:      resourceID,
			ActivityID:      &activityID,
			StartTime:       startTime,
			DurationMinutes: durationMinutes,
		},
	)

	if err == nil {
		t.Fatalf(
			"OPEN_USE ignored administrative block and created reservation %d",
			created.ID,
		)
	}

	t.Logf(
		"PASS: OPEN_USE resource %d rejected usage during administrative block %d",
		resourceID,
		blockID,
	)
}
