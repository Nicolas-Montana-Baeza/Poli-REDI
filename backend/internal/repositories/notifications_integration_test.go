//go:build integration

package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"poli-redi-api/internal/database"
)

func TestNotificationsPostgresIntegration(t *testing.T) {
	if err := database.Connect(); err != nil {
		t.Fatalf("connect PostgreSQL: %v", err)
	}
	defer database.Close()

	ctx := context.Background()

	rows, err := database.DB.QueryContext(
		ctx,
		`
		SELECT id
		FROM users
		WHERE is_blocked = false
		ORDER BY id
		LIMIT 2
		`,
	)
	if err != nil {
		t.Fatalf("load users: %v", err)
	}
	defer rows.Close()

	userIDs := []int{}

	for rows.Next() {
		var id int

		if err := rows.Scan(&id); err != nil {
			t.Fatalf("scan user: %v", err)
		}

		userIDs = append(userIDs, id)
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("iterate users: %v", err)
	}

	if len(userIDs) < 2 {
		t.Fatal("integration test requires at least two active users")
	}

	userA := userIDs[0]
	userB := userIDs[1]

	suffix := time.Now().UnixNano()

	createNotification := func(
		userID int,
		label string,
	) int {
		t.Helper()

		var id int

		err := database.DB.QueryRowContext(
			ctx,
			`
			INSERT INTO notifications (
				user_id,
				reservation_id,
				title,
				message,
				type
			)
			VALUES (
				$1,
				NULL,
				$2,
				$3,
				'SYSTEM'
			)
			RETURNING id
			`,
			userID,
			fmt.Sprintf(
				"Integration %s %d",
				label,
				suffix,
			),
			"Notificacion temporal de prueba",
		).Scan(&id)

		if err != nil {
			t.Fatalf(
				"create notification for user %d: %v",
				userID,
				err,
			)
		}

		return id
	}

	notificationA := createNotification(userA, "A")
	notificationB := createNotification(userB, "B")

	defer func() {
		_, _ = database.DB.ExecContext(
			context.Background(),
			`
			DELETE FROM notifications
			WHERE id = ANY($1)
			`,
			[]int{notificationA, notificationB},
		)
	}()

	// ------------------------------------------------------------
	// Usuario A puede marcar SU notificación.
	// ------------------------------------------------------------

	updated, err := MarkNotificationRead(
		userA,
		notificationA,
	)

	if err != nil {
		t.Fatalf("mark own notification read: %v", err)
	}

	if !updated.IsRead {
		t.Fatal("own notification should be marked read")
	}

	if updated.UserID != userA {
		t.Fatalf(
			"notification belongs to %d, expected %d",
			updated.UserID,
			userA,
		)
	}

	// ------------------------------------------------------------
	// Usuario A NO puede modificar la notificación de B.
	// ------------------------------------------------------------

	_, err = MarkNotificationRead(
		userA,
		notificationB,
	)

	if !errors.Is(err, ErrNotificationNotFound) {
		t.Fatalf(
			"expected ErrNotificationNotFound, got %v",
			err,
		)
	}

	// Confirmamos directamente que B sigue unread.
	var otherIsRead bool

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT is_read
		FROM notifications
		WHERE id = $1
		`,
		notificationB,
	).Scan(&otherIsRead)

	if err != nil {
		t.Fatalf("verify foreign notification: %v", err)
	}

	if otherIsRead {
		t.Fatal(
			"foreign notification must remain unread",
		)
	}

	// ------------------------------------------------------------
	// Listado propio.
	// ------------------------------------------------------------

	notifications, err :=
		GetNotificationsByUserID(userA)

	if err != nil {
		t.Fatalf("get own notifications: %v", err)
	}

	found := false

	for _, notification := range notifications {
		if notification.ID == notificationA {
			found = true

			if !notification.IsRead {
				t.Fatal(
					"listed notification should be read",
				)
			}
		}

		if notification.ID == notificationB {
			t.Fatal(
				"user A received user B notification",
			)
		}
	}

	if !found {
		t.Fatal(
			"own integration notification was not listed",
		)
	}
}
