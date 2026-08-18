package repositories

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func TestParticipantsIntegrationConfirmedToAtRisk(t *testing.T) {
	if os.Getenv("POLIREDI_INTEGRATION") != "1" {
		t.Skip("integration test disabled")
	}

	database.Close()

	if err := database.Connect(); err != nil {
		t.Fatalf("connect postgres: %v", err)
	}

	defer database.Close()

	ctx := context.Background()

	// ---------------------------------------------------------------------
	// Política, recurso y actividad.
	// ---------------------------------------------------------------------

	var (
		policyID   int
		resourceID int
		capacity   int
		activityID int
	)

	err := database.DB.QueryRowContext(
		ctx,
		`
		SELECT id
		FROM reservation_policies
		WHERE idempotency_key = 'mvp2-group-participants-v1'
		`,
	).Scan(&policyID)

	if err != nil {
		t.Fatalf("load policy: %v", err)
	}

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT id, capacity
		FROM resources
		WHERE name = 'Cancha 1, Centro Deportivo'
		  AND reservation_mode = 'RESERVABLE'
		`,
	).Scan(&resourceID, &capacity)

	if err != nil {
		t.Fatalf("load resource: %v", err)
	}

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT id
		FROM activities
		WHERE name = 'Fútbol'
		`,
	).Scan(&activityID)

	if err != nil {
		t.Fatalf("load activity: %v", err)
	}

	// ---------------------------------------------------------------------
	// Usuarios temporales:
	//
	// 1 owner + 9 participantes.
	// ---------------------------------------------------------------------

	stamp := time.Now().UnixNano()

	userIDs := make([]int, 10)

	baseRUT := int(stamp%7000000) + 20000000

	for i := 0; i < 10; i++ {
		email := fmt.Sprintf(
			"mvp2.integration.%d.%d@test.local",
			stamp,
			i,
		)

		rut := fmt.Sprintf(
			"%08d-%d",
			baseRUT+i,
			i%10,
		)

		err := database.DB.QueryRowContext(
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
				$1,
				$2,
				$3,
				false,
				false
			)
			RETURNING id
			`,
			email,
			fmt.Sprintf("Integration User %d", i),
			rut,
		).Scan(&userIDs[i])

		if err != nil {
			t.Fatalf(
				"create user %d: %v",
				i,
				err,
			)
		}
	}

	// ---------------------------------------------------------------------
	// Código de unión.
	// ---------------------------------------------------------------------

	joinCode := fmt.Sprintf(
		"TEST-%d",
		stamp,
	)

	// ---------------------------------------------------------------------
	// Reserva para mañana a las 20:00.
	// ---------------------------------------------------------------------

	now := businessclock.Now()

	startTime := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		20,
		0,
		0,
		0,
		businessclock.Location(),
	).AddDate(0, 0, 1)

	var reservationID int

	err = database.DB.QueryRowContext(
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
			group_capacity_snapshot
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			60,
			'PENDING',
			$6,
			$7
		)
		RETURNING id
		`,
		policyID,
		userIDs[0],
		resourceID,
		activityID,
		startTime,
		codeHash(joinCode),
		capacity,
	).Scan(&reservationID)

	if err != nil {
		t.Fatalf("create reservation: %v", err)
	}

	// ---------------------------------------------------------------------
	// Cleanup.
	// ---------------------------------------------------------------------

	defer func() {
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

		for _, id := range userIDs {
			_, _ = database.DB.ExecContext(
				ctx,
				`
				DELETE FROM users
				WHERE id = $1
				`,
				id,
			)
		}
	}()

	// ---------------------------------------------------------------------
	// Owner confirmado.
	// ---------------------------------------------------------------------

	_, err = database.DB.ExecContext(
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
		userIDs[0],
	)

	if err != nil {
		t.Fatalf("insert owner: %v", err)
	}

	// ---------------------------------------------------------------------
	// Agregamos 8 participantes directamente.
	//
	// Owner + 8 = 9/10.
	// ---------------------------------------------------------------------

	for i := 1; i <= 8; i++ {
		_, err = database.DB.ExecContext(
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
				false,
				CURRENT_TIMESTAMP
			)
			`,
			reservationID,
			userIDs[i],
		)

		if err != nil {
			t.Fatalf(
				"insert participant %d: %v",
				i,
				err,
			)
		}
	}

	// ---------------------------------------------------------------------
	// Estado inicial: 9/10 PENDING.
	// ---------------------------------------------------------------------

	initial, err := GetReservationProgress(
		joinCode,
		userIDs[9],
	)

	if err != nil {
		t.Fatalf(
			"initial progress: %v",
			err,
		)
	}

	if initial.ParticipantCount != 9 {
		t.Fatalf(
			"expected initial 9 participants, got %d",
			initial.ParticipantCount,
		)
	}

	if initial.Status != models.ReservationStatusPending {
		t.Fatalf(
			"expected initial PENDING, got %s",
			initial.Status,
		)
	}

	if initial.GroupCondition != GroupConditionPending {
		t.Fatalf(
			"expected PENDING_MINIMUM, got %s",
			initial.GroupCondition,
		)
	}

	// ---------------------------------------------------------------------
	// Participante 10 confirma.
	//
	// 9/10 -> 10/10
	// PENDING -> CONFIRMED
	// ---------------------------------------------------------------------

	confirmed, err := ChangeParticipation(
		joinCode,
		userIDs[9],
		true,
	)

	if err != nil {
		t.Fatalf(
			"confirm participant: %v",
			err,
		)
	}

	if confirmed.ParticipantCount != 10 {
		t.Fatalf(
			"expected 10 participants, got %d",
			confirmed.ParticipantCount,
		)
	}

	if confirmed.Status != models.ReservationStatusConfirmed {
		t.Fatalf(
			"expected CONFIRMED, got %s",
			confirmed.Status,
		)
	}

	if confirmed.GroupCondition != GroupConditionHealthy {
		t.Fatalf(
			"expected HEALTHY, got %s",
			confirmed.GroupCondition,
		)
	}

	if !confirmed.IsMember {
		t.Fatal(
			"expected participant to be member",
		)
	}

	// ---------------------------------------------------------------------
	// El mismo participante se retira.
	//
	// Regla B:
	//
	// 10/10 -> 9/10
	// reservation sigue CONFIRMED
	// groupCondition -> AT_RISK
	// ---------------------------------------------------------------------

	withdrawn, err := ChangeParticipation(
		joinCode,
		userIDs[9],
		false,
	)

	if err != nil {
		t.Fatalf(
			"withdraw participant: %v",
			err,
		)
	}

	if withdrawn.ParticipantCount != 9 {
		t.Fatalf(
			"expected 9 participants after withdrawal, got %d",
			withdrawn.ParticipantCount,
		)
	}

	if withdrawn.Status != models.ReservationStatusConfirmed {
		t.Fatalf(
			"reservation regressed from CONFIRMED to %s",
			withdrawn.Status,
		)
	}

	if withdrawn.GroupCondition != GroupConditionAtRisk {
		t.Fatalf(
			"expected AT_RISK, got %s",
			withdrawn.GroupCondition,
		)
	}

	if withdrawn.IsMember {
		t.Fatal(
			"withdrawn participant must not remain member",
		)
	}

	// ---------------------------------------------------------------------
	// Verificación directa del estado persistido.
	// ---------------------------------------------------------------------

	var storedStatus string

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT status
		FROM reservations
		WHERE id = $1
		`,
		reservationID,
	).Scan(&storedStatus)

	if err != nil {
		t.Fatalf(
			"read stored status: %v",
			err,
		)
	}

	if storedStatus != models.ReservationStatusConfirmed {
		t.Fatalf(
			"database reservation status must remain CONFIRMED, got %s",
			storedStatus,
		)
	}

	t.Logf(
		"PASS: 9/10 PENDING -> 10/10 CONFIRMED -> 9/10 CONFIRMED + AT_RISK",
	)
}
