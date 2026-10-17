//go:build integration

package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"poli-redi-api/internal/database"
)

func TestReservationVenueConcurrencyPostgresIntegration(t *testing.T) {
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

	ctx, cancel := context.WithTimeout(
		context.Background(), 15*time.Second,
	)
	defer cancel()

	name := fmt.Sprintf(
		"Sede concurrencia 3D11 %d", time.Now().UnixNano(),
	)

	var venueID, resourceID int

	err := database.DB.QueryRowContext(ctx, `
        INSERT INTO venues (name, country)
        VALUES ($1, 'Chile')
        RETURNING id
    `, name).Scan(&venueID)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, _ = database.DB.Exec(
			"DELETE FROM resources WHERE id = $1", resourceID,
		)
		_, _ = database.DB.Exec(
			"DELETE FROM venues WHERE id = $1", venueID,
		)
	})

	err = database.DB.QueryRowContext(ctx, `
        INSERT INTO resources (
            venue_id, name, type,
            reservation_mode, capacity, is_active
        )
        VALUES (
            $1, 'Cancha concurrencia', 'Cancha',
            'RESERVABLE', 20, true
        )
        RETURNING id
    `, venueID).Scan(&resourceID)
	if err != nil {
		t.Fatal(err)
	}

	var userID, policyID int

	err = database.DB.QueryRowContext(ctx, `
        SELECT id FROM users
        WHERE NOT is_blocked
          AND (is_admin OR rut IS NOT NULL)
        ORDER BY id LIMIT 1
    `).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	err = database.DB.QueryRowContext(ctx, `
        SELECT id FROM reservation_policies
        ORDER BY id LIMIT 1
    `).Scan(&policyID)
	if err != nil {
		t.Fatal(err)
	}

	// Conexión independiente para intentar insertar la reserva.
	insertConn, err := database.DB.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer insertConn.Close()

	var insertPID int
	if err := insertConn.QueryRowContext(
		ctx, "SELECT pg_backend_pid()",
	).Scan(&insertPID); err != nil {
		t.Fatal(err)
	}

	// Primera transacción: desactivación sin confirmar.
	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
        UPDATE venues
        SET is_active = false
        WHERE id = $1
    `, venueID)
	if err != nil {
		t.Fatal(err)
	}

	// Segunda conexión: intenta reservar mientras la sede
	// permanece bloqueada por la primera transacción.
	result := make(chan error, 1)

	go func() {
		_, insertErr := insertConn.ExecContext(ctx, `
            INSERT INTO reservations (
                policy_id, user_id, resource_id,
                start_time, duration_minutes, status
            )
            VALUES ($1, $2, $3, $4, 60, 'CONFIRMED')
        `,
			policyID, userID, resourceID,
			time.Now().Add(24*time.Hour),
		)
		result <- insertErr
	}()

	// Confirmar que PostgreSQL realmente está esperando
	// un bloqueo, no simplemente que el goroutine sea lento.
	waiting := false
	deadline := time.Now().Add(3 * time.Second)

	for time.Now().Before(deadline) {
		select {
		case err := <-result:
			t.Fatalf(
				"insercion termino antes del commit: %v", err,
			)
		default:
		}

		var waitType sql.NullString

		err := database.DB.QueryRowContext(ctx, `
            SELECT wait_event_type
            FROM pg_stat_activity
            WHERE pid = $1
        `, insertPID).Scan(&waitType)

		if err != nil {
			t.Fatal(err)
		}
		if waitType.Valid && waitType.String == "Lock" {
			waiting = true
			break
		}

		time.Sleep(25 * time.Millisecond)
	}

	if !waiting {
		t.Fatal("no se observo espera de bloqueo en PostgreSQL")
	}

	// Confirmar la desactivación.
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	select {
	case insertErr := <-result:
		var pgErr *pgconn.PgError
		if !errors.As(insertErr, &pgErr) ||
			pgErr.Code != "P1002" {
			t.Fatalf(
				"esperaba P1002 tras desactivar sede: %v",
				insertErr,
			)
		}
	case <-ctx.Done():
		t.Fatal("timeout esperando rechazo de reserva")
	}

	var count int
	err = database.DB.QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM reservations
        WHERE resource_id = $1
    `, resourceID).Scan(&count)

	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("se crearon %d reservas indebidamente", count)
	}

	t.Log("Bloqueo concurrente y rechazo P1002: PASS")
}
