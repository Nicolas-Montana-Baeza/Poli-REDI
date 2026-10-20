//go:build integration

package repositories

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

// Requiere el PostgreSQL efimero con PG16_0012 aplicada.
func TestAdminAvailabilityBlocksPostgresIntegration(t *testing.T) {
	if os.Getenv("POLIREDI_3C_ISOLATED") != "1" {
		t.Skip("requiere PostgreSQL efimero")
	}
	if os.Getenv("DATABASE_URL") == "" {
		t.Fatal("DATABASE_URL obligatorio")
	}
	if err := database.Connect(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(database.Close)
	ctx := context.Background()
	suffix := time.Now().UnixNano()

	var venueID, resourceID, userID int
	err := database.DB.QueryRowContext(ctx,
		`INSERT INTO venues (name,country) VALUES ($1,'Chile') RETURNING id`,
		fmt.Sprintf("Sede bloqueos 4A %d", suffix)).Scan(&venueID)
	if err != nil {
		t.Fatal(err)
	}
	err = database.DB.QueryRowContext(ctx,
		`INSERT INTO resources (venue_id,name,type,reservation_mode,is_active)
         VALUES ($1,'Cancha de bloqueos','COURT','RESERVABLE',true) RETURNING id`, venueID).Scan(&resourceID)
	if err != nil {
		t.Fatal(err)
	}
	err = database.DB.QueryRowContext(ctx,
		`INSERT INTO users (email,full_name,is_admin) VALUES ($1,'Prueba 4A',true) RETURNING id`,
		fmt.Sprintf("blocks-%d@example.org", suffix)).Scan(&userID)
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Hour)
	input := models.AdminAvailabilityBlock{
		ResourceID: resourceID, CreatedByUserID: userID,
		BlockType: "MAINTENANCE", Reason: "Prueba 4A", StartTime: start, EndTime: start.Add(time.Hour),
	}
	first, err := CreateAdminAvailabilityBlock(input)
	if err != nil {
		t.Fatalf("crear: %v", err)
	}
	if first.ID <= 0 || first.ResourceID != resourceID || first.CreatedByUserID != userID || !first.IsActive {
		t.Fatalf("bloqueo creado incorrectamente: %+v", first)
	}
	listed, err := ListAdminAvailabilityBlocks(start.Add(-time.Minute), start.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].ID != first.ID {
		t.Fatalf("listar: %+v", listed)
	}

	_, err = CreateAdminAvailabilityBlock(input)
	if !errors.Is(err, ErrBlockOverlap) {
		t.Fatalf("bloqueo solapado debe rechazarse: %v", err)
	}

	// Intervalos adyacentes no se solapan.
	adjacent := input
	adjacent.StartTime = input.EndTime
	adjacent.EndTime = adjacent.StartTime.Add(time.Hour)
	second, err := CreateAdminAvailabilityBlock(adjacent)
	if err != nil || second.ID <= 0 {
		t.Fatalf("adyacente: %v", err)
	}

	// Una vez desactivado, no bloquea el horario y conserva historial.
	disabled, err := DeactivateAdminAvailabilityBlock(first.ID)
	if err != nil || disabled.IsActive {
		t.Fatalf("desactivar: %v", err)
	}
	disabled, err = DeactivateAdminAvailabilityBlock(first.ID)
	if err != nil || disabled.IsActive {
		t.Fatalf("desactivacion idempotente: %v", err)
	}
	_, err = DeactivateAdminAvailabilityBlock(999999999)
	if !errors.Is(err, ErrBlockNotFound) {
		t.Fatalf("no encontrado: %v", err)
	}
	recreated, err := CreateAdminAvailabilityBlock(input)
	if err != nil || recreated.ID <= 0 {
		t.Fatalf("recrear despues de desactivar: %v", err)
	}

	// Dos administradores que intentan bloquear el mismo rango a la vez.
	competing := input
	competing.StartTime = start.Add(4 * time.Hour)
	competing.EndTime = competing.StartTime.Add(time.Hour)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := CreateAdminAvailabilityBlock(competing); results <- err }()
	}
	wg.Wait()
	close(results)
	successes, conflicts := 0, 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrBlockOverlap):
			conflicts++
		default:
			t.Fatalf("fallo de concurrencia inesperado: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrencia: %d exitos, %d conflictos", successes, conflicts)
	}

	_, err = database.DB.ExecContext(ctx, `UPDATE resources SET is_active=false WHERE id=$1`, resourceID)
	if err != nil {
		t.Fatal(err)
	}
	competing.StartTime = start.Add(6 * time.Hour)
	competing.EndTime = competing.StartTime.Add(time.Hour)
	_, err = CreateAdminAvailabilityBlock(competing)
	if !errors.Is(err, ErrBlockResourceUnavailable) {
		t.Fatalf("recurso inactivo: %v", err)
	}
}
