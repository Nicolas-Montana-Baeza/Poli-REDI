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
