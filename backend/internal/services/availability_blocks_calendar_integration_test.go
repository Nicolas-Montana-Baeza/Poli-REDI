//go:build integration

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
)

// TestAdminBlockCalendarIntegration cubre el flujo real repositorio admin ->
// disponibilidad, sin abrir las rutas HTTP ni publicar el frontend.
// Requiere un PostgreSQL efimero con PG16_0012 aplicada.
func TestAdminBlockCalendarIntegration(t *testing.T) {
	if os.Getenv("POLIREDI_3C_ISOLATED") != "1" {
		t.Skip("requiere PostgreSQL efimero")
	}
	if os.Getenv("DATABASE_URL") == "" {
		t.Fatal("DATABASE_URL obligatorio")
	}

	database.Close()
	if err := database.Connect(); err != nil {
		t.Fatalf("connect PostgreSQL: %v", err)
	}
	defer database.Close()

	ctx := context.Background()
	suffix := time.Now().UnixNano()
	var venueID, resourceID, adminID, regularUserID int

	// La base es efimera, pero igualmente limpiamos lo creado por la prueba.
	defer func() {
		if resourceID > 0 {
			_, _ = database.DB.ExecContext(ctx,
				`DELETE FROM availability_blocks WHERE resource_id = $1`, resourceID)
			_, _ = database.DB.ExecContext(ctx,
				`DELETE FROM resources WHERE id = $1`, resourceID)
		}
		for _, id := range []int{adminID, regularUserID} {
			if id > 0 {
				_, _ = database.DB.ExecContext(ctx,
					`DELETE FROM users WHERE id = $1`, id)
			}
		}
		if venueID > 0 {
			_, _ = database.DB.ExecContext(ctx,
				`DELETE FROM venues WHERE id = $1`, venueID)
		}
	}()

	if err := database.DB.QueryRowContext(ctx,
		`INSERT INTO venues (name, country) VALUES ($1, 'Chile') RETURNING id`,
		fmt.Sprintf("Sede disponibilidad 4B %d", suffix),
	).Scan(&venueID); err != nil {
		t.Fatalf("crear sede: %v", err)
	}
	if err := database.DB.QueryRowContext(ctx, `
        INSERT INTO resources (venue_id, name, type, reservation_mode, is_active)
        VALUES ($1, $2, 'COURT', 'ADMIN_ONLY', true) RETURNING id`,
		venueID, fmt.Sprintf("Recurso disponibilidad 4B %d", suffix),
	).Scan(&resourceID); err != nil {
		t.Fatalf("crear recurso: %v", err)
	}
	if err := database.DB.QueryRowContext(ctx, `
        INSERT INTO users (email, full_name, is_admin)
        VALUES ($1, 'Admin calendario 4B', true) RETURNING id`,
		fmt.Sprintf("admin-4b-%d@test.local", suffix),
	).Scan(&adminID); err != nil {
		t.Fatalf("crear administrador: %v", err)
	}
	if err := database.DB.QueryRowContext(ctx, `
        INSERT INTO users (email, full_name, is_admin)
        VALUES ($1, 'Usuario calendario 4B', false) RETURNING id`,
		fmt.Sprintf("user-4b-%d@test.local", suffix),
	).Scan(&regularUserID); err != nil {
		t.Fatalf("crear usuario: %v", err)
	}

	// Horas futuras y rango semiabierto [inicio, termino).
	start := time.Now().UTC().Add(72 * time.Hour).Truncate(time.Hour)
	end := start.Add(time.Hour)
	const privateReason = "Motivo privado 4B - no exponer al calendario"
	input := models.AdminAvailabilityBlock{
		ResourceID:      resourceID,
		CreatedByUserID: adminID,
		BlockType:       "MAINTENANCE",
		Reason:          privateReason,
		StartTime:       start,
		EndTime:         end,
	}
	created, err := repositories.CreateAdminAvailabilityBlock(input)
	if err != nil {
		t.Fatalf("crear bloqueo por repositorio admin: %v", err)
	}
	if created.ID <= 0 || !created.IsActive {
		t.Fatalf("bloqueo invalido: %+v", created)
	}

	// Un administrador consulta detalles internos; el calendario no los expone.
	listed, err := repositories.ListAdminAvailabilityBlocks(start, end)
	if err != nil {
		t.Fatalf("listar bloqueos administrativos: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != created.ID || listed[0].Reason != privateReason {
		t.Fatalf("no se conserva el detalle administrativo: %+v", listed)
	}

	// La misma ocupacion debe ser visible al usuario normal y al admin.
	for _, role := range []struct {
		name  string
		id    int
		admin bool
	}{
		{name: "usuario", id: regularUserID, admin: false},
		{name: "administrador", id: adminID, admin: true},
	} {
		t.Run(role.name, func(t *testing.T) {
			items, err := GetAvailabilityItems(start.Add(-time.Minute), end.Add(time.Minute), role.id, role.admin)
			if err != nil {
				t.Fatalf("consultar disponibilidad: %v", err)
			}
			matches := calendarBlockItemsForResource(items, resourceID)
			if len(matches) != 1 {
				t.Fatalf("se esperaba un solo bloqueo, se encontraron %d", len(matches))
			}
			got := matches[0]
			if got.ID != created.ID || got.AvailabilityKey != fmt.Sprintf("block-%d", created.ID) ||
				got.Type != "blocked" || !got.IsAvailabilityBlock || got.IsScheduledActivity ||
				got.Status != "CONFIRMED" || got.Title != "No disponible" ||
				got.DurationMinutes != 60 || got.UserID != 0 || got.UserEmail != "" ||
				got.UserFullName != "" || got.UserRUT != "" {
				t.Fatalf("contrato de bloqueo incorrecto: %+v", got)
			}
			data, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(data), privateReason) || strings.Contains(string(data), "createdByUserId") {
				t.Fatalf("disponibilidad expone informacion privada: %s", data)
			}
		})
	}

	// Los intervalos adyacentes NO intersectan [start,end).
	for _, window := range [][2]time.Time{
		{start.Add(-time.Hour), start},
		{end, end.Add(time.Hour)},
	} {
		items, err := GetAvailabilityItems(window[0], window[1], regularUserID, false)
		if err != nil {
			t.Fatalf("disponibilidad fuera del intervalo: %v", err)
		}
		if got := calendarBlockItemsForResource(items, resourceID); len(got) != 0 {
			t.Fatalf("bloqueo incluido fuera del rango: %+v", got)
		}
	}

	deactivated, err := repositories.DeactivateAdminAvailabilityBlock(created.ID)
	if err != nil || deactivated.IsActive {
		t.Fatalf("desactivar bloqueo: %+v, %v", deactivated, err)
	}
	items, err := GetAvailabilityItems(start, end, regularUserID, false)
	if err != nil {
		t.Fatalf("consultar despues de desactivar: %v", err)
	}
	if got := calendarBlockItemsForResource(items, resourceID); len(got) != 0 {
		t.Fatalf("bloqueo inactivo aun ocupa disponibilidad: %+v", got)
	}
	listed, err = repositories.ListAdminAvailabilityBlocks(start, end)
	if err != nil || len(listed) != 1 || listed[0].IsActive {
		t.Fatalf("historial administrativo despues de desactivar: %+v, %v", listed, err)
	}
}

func calendarBlockItemsForResource(items []models.AvailabilityItem, resourceID int) []models.AvailabilityItem {
	result := make([]models.AvailabilityItem, 0)
	for _, item := range items {
		if item.ResourceID == resourceID && item.IsAvailabilityBlock {
			result = append(result, item)
		}
	}
	return result
}
