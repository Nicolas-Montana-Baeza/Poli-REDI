package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/database"

	"github.com/gofiber/fiber/v2"
)

func TestNotificationsAdminCancellationHTTPPostgres(t *testing.T) {
	if os.Getenv("POLIREDI_INTEGRATION") != "1" ||
		os.Getenv("POLIREDI_3C_ISOLATED") != "1" {
		t.Skip("requiere verificador PostgreSQL efimero")
	}
	if !strings.Contains(os.Getenv("DATABASE_URL"), "@127.0.0.1:") {
		t.Fatal("se requiere DATABASE_URL de PostgreSQL efimero local")
	}

	t.Setenv("DEV_AUTH_ENABLED", "true")
	t.Setenv("MVP_SCOPE", "mvp3")

	database.Close()
	if err := database.Connect(); err != nil {
		t.Fatalf("connect PostgreSQL: %v", err)
	}
	t.Cleanup(database.Close)

	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	var policyID, duration, venueID, adminID int
	var adminEmail string

	err = tx.QueryRowContext(ctx, `
		SELECT p.id,
		       (SELECT MIN(d.duration_minutes)
		        FROM reservation_policy_durations d
		        WHERE d.policy_id = p.id)
		FROM reservation_policies p
		WHERE p.is_published = true
		  AND p.effective_from <= CURRENT_TIMESTAMP
		  AND (p.effective_to IS NULL OR p.effective_to > CURRENT_TIMESTAMP)
		ORDER BY p.effective_from DESC, p.id DESC
		LIMIT 1
	`).Scan(&policyID, &duration)
	if err != nil || duration <= 0 {
		t.Fatalf("load current policy: duration=%d err=%v", duration, err)
	}

	err = tx.QueryRowContext(ctx, `
		SELECT id, email FROM users
		WHERE is_admin = true AND is_blocked = false
		ORDER BY id LIMIT 1
	`).Scan(&adminID, &adminEmail)
	if err != nil {
		t.Fatalf("load administrator: %v", err)
	}

	err = tx.QueryRowContext(ctx, `
		SELECT id FROM venues
		WHERE is_active = true ORDER BY id LIMIT 1
	`).Scan(&venueID)
	if err != nil {
		t.Fatalf("load venue: %v", err)
	}

	stamp := time.Now().UnixNano()
	ownerEmail := fmt.Sprintf("notif.http.%d@test.local", stamp)

	var ownerID, resourceID, reservationID int

	err = tx.QueryRowContext(ctx, `
		INSERT INTO users (email, full_name, rut, is_admin, is_blocked)
		VALUES ($1, 'HTTP Notification Owner', $2, false, false)
		RETURNING id
	`, ownerEmail, fmt.Sprintf("%08d-5", stamp%90000000)).Scan(&ownerID)
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO resources (
			venue_id, name, type, reservation_mode, capacity, is_active
		)
		VALUES ($1, $2, 'Test', 'RESERVABLE', 20, true)
		RETURNING id
	`, venueID, fmt.Sprintf("HTTP Notification Test %d", stamp)).Scan(&resourceID)
	if err != nil {
		t.Fatalf("create resource: %v", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO reservation_policy_resources (policy_id, resource_id)
		VALUES ($1, $2)
	`, policyID, resourceID)
	if err != nil {
		t.Fatalf("configure resource: %v", err)
	}

	now := businessclock.Now()
	start := time.Date(
		now.Year(), now.Month(), now.Day(),
		20, 0, 0, 0, businessclock.Location(),
	).AddDate(0, 0, 1)

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

	// Las rutas HTTP usan transacciones independientes.
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit fixture: %v", err)
	}

	t.Cleanup(func() {
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
			if _, err := database.DB.ExecContext(
				context.Background(), item.query, item.id,
			); err != nil {
				t.Errorf("cleanup: %v", err)
			}
		}
	})

	app := fiber.New()
	RegisterRoutes(app)

	request := func(method, path, email, body string) (int, []byte) {
		t.Helper()

		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if email != "" {
			req.Header.Set("X-Dev-Auth-Email", email)
		}
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}

		response, err := app.Test(req, 30000)
		if err != nil {
			t.Fatalf("%s %s: %v", method, path, err)
		}
		defer response.Body.Close()

		data, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		return response.StatusCode, data
	}

	// Sin identidad, la API debe rechazar la consulta.
	status, body := request(http.MethodGet, "/api/notifications", "", "")
	if status != http.StatusUnauthorized {
		t.Fatalf("unauthenticated GET: %d %s", status, body)
	}

	// La cancelacion se inicia a traves de HTTP.
	status, body = request(
		http.MethodPatch,
		"/api/reservations/cancel",
		adminEmail,
		fmt.Sprintf(`{"reservationId":%d}`, reservationID),
	)
	if status != http.StatusOK {
		t.Fatalf("HTTP cancellation: %d %s", status, body)
	}

	var cancelled struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &cancelled); err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != "CANCELLED" {
		t.Fatalf("unexpected cancellation status: %s", cancelled.Status)
	}

	type notification struct {
		ID            int    `json:"id"`
		UserID        int    `json:"userId"`
		ReservationID *int   `json:"reservationId"`
		Type          string `json:"type"`
		IsRead        bool   `json:"isRead"`
	}

	getList := func(email string) []notification {
		t.Helper()
		code, payload := request(http.MethodGet, "/api/notifications", email, "")
		if code != http.StatusOK {
			t.Fatalf("GET notifications: %d %s", code, payload)
		}

		var list []notification
		if err := json.Unmarshal(payload, &list); err != nil {
			t.Fatal(err)
		}
		return list
	}

	ownerList := getList(ownerEmail)
	if len(ownerList) != 1 {
		t.Fatalf("expected one owner notification, got %d", len(ownerList))
	}

	n := ownerList[0]
	if n.UserID != ownerID ||
		n.ReservationID == nil ||
		*n.ReservationID != reservationID ||
		n.Type != "RESERVATION_CANCELLED" ||
		n.IsRead {
		t.Fatalf("unexpected owner notification: %+v", n)
	}

	for _, other := range getList(adminEmail) {
		if other.ReservationID != nil &&
			*other.ReservationID == reservationID {
			t.Fatal("administrator received owner's notification")
		}
	}

	readPath := fmt.Sprintf("/api/notifications/%d/read", n.ID)

	// Un tercero no puede marcar la notificacion ajena.
	status, body = request(http.MethodPatch, readPath, adminEmail, "")
	if status != http.StatusNotFound {
		t.Fatalf("foreign notification read: %d %s", status, body)
	}
	if getList(ownerEmail)[0].IsRead {
		t.Fatal("foreign request modified owner notification")
	}

	// El propietario la marca como leida mediante HTTP.
	status, body = request(http.MethodPatch, readPath, ownerEmail, "")
	if status != http.StatusOK {
		t.Fatalf("owner notification read: %d %s", status, body)
	}

	var updated notification
	if err := json.Unmarshal(body, &updated); err != nil {
		t.Fatal(err)
	}
	if !updated.IsRead || updated.ID != n.ID {
		t.Fatalf("incorrect PATCH result: %+v", updated)
	}
	if !getList(ownerEmail)[0].IsRead {
		t.Fatal("read state was not persisted")
	}

	// La cancelacion duplicada no debe generar otra notificacion.
	status, _ = request(
		http.MethodPatch,
		"/api/reservations/cancel",
		adminEmail,
		fmt.Sprintf(`{"reservationId":%d}`, reservationID),
	)
	if status != http.StatusBadRequest {
		t.Fatalf("duplicate cancellation should return 400, got %d", status)
	}

	var count int
	err = database.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM notifications WHERE reservation_id = $1
	`, reservationID).Scan(&count)
	if err != nil || count != 1 {
		t.Fatalf("notification duplication: count=%d err=%v", count, err)
	}

	t.Log("NOTIF-001 HTTP E2E: cancelacion, consulta, aislamiento, lectura y persistencia PASS")
}
