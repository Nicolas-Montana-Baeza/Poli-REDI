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

func TestResourceConfigurationPostgresIntegration(t *testing.T) {
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

	venue, err := CreateVenue(models.Venue{
		Name:    "Sede EV012 Edicion",
		Country: "Chile",
	})
	if err != nil {
		t.Fatal(err)
	}

	initialCapacity := 20

	first, err := CreateResource(CreateResourceInput{
		VenueID:         venue.ID,
		Name:            "Cancha EV012 Original",
		Type:            "Cancha",
		ReservationMode: "RESERVABLE",
		Capacity:        &initialCapacity,
		IsActive:        true,
	})
	if err != nil {
		t.Fatal(err)
	}

	second, err := CreateResource(CreateResourceInput{
		VenueID:         venue.ID,
		Name:            "Cancha EV012 Existente",
		Type:            "Cancha",
		ReservationMode: "RESERVABLE",
		Capacity:        &initialCapacity,
		IsActive:        true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Actualizacion combinada.
	newName := "Cancha EV012 Renovada"
	newType := "Multideporte"
	newMode := "OPEN_USE"
	newCapacity := 30

	updated, err := UpdateResourceConfiguration(
		first.ID,
		ResourceConfigurationPatch{
			Name:            &newName,
			Type:            &newType,
			ReservationMode: &newMode,
			CapacityPresent: true,
			Capacity:        &newCapacity,
		},
	)
	if err != nil {
		t.Fatalf("actualizar configuracion: %v", err)
	}

	if updated.Name != newName ||
		updated.Type != newType ||
		updated.ReservationMode != newMode ||
		updated.Capacity == nil ||
		*updated.Capacity != newCapacity ||
		updated.VenueID != venue.ID {
		t.Fatalf("respuesta incorrecta: %+v", updated)
	}

	// Un nombre duplicado debe rechazar TODA la operacion.
	rejectedType := "Tipo no persistente"
	inactive := false

	_, err = UpdateResourceConfiguration(
		first.ID,
		ResourceConfigurationPatch{
			Name:     &second.Name,
			Type:     &rejectedType,
			IsActive: &inactive,
		},
	)
	if !errors.Is(err, ErrResourceDuplicate) {
		t.Fatalf("esperaba rechazo por duplicado, obtuve %v", err)
	}

	// Verificar directamente que PostgreSQL no guardo
	// modificaciones parciales.
	var (
		storedName   string
		storedType   string
		storedMode   string
		storedActive bool
		storedVenue  int
		storedCap    int
	)

	err = database.DB.QueryRowContext(ctx, `
		SELECT
			name,
			type,
			reservation_mode,
			is_active,
			venue_id,
			capacity
		FROM resources
		WHERE id = $1
	`, first.ID).Scan(
		&storedName,
		&storedType,
		&storedMode,
		&storedActive,
		&storedVenue,
		&storedCap,
	)
	if err != nil {
		t.Fatal(err)
	}

	if storedName != newName ||
		storedType != newType ||
		storedMode != newMode ||
		!storedActive ||
		storedVenue != venue.ID ||
		storedCap != newCapacity {
		t.Fatalf(
			"actualizacion rechazada dejo cambios parciales: "+
				"name=%q type=%q mode=%q active=%v venue=%d capacity=%d",
			storedName, storedType, storedMode,
			storedActive, storedVenue, storedCap,
		)
	}

	// Identificador inexistente.
	_, err = UpdateResourceConfiguration(
		99999999,
		ResourceConfigurationPatch{Name: &newName},
	)
	if !errors.Is(err, ErrResourceNotFound) {
		t.Fatalf("esperaba recurso inexistente: %v", err)
	}

	t.Log("Edicion integral parcial y atomicidad: PASS")
}
