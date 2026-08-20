package repositories

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func TestPostgresFullRepositoriesIntegration(
	t *testing.T,
) {
	if os.Getenv("POLIREDI_INTEGRATION") != "1" {
		t.Skip("integration test disabled")
	}

	database.Close()

	if err := database.Connect(); err != nil {
		t.Fatalf("connect postgres: %v", err)
	}

	defer database.Close()

	ctx := context.Background()

	var adminID int

	if err := database.DB.QueryRowContext(
		ctx,
		`
		SELECT id
		FROM users
		WHERE lower(email) = 'admin@poliredi.local'
		  AND is_admin = true
		  AND is_blocked = false
		LIMIT 1
		`,
	).Scan(&adminID); err != nil {
		t.Fatalf("load admin: %v", err)
	}

	current, err := GetCurrentReservationPolicyComplete()
	if err != nil {
		t.Fatalf("load current policy: %v", err)
	}

	var (
		lateWithdrawalMinutes        int
		groupRecoveryDeadlineMinutes int
	)

	if err := database.DB.QueryRowContext(
		ctx,
		`
		SELECT
			late_withdrawal_minutes,
			group_recovery_deadline_minutes
		FROM reservation_policies
		WHERE id = $1
		`,
		current.ID,
	).Scan(
		&lateWithdrawalMinutes,
		&groupRecoveryDeadlineMinutes,
	); err != nil {
		t.Fatalf(
			"load group policy parameters: %v",
			err,
		)
	}

	oldGroupResources, err :=
		loadGroupResourceIDsForIntegration(
			ctx,
			current.ID,
		)
	if err != nil {
		t.Fatalf(
			"load old group resources: %v",
			err,
		)
	}

	request := models.PublishReservationPolicyRequest{
		ReservableWindowDays: current.ReservableWindowDays,

		RequestFrequencyDays: current.RequestFrequencyDays,

		ConfirmationDeadlineMinutes: current.ConfirmationDeadlineMinutes,

		MinimumParticipants: current.MinimumParticipants,

		OpeningMinute: current.OpeningMinute,

		ClosingMinute: current.ClosingMinute,

		SlotIntervalMinutes: current.SlotIntervalMinutes,

		AllowedDurations: append(
			[]int(nil),
			current.AllowedDurations...,
		),

		ResourceIDs: append(
			[]int(nil),
			current.ResourceIDs...,
		),
	}

	key := fmt.Sprintf(
		"integration-policy-%d",
		time.Now().UnixNano(),
	)

	sum := sha256.Sum256([]byte(key))
	payloadHash := hex.EncodeToString(sum[:])

	published, replayed, err :=
		PublishReservationPolicy(
			request,
			adminID,
			key,
			payloadHash,
		)
	if err != nil {
		t.Fatalf(
			"publish PostgreSQL policy: %v",
			err,
		)
	}

	if replayed {
		t.Fatal(
			"first publication must not be replayed",
		)
	}

	var (
		newLateWithdrawalMinutes        int
		newGroupRecoveryDeadlineMinutes int
	)

	if err := database.DB.QueryRowContext(
		ctx,
		`
		SELECT
			late_withdrawal_minutes,
			group_recovery_deadline_minutes
		FROM reservation_policies
		WHERE id = $1
		`,
		published.ID,
	).Scan(
		&newLateWithdrawalMinutes,
		&newGroupRecoveryDeadlineMinutes,
	); err != nil {
		t.Fatalf(
			"load published group parameters: %v",
			err,
		)
	}

	if newLateWithdrawalMinutes !=
		lateWithdrawalMinutes {
		t.Fatalf(
			"late withdrawal changed: got %d want %d",
			newLateWithdrawalMinutes,
			lateWithdrawalMinutes,
		)
	}

	if newGroupRecoveryDeadlineMinutes !=
		groupRecoveryDeadlineMinutes {
		t.Fatalf(
			"group recovery deadline changed: got %d want %d",
			newGroupRecoveryDeadlineMinutes,
			groupRecoveryDeadlineMinutes,
		)
	}

	newGroupResources, err :=
		loadGroupResourceIDsForIntegration(
			ctx,
			published.ID,
		)
	if err != nil {
		t.Fatalf(
			"load new group resources: %v",
			err,
		)
	}

	if !reflect.DeepEqual(
		oldGroupResources,
		newGroupResources,
	) {
		t.Fatalf(
			"group resources changed: got %v want %v",
			newGroupResources,
			oldGroupResources,
		)
	}

	replayedPolicy, replayed, err :=
		PublishReservationPolicy(
			request,
			adminID,
			key,
			payloadHash,
		)
	if err != nil {
		t.Fatalf(
			"replay publication: %v",
			err,
		)
	}

	if !replayed {
		t.Fatal(
			"second publication must be replayed",
		)
	}

	if replayedPolicy.ID != published.ID {
		t.Fatalf(
			"replay changed policy id: got %d want %d",
			replayedPolicy.ID,
			published.ID,
		)
	}

	_, _, err = PublishReservationPolicy(
		request,
		adminID,
		key,
		"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	)

	if !errors.Is(
		err,
		ErrIdempotencyPayloadConflict,
	) {
		t.Fatalf(
			"idempotency conflict: got %v",
			err,
		)
	}

	var userID int

	if err := database.DB.QueryRowContext(
		ctx,
		`
		SELECT id
		FROM users
		WHERE lower(email) = 'usuario@poliredi.local'
		LIMIT 1
		`,
	).Scan(&userID); err != nil {
		t.Fatalf("load user: %v", err)
	}

	notificationTitle := fmt.Sprintf(
		"Integracion %d",
		time.Now().UnixNano(),
	)

	if _, err := database.DB.ExecContext(
		ctx,
		`
		INSERT INTO notifications (
			user_id,
			reservation_id,
			title,
			message,
			type,
			is_read
		)
		VALUES (
			$1,
			NULL,
			$2,
			'Prueba de integracion PostgreSQL',
			'INTEGRATION_TEST',
			false
		)
		`,
		userID,
		notificationTitle,
	); err != nil {
		t.Fatalf(
			"insert notification: %v",
			err,
		)
	}

	notifications, err :=
		GetNotificationsByUserID(userID)
	if err != nil {
		t.Fatalf(
			"get notifications: %v",
			err,
		)
	}

	found := false

	for _, notification := range notifications {
		if notification.Title ==
			notificationTitle {
			found = true

			if notification.UserID != userID {
				t.Fatalf(
					"notification user: got %d want %d",
					notification.UserID,
					userID,
				)
			}

			if notification.Type !=
				"INTEGRATION_TEST" {
				t.Fatalf(
					"notification type: %q",
					notification.Type,
				)
			}

			break
		}
	}

	if !found {
		t.Fatalf(
			"inserted notification not returned",
		)
	}
}

func loadGroupResourceIDsForIntegration(
	ctx context.Context,
	policyID int,
) ([]int, error) {
	rows, err := database.DB.QueryContext(
		ctx,
		`
		SELECT resource_id
		FROM reservation_policy_group_resources
		WHERE policy_id = $1
		ORDER BY resource_id
		`,
		policyID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	values := []int{}

	for rows.Next() {
		var value int

		if err := rows.Scan(&value); err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return values, nil
}
