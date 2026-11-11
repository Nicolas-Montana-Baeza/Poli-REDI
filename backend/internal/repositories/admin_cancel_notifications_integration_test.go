package repositories

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func TestAdminCancellationNotificationPostgresIntegration(t *testing.T) {
	if os.Getenv("POLIREDI_INTEGRATION") != "1" ||
		os.Getenv("MVP_SCOPE") != "mvp3" {
		t.Skip("requiere integracion PostgreSQL con MVP3")
	}

	database.Close()
	if err := database.Connect(); err != nil {
		t.Fatalf("connect PostgreSQL: %v", err)
	}
	defer database.Close()

	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	var policyID, windowDays, opening, closing, interval, duration int
	err = tx.QueryRowContext(ctx, `
		SELECT p.id, p.reservable_window_days,
		       p.opening_minute, p.closing_minute,
		       p.slot_interval_minutes,
		       (SELECT MIN(d.duration_minutes)
		        FROM reservation_policy_durations d
		        WHERE d.policy_id = p.id)
		FROM reservation_policies p
		WHERE p.is_published = true
		  AND p.effective_from <= CURRENT_TIMESTAMP
		  AND (p.effective_to IS NULL OR p.effective_to > CURRENT_TIMESTAMP)
		ORDER BY p.effective_from DESC, p.id DESC
		LIMIT 1
	`).Scan(&policyID, &windowDays, &opening, &closing, &interval, &duration)
	if err != nil {
		t.Fatalf("load policy: %v", err)
	}
	if interval <= 0 {
		t.Fatalf("invalid slot interval: %d", interval)
	}

	now := businessclock.Now()
	start := firstFuturePolicySlot(now, windowDays, opening, closing, interval, duration)
	if start.IsZero() {
		t.Fatal("no future policy slot available")
	}

	var adminID, venueID int
	if err := tx.QueryRowContext(ctx, `
		SELECT id FROM users
		WHERE is_admin = true AND is_blocked = false
		ORDER BY id LIMIT 1
	`).Scan(&adminID); err != nil {
		t.Fatalf("load admin: %v", err)
	}

	if err := tx.QueryRowContext(ctx, `
		SELECT id FROM venues WHERE is_active = true
		ORDER BY id LIMIT 1
	`).Scan(&venueID); err != nil {
		t.Fatalf("load venue: %v", err)
	}

	stamp := time.Now().UnixNano()
	var ownerID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO users (email, full_name, rut, is_admin, is_blocked)
		VALUES ($1, 'Notification Test Owner', $2, false, false)
		RETURNING id
	`,
		fmt.Sprintf("admin-cancel.%d@test.local", stamp),
		fmt.Sprintf("%08d-5", stamp%90000000),
	).Scan(&ownerID)
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}

	var resourceID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO resources (
			venue_id, name, type, reservation_mode, capacity, is_active
		)
		VALUES ($1, $2, 'Test', 'RESERVABLE', 20, true)
		RETURNING id
	`, venueID, fmt.Sprintf("Admin Cancel Test %d", stamp)).Scan(&resourceID)
	if err != nil {
		t.Fatalf("create resource: %v", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO reservation_policy_resources (policy_id, resource_id)
		VALUES ($1, $2)
	`, policyID, resourceID)
	if err != nil {
		t.Fatalf("scope resource: %v", err)
	}

	var reservationID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO reservations (
			policy_id, user_id, resource_id,
			start_time, duration_minutes, status
		)
		VALUES ($1, $2, $3, $4, $5, 'CONFIRMED')
		RETURNING id
	`, policyID, ownerID, resourceID, start, duration).Scan(&reservationID)
	if err != nil {
		t.Fatalf("create reservation: %v", err)
	}

	// La operación real abre su propia transacción.
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit fixture: %v", err)
	}

	defer func() {
		for _, item := range []struct {
			query string
			id    int
		}{
			{"DELETE FROM notifications WHERE reservation_id = $1", reservationID},
			{"DELETE FROM reservations WHERE id = $1", reservationID},
			{"DELETE FROM reservation_policy_resources WHERE resource_id = $1", resourceID},
			{"DELETE FROM resources WHERE id = $1", resourceID},
			{"DELETE FROM users WHERE id = $1", ownerID},
		} {
			if _, err := database.DB.ExecContext(context.Background(), item.query, item.id); err != nil {
				t.Errorf("cleanup failed: %v", err)
			}
		}
	}()

	admin := models.LocalAuthUser{ID: adminID, IsAdmin: true}

	cancelled, err := CancelReservationAuthorized(
		reservationID, admin, businessclock.Now(),
	)
	if err != nil {
		t.Fatalf("admin cancellation: %v", err)
	}
	if cancelled.Status != models.ReservationStatusCancelled {
		t.Fatalf("expected CANCELLED, got %s", cancelled.Status)
	}

	assertNotification := func() {
		t.Helper()
		var total, correct, unread, adminCount int

		err := database.DB.QueryRowContext(ctx, `
			SELECT COUNT(*),
			       COUNT(*) FILTER (
			         WHERE user_id = $2 AND type = 'RESERVATION_CANCELLED'
			       ),
			       COUNT(*) FILTER (
			         WHERE user_id = $2 AND type = 'RESERVATION_CANCELLED'
			           AND is_read = false
			       ),
			       COUNT(*) FILTER (WHERE user_id = $3)
			FROM notifications
			WHERE reservation_id = $1
		`, reservationID, ownerID, adminID).Scan(
			&total, &correct, &unread, &adminCount,
		)
		if err != nil {
			t.Fatalf("query notifications: %v", err)
		}
		if total != 1 || correct != 1 || unread != 1 || adminCount != 0 {
			t.Fatalf(
				"unexpected notifications: total=%d correct=%d unread=%d admin=%d",
				total, correct, unread, adminCount,
			)
		}
	}

	assertNotification()

	_, err = CancelReservationAuthorized(
		reservationID, admin, businessclock.Now(),
	)
	if !errors.Is(err, ErrReservationNotCancellable) {
		t.Fatalf("expected ErrReservationNotCancellable, got %v", err)
	}

	assertNotification()

	ownerNotifications, err := GetNotificationsByUserID(ownerID)
	if err != nil {
		t.Fatalf("read owner notifications: %v", err)
	}
	if len(ownerNotifications) != 1 ||
		ownerNotifications[0].Type != "RESERVATION_CANCELLED" {
		t.Fatalf("unexpected owner notification listing: %+v", ownerNotifications)
	}

	t.Log("NOTIF-001: cancelacion administrativa, destinatario e idempotencia PASS")
}
