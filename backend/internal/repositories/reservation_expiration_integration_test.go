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

func TestExpirePendingGroupReservationsIntegration(t *testing.T) {
	if os.Getenv("POLIREDI_INTEGRATION") != "1" {
		t.Skip("integration test disabled")
	}

	database.Close()

	if err := database.Connect(); err != nil {
		t.Fatalf("connect postgres: %v", err)
	}

	defer database.Close()

	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin transaction: %v", err)
	}

	defer tx.Rollback()

	var (
		policyID                    int
		confirmationDeadlineMinutes int
		reservableWindowDays        int
		openingMinute               int
		closingMinute               int
		slotIntervalMinutes         int
		durationMinutes             int
	)

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT
			policy.id,
			policy.confirmation_deadline_minutes,
			policy.reservable_window_days,
			policy.opening_minute,
			policy.closing_minute,
			policy.slot_interval_minutes,
			(
				SELECT MIN(duration.duration_minutes)
				FROM reservation_policy_durations duration
				WHERE duration.policy_id = policy.id
			)
		FROM reservation_policies policy
		WHERE policy.is_published = true
		  AND policy.effective_from <= CURRENT_TIMESTAMP
		  AND (
			policy.effective_to IS NULL
			OR policy.effective_to > CURRENT_TIMESTAMP
		  )
		ORDER BY policy.effective_from DESC, policy.id DESC
		LIMIT 1
		`,
	).Scan(
		&policyID,
		&confirmationDeadlineMinutes,
		&reservableWindowDays,
		&openingMinute,
		&closingMinute,
		&slotIntervalMinutes,
		&durationMinutes,
	)

	if err != nil {
		t.Fatalf("load current policy: %v", err)
	}

	if slotIntervalMinutes <= 0 {
		t.Fatalf(
			"current policy has invalid slot interval: %d",
			slotIntervalMinutes,
		)
	}

	now := businessclock.Now()
	startTime := firstFuturePolicySlot(
		now,
		reservableWindowDays,
		openingMinute,
		closingMinute,
		slotIntervalMinutes,
		durationMinutes,
	)

	if startTime.IsZero() {
		t.Fatal("current policy has no future slot for expiration test")
	}

	stamp := time.Now().UnixNano()

	var userID int
	err = tx.QueryRowContext(
		ctx,
		`
		INSERT INTO users (
			email,
			full_name,
			rut,
			is_admin,
			is_blocked
		)
		VALUES ($1, $2, $3, false, false)
		RETURNING id
		`,
		fmt.Sprintf("expiration.%d@test.local", stamp),
		"Expiration Integration User",
		fmt.Sprintf("%08d-5", stamp%90000000),
	).Scan(&userID)

	if err != nil {
		t.Fatalf("create integration user: %v", err)
	}

	var venueID int
	err = tx.QueryRowContext(
		ctx,
		`SELECT id FROM venues ORDER BY id LIMIT 1`,
	).Scan(&venueID)
	if err != nil {
		t.Fatalf("load venue: %v", err)
	}

	const capacity = 20
	const minimum = 2

	var resourceID int
	err = tx.QueryRowContext(
		ctx,
		`
		INSERT INTO resources (
			venue_id,
			name,
			type,
			reservation_mode,
			capacity,
			is_active
		)
		VALUES ($1, $2, 'Test', 'RESERVABLE', $3, true)
		RETURNING id
		`,
		venueID,
		fmt.Sprintf("Expiration Test Resource %d", stamp),
		capacity,
	).Scan(&resourceID)

	if err != nil {
		t.Fatalf("create integration resource: %v", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`
		INSERT INTO reservation_policy_resources (policy_id, resource_id)
		VALUES ($1, $2)
		`,
		policyID,
		resourceID,
	); err != nil {
		t.Fatalf("scope integration resource: %v", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`
		INSERT INTO reservation_policy_group_resources (
			policy_id,
			resource_id,
			minimum_participants
		)
		VALUES ($1, $2, $3)
		`,
		policyID,
		resourceID,
		minimum,
	); err != nil {
		t.Fatalf("configure integration group resource: %v", err)
	}

	var reservationID int
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
			NULL,
			$4,
			$5,
			'PENDING',
			$6,
			$7,
			$8
		)
		RETURNING id
		`,
		policyID,
		userID,
		resourceID,
		startTime,
		durationMinutes,
		fmt.Sprintf("%064x", stamp),
		capacity,
		minimum,
	).Scan(&reservationID)

	if err != nil {
		t.Fatalf("create pending group reservation: %v", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`
		INSERT INTO participants (
			reservation_id,
			user_id,
			status,
			is_owner,
			confirmed_at
		)
		VALUES ($1, $2, 'CONFIRMED', true, CURRENT_TIMESTAMP)
		`,
		reservationID,
		userID,
	); err != nil {
		t.Fatalf("create owner participant: %v", err)
	}

	deadline := startTime.Add(
		-time.Duration(confirmationDeadlineMinutes) * time.Minute,
	)

	affected, err := expirePendingGroupReservations(
		ctx,
		tx,
		deadline,
	)
	if err != nil {
		t.Fatalf("expire pending group reservation: %v", err)
	}

	if affected != 1 {
		t.Fatalf("expected one expired reservation, got %d", affected)
	}

	var status string
	var reason string
	err = tx.QueryRowContext(
		ctx,
		`
		SELECT status, COALESCE(cancellation_reason, '')
		FROM reservations
		WHERE id = $1
		`,
		reservationID,
	).Scan(&status, &reason)

	if err != nil {
		t.Fatalf("read expired reservation: %v", err)
	}

	if status != models.ReservationStatusCancelled {
		t.Fatalf("expected CANCELLED, got %s", status)
	}

	if reason != models.CancellationReasonMinimumNotMet {
		t.Fatalf(
			"expected %s, got %s",
			models.CancellationReasonMinimumNotMet,
			reason,
		)
	}

	affected, err = expirePendingGroupReservations(ctx, tx, deadline)
	if err != nil {
		t.Fatalf("repeat expiration: %v", err)
	}

	if affected != 0 {
		t.Fatalf("idempotent expiration changed %d rows", affected)
	}

	// NOTIF-001: vencimiento notificado una sola vez en MVP3.
	if notificationsEnabled() {
		var total, matching, unread int

		err := tx.QueryRowContext(ctx, `
			SELECT
				COUNT(*),
				COUNT(*) FILTER (
					WHERE user_id = $2 AND type = 'MINIMUM_NOT_MET'
				),
				COUNT(*) FILTER (
					WHERE user_id = $2
					  AND type = 'MINIMUM_NOT_MET'
					  AND is_read = false
				)
			FROM notifications
			WHERE reservation_id = $1
		`, reservationID, userID).Scan(&total, &matching, &unread)

		if err != nil {
			t.Fatalf("read expiration notifications: %v", err)
		}
		if total != 1 || matching != 1 || unread != 1 {
			t.Fatalf(
				"NOTIF-001: expected one unread MINIMUM_NOT_MET for owner; got total=%d matching=%d unread=%d",
				total, matching, unread,
			)
		}

		t.Log("NOTIF-001: vencimiento, destinatario e idempotencia PASS")
	}

	// PT-020: una reserva previamente confirmada que perdió participantes
	// no debe cancelarse por MINIMUM_NOT_MET al alcanzar el deadline.
	// La transición real CONFIRMED -> AT_RISK se verifica por separado
	// en TestParticipantsIntegrationConfirmedToAtRisk.
	var atRiskID int

	err = tx.QueryRowContext(ctx, `
		INSERT INTO reservations (
			policy_id, user_id, resource_id, activity_id,
			start_time, duration_minutes, status, join_code_hash,
			group_capacity_snapshot, group_minimum_participants_snapshot
		)
		VALUES ($1, $2, $3, NULL, $4, $5, 'CONFIRMED', $6, $7, $8)
		RETURNING id
	`,
		policyID, userID, resourceID, startTime, durationMinutes,
		fmt.Sprintf("%064x", stamp+1), capacity, minimum,
	).Scan(&atRiskID)

	if err != nil {
		t.Fatalf("create confirmed fixture: %v", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO participants (
			reservation_id, user_id, status, is_owner, confirmed_at
		)
		VALUES ($1, $2, 'CONFIRMED', true, CURRENT_TIMESTAMP)
	`, atRiskID, userID)

	if err != nil {
		t.Fatalf("insert confirmed owner: %v", err)
	}

	var participantCount int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM participants
		WHERE reservation_id = $1 AND status = 'CONFIRMED'
	`, atRiskID).Scan(&participantCount)

	if err != nil {
		t.Fatalf("count at-risk participants: %v", err)
	}

	if participantGroupCondition(
		models.ReservationStatusConfirmed,
		participantCount,
		minimum,
	) != GroupConditionAtRisk {
		t.Fatalf("fixture must represent CONFIRMED + AT_RISK")
	}

	for _, instant := range []time.Time{
		deadline,
		deadline.Add(time.Second),
	} {
		changed, err := expirePendingGroupReservations(ctx, tx, instant)
		if err != nil {
			t.Fatalf("housekeeping at %s: %v", instant, err)
		}
		if changed != 0 {
			t.Fatalf("unexpected cancellations: %d", changed)
		}

		var currentStatus, currentReason string
		err = tx.QueryRowContext(ctx, `
			SELECT status, COALESCE(cancellation_reason, '')
			FROM reservations WHERE id = $1
		`, atRiskID).Scan(&currentStatus, &currentReason)

		if err != nil {
			t.Fatalf("read at-risk reservation: %v", err)
		}
		if currentStatus != models.ReservationStatusConfirmed ||
			currentReason != "" {
			t.Fatalf(
				"PT-020: expected CONFIRMED without cancellation, got %s / %s",
				currentStatus, currentReason,
			)
		}
	}

	var cancellationNotifications int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM notifications
		WHERE reservation_id = $1 AND type = 'MINIMUM_NOT_MET'
	`, atRiskID).Scan(&cancellationNotifications)

	if err != nil {
		t.Fatalf("check cancellation notifications: %v", err)
	}
	if cancellationNotifications != 0 {
		t.Fatalf("PT-020: unexpected cancellation notification")
	}

}

func firstFuturePolicySlot(
	now time.Time,
	reservableWindowDays int,
	openingMinute int,
	closingMinute int,
	slotIntervalMinutes int,
	durationMinutes int,
) time.Time {
	for dayOffset := 0; dayOffset < reservableWindowDays; dayOffset++ {
		day := time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			0,
			0,
			0,
			0,
			businessclock.Location(),
		).AddDate(0, 0, dayOffset)

		for minute := openingMinute; minute+durationMinutes <= closingMinute; minute += slotIntervalMinutes {
			candidate := day.Add(
				time.Duration(minute) * time.Minute,
			)

			if candidate.After(now.Add(time.Minute)) {
				return candidate
			}
		}
	}

	return time.Time{}
}
