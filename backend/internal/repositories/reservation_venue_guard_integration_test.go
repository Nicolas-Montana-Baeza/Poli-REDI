//go:build integration

package repositories

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"poli-redi-api/internal/database"
)

func TestReservationVenueGuardPostgresIntegration(t *testing.T) {
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

	var venueID int

	err = tx.QueryRowContext(ctx, `
        INSERT INTO venues (name, country, is_active)
        VALUES ('Sede prueba guard MVP3', 'Chile', true)
        RETURNING id
    `).Scan(&venueID)

	if err != nil {
		t.Fatalf("crear sede: %v", err)
	}

	var resourceID int

	err = tx.QueryRowContext(ctx, `
        INSERT INTO resources (
            venue_id, name, type,
            reservation_mode, capacity, is_active
        )
        VALUES (
            $1, 'Recurso prueba guard MVP3', 'Cancha',
            'RESERVABLE', 20, true
        )
        RETURNING id
    `, venueID).Scan(&resourceID)

	if err != nil {
		t.Fatalf("crear recurso: %v", err)
	}

	var userID int

	err = tx.QueryRowContext(ctx, `
        SELECT id
        FROM users
        WHERE NOT is_blocked
          AND (is_admin OR rut IS NOT NULL)
        ORDER BY id
        LIMIT 1
    `).Scan(&userID)

	if err != nil {
		t.Fatalf("obtener usuario de prueba: %v", err)
	}

	var policyID int

	err = tx.QueryRowContext(ctx, `
        SELECT id
        FROM reservation_policies
        ORDER BY id
        LIMIT 1
    `).Scan(&policyID)

	if err != nil {
		t.Fatalf("obtener politica de prueba: %v", err)
	}

	// Simula una actualizacion SQL directa que elude
	// las protecciones del repositorio administrativo.
	_, err = tx.ExecContext(ctx, `
        UPDATE venues
        SET is_active = false
        WHERE id = $1
    `, venueID)

	if err != nil {
		t.Fatalf("desactivar sede: %v", err)
	}

	_, err = tx.ExecContext(ctx, `SAVEPOINT before_reservation`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = tx.ExecContext(ctx, `
        INSERT INTO reservations (
            policy_id,
            user_id,
            resource_id,
            start_time,
            duration_minutes,
            status
        )
        VALUES ($1, $2, $3, $4, 60, 'CONFIRMED')
    `,
		policyID,
		userID,
		resourceID,
		time.Now().Add(24*time.Hour),
	)

	var pgErr *pgconn.PgError

	if !errors.As(err, &pgErr) || pgErr.Code != "P1002" {
		t.Fatalf(
			"esperaba rechazo P1002 por sede inactiva, obtuve: %v",
			err,
		)
	}

	// Recupera la transaccion tras el error esperado.
	_, err = tx.ExecContext(ctx, `
        ROLLBACK TO SAVEPOINT before_reservation
    `)
	if err != nil {
		t.Fatal(err)
	}

	var count int

	err = tx.QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM reservations
        WHERE resource_id = $1
    `, resourceID).Scan(&count)

	if err != nil {
		t.Fatal(err)
	}

	if count != 0 {
		t.Fatalf(
			"se insertaron %d reservas pese a la sede inactiva",
			count,
		)
	}
}
