//go:build integration

package repositories

import (
	"context"
	"errors"
	"os"
	"testing"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func TestVenueAdministrationPostgresIntegration(t *testing.T) {
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

	address := "Direccion de prueba"
	latitude := -33.45
	longitude := -70.64

	// 1. Crear sede.
	venue, err := CreateVenue(models.Venue{
		Name:        "Sede integracion EV-012",
		Country:     "Chile",
		AddressLine: &address,
		Latitude:    &latitude,
		Longitude:   &longitude,
	})
	if err != nil {
		t.Fatalf("crear sede: %v", err)
	}
	if venue.ID <= 0 || !venue.IsActive {
		t.Fatalf("sede creada invalida: %+v", venue)
	}

	// 2. Consultar.
	stored, err := GetVenueByID(venue.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Name != venue.Name || stored.ResourceCount != 0 {
		t.Fatalf("consulta incorrecta: %+v", stored)
	}

	// 3. Rechazar duplicado.
	_, err = CreateVenue(models.Venue{
		Name:    venue.Name,
		Country: "Chile",
	})
	if !errors.Is(err, ErrVenueDuplicate) {
		t.Fatalf("esperaba duplicado, obtuve %v", err)
	}

	// 4. Editar y limpiar campos opcionales.
	newName := "Sede integracion actualizada"

	stored, err = UpdateVenue(venue.ID, VenuePatch{
		Name:               &newName,
		AddressLinePresent: true,
		AddressLine:        nil,
		LatitudePresent:    true,
		Latitude:           nil,
		LongitudePresent:   true,
		Longitude:          nil,
	})
	if err != nil {
		t.Fatalf("editar sede: %v", err)
	}

	if stored.Name != newName ||
		stored.AddressLine != nil ||
		stored.Latitude != nil ||
		stored.Longitude != nil {
		t.Fatalf("edicion incorrecta: %+v", stored)
	}

	// 5. Asociar un recurso activo a la sede.
	var resourceID int
	err = database.DB.QueryRowContext(ctx, `
		INSERT INTO resources (
			venue_id,
			name,
			type,
			reservation_mode,
			capacity,
			is_active
		)
		VALUES ($1, $2, 'Cancha', 'RESERVABLE', 20, true)
		RETURNING id
	`, venue.ID, "Recurso integracion EV-012").Scan(&resourceID)

	if err != nil {
		t.Fatalf("crear recurso de prueba: %v", err)
	}

	stored, err = GetVenueByID(venue.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.ResourceCount != 1 || stored.ActiveResourceCount != 1 {
		t.Fatalf("conteos incorrectos: %+v", stored)
	}

	// 6. Impedir desactivar sede con recurso activo.
	inactive := false

	_, err = UpdateVenue(venue.ID, VenuePatch{
		IsActive: &inactive,
	})
	if !errors.Is(err, ErrVenueHasActiveResources) {
		t.Fatalf("esperaba conflicto, obtuve %v", err)
	}

	stored, err = GetVenueByID(venue.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !stored.IsActive {
		t.Fatal("la sede cambio pese al rechazo")
	}

	// 7. Desactivar recurso y luego la sede.
	_, err = database.DB.ExecContext(ctx, `
		UPDATE resources
		SET is_active = false
		WHERE id = $1
	`, resourceID)
	if err != nil {
		t.Fatal(err)
	}

	stored, err = UpdateVenue(venue.ID, VenuePatch{
		IsActive: &inactive,
	})
	if err != nil {
		t.Fatalf("desactivar sede: %v", err)
	}
	if stored.IsActive || stored.ResourceCount != 1 ||
		stored.ActiveResourceCount != 0 {
		t.Fatalf("estado inesperado: %+v", stored)
	}

	// 8. Reactivar sede sin perder el recurso.
	active := true

	stored, err = UpdateVenue(venue.ID, VenuePatch{
		IsActive: &active,
	})
	if err != nil {
		t.Fatalf("reactivar sede: %v", err)
	}
	if !stored.IsActive || stored.ResourceCount != 1 {
		t.Fatalf("reactivacion incorrecta: %+v", stored)
	}

	// 9. Desactivar nuevamente la sede sin recursos activos.
	stored, err = UpdateVenue(venue.ID, VenuePatch{
		IsActive: &inactive,
	})
	if err != nil || stored.IsActive {
		t.Fatalf("desactivar sede para prueba de recurso: %v", err)
	}

	// 10. No permitir activar un recurso en sede inactiva.
	_, err = UpdateResourceOperationalSettings(
		resourceID,
		nil,
		false,
		nil,
		&active,
	)
	if !errors.Is(err, ErrResourceVenueInactive) {
		t.Fatalf(
			"esperaba rechazo por sede inactiva, obtuve %v",
			err,
		)
	}

	var resourceActive bool
	err = database.DB.QueryRowContext(ctx, `
		SELECT is_active
		FROM resources
		WHERE id = $1
	`, resourceID).Scan(&resourceActive)
	if err != nil {
		t.Fatal(err)
	}
	if resourceActive {
		t.Fatal("el recurso fue activado pese al rechazo")
	}

	// 11. Reactivar sede y comprobar que el recurso
	// puede volver a activarse.
	_, err = UpdateVenue(venue.ID, VenuePatch{
		IsActive: &active,
	})
	if err != nil {
		t.Fatalf("reactivar sede: %v", err)
	}

	updatedResource, err := UpdateResourceOperationalSettings(
		resourceID,
		nil,
		false,
		nil,
		&active,
	)
	if err != nil {
		t.Fatalf("reactivar recurso: %v", err)
	}
	if !updatedResource.IsActive {
		t.Fatal("el recurso deberia estar activo")
	}

	t.Log("Administracion de sedes e invariante de recursos: PASS")
}
