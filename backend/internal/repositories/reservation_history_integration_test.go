//go:build integration

package repositories

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"poli-redi-api/internal/database"
)

func TestReservationHistoryPostgresIntegration(t *testing.T) {
	if os.Getenv("POLIREDI_3C_ISOLATED") != "1" {
		t.Skip("requiere PostgreSQL efimero")
	}
	if os.Getenv("DATABASE_URL") == "" {
		t.Fatal("DATABASE_URL es obligatorio")
	}

	if err := database.Connect(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(database.Close)

	ctx := context.Background()

	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback()
	})

	// Crear sede y recurso exclusivamente para esta prueba.
	var venueID int
	err = tx.QueryRowContext(ctx, `
        INSERT INTO venues (name, country, is_active)
        VALUES ('Sede historial 3D10', 'Chile', true)
        RETURNING id
    `).Scan(&venueID)
	if err != nil {
		t.Fatal(err)
	}

	var resourceID int
	err = tx.QueryRowContext(ctx, `
        INSERT INTO resources (
            venue_id, name, type,
            reservation_mode, capacity, is_active
        )
        VALUES (
            $1, 'Cancha historial 3D10', 'Cancha',
            'RESERVABLE', 20, true
        )
        RETURNING id
    `, venueID).Scan(&resourceID)
	if err != nil {
		t.Fatal(err)
	}

	var policyID, userID int

	err = tx.QueryRowContext(ctx, `
        SELECT id
        FROM reservation_policies
        ORDER BY id
        LIMIT 1
    `).Scan(&policyID)
	if err != nil {
		t.Fatal(err)
	}

	err = tx.QueryRowContext(ctx, `
        SELECT id
        FROM users
        WHERE NOT is_blocked
          AND (is_admin OR rut IS NOT NULL)
        ORDER BY id
        LIMIT 1
    `).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	// Fixture de reserva historica ya cancelada.
	// Permite verificar integridad y snapshots sin
	// depender de la politica reservable del dia actual.
	var reservationID int
	err = tx.QueryRowContext(ctx, `
        INSERT INTO reservations (
            policy_id,
            user_id,
            resource_id,
            start_time,
            duration_minutes,
            status,
            reservation_mode_snapshot,
            group_capacity_snapshot,
            group_minimum_participants_snapshot
        )
        VALUES (
            $1, $2, $3, $4, 60, 'CANCELLED',
            'RESERVABLE', 20, 10
        )
        RETURNING id
    `,
		policyID,
		userID,
		resourceID,
		time.Now().Add(-24*time.Hour),
	).Scan(&reservationID)
	if err != nil {
		t.Fatalf("crear historial de prueba: %v", err)
	}

	type snapshot struct {
		Mode     string
		Capacity int
		Minimum  int
		Status   string
		Start    time.Time
		End      time.Time
	}

	readSnapshot := func() (snapshot, error) {
		var s snapshot
		err := tx.QueryRowContext(ctx, `
            SELECT
                reservation_mode_snapshot,
                group_capacity_snapshot,
                group_minimum_participants_snapshot,
                status,
                start_time,
                end_time
            FROM reservations
            WHERE id = $1
              AND resource_id = $2
        `, reservationID, resourceID).Scan(
			&s.Mode,
			&s.Capacity,
			&s.Minimum,
			&s.Status,
			&s.Start,
			&s.End,
		)
		return s, err
	}

	before, err := readSnapshot()
	if err != nil {
		t.Fatal(err)
	}

	// Cambiar la configuracion actual del recurso.
	// Esto no debe alterar la reserva historica.
	_, err = tx.ExecContext(ctx, `
        UPDATE resources
        SET
            capacity = 12,
            reservation_mode = 'ADMIN_ONLY',
            is_active = false
        WHERE id = $1
    `, resourceID)
	if err != nil {
		t.Fatalf("desactivar recurso: %v", err)
	}

	_, err = tx.ExecContext(ctx, `
        UPDATE venues
        SET is_active = false
        WHERE id = $1
    `, venueID)
	if err != nil {
		t.Fatalf("desactivar sede: %v", err)
	}

	after, err := readSnapshot()
	if err != nil {
		t.Fatalf("reserva historica desaparecio: %v", err)
	}

	if !reflect.DeepEqual(before, after) {
		t.Fatalf(
			"historial alterado:\nantes: %+v\ndespues: %+v",
			before,
			after,
		)
	}

	// Confirmar que se preserva la asociacion con la sede.
	var storedVenueID int
	err = tx.QueryRowContext(ctx, `
        SELECT r.venue_id
        FROM reservations reservation
        JOIN resources r ON r.id = reservation.resource_id
        WHERE reservation.id = $1
    `, reservationID).Scan(&storedVenueID)

	if err != nil || storedVenueID != venueID {
		t.Fatalf("asociacion historica alterada: %v", err)
	}

	// Verificar ON DELETE RESTRICT sin perder la transaccion.
	attempts := []struct {
		name  string
		query string
		id    int
	}{
		{
			name:  "recurso con historial",
			query: "DELETE FROM resources WHERE id = $1",
			id:    resourceID,
		},
		{
			name:  "sede con recurso asociado",
			query: "DELETE FROM venues WHERE id = $1",
			id:    venueID,
		},
	}

	for _, attempt := range attempts {
		if _, err := tx.ExecContext(
			ctx, "SAVEPOINT history_delete_guard",
		); err != nil {
			t.Fatal(err)
		}

		_, deleteErr := tx.ExecContext(
			ctx, attempt.query, attempt.id,
		)

		var pgErr *pgconn.PgError
		if !errors.As(deleteErr, &pgErr) ||
			pgErr.Code != "23503" {
			t.Fatalf(
				"%s: esperaba violacion FK 23503; obtuve %v",
				attempt.name, deleteErr,
			)
		}

		if _, err := tx.ExecContext(
			ctx, "ROLLBACK TO SAVEPOINT history_delete_guard",
		); err != nil {
			t.Fatal(err)
		}

		if _, err := tx.ExecContext(
			ctx, "RELEASE SAVEPOINT history_delete_guard",
		); err != nil {
			t.Fatal(err)
		}
	}

	// Confirmar que los intentos rechazados no alteraron
	// el historial ni las relaciones.
	preserved, err := readSnapshot()
	if err != nil {
		t.Fatalf("historial perdido tras DELETE rechazado: %v", err)
	}

	if !reflect.DeepEqual(before, preserved) {
		t.Fatalf("snapshots alterados tras DELETE rechazado")
	}

	var relatedVenueID int

	err = tx.QueryRowContext(ctx, `
        SELECT resource.venue_id
        FROM reservations reservation
        JOIN resources resource
          ON resource.id = reservation.resource_id
        WHERE reservation.id = $1
    `, reservationID).Scan(&relatedVenueID)

	if err != nil || relatedVenueID != venueID {
		t.Fatalf(
			"relacion historica alterada tras DELETE: %v",
			err,
		)
	}

	t.Log("Historial, snapshots y restricciones DELETE: PASS")
}
