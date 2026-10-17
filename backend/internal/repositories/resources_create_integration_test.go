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

func TestResourceCreationPostgresIntegration(t *testing.T) {
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

	venueA, err := CreateVenue(models.Venue{
		Name: "Sede EV012 Recursos A", Country: "Chile",
	})
	if err != nil {
		t.Fatal(err)
	}

	venueB, err := CreateVenue(models.Venue{
		Name: "Sede EV012 Recursos B", Country: "Chile",
	})
	if err != nil {
		t.Fatal(err)
	}

	capacity := 20
	input := CreateResourceInput{
		VenueID:         venueA.ID,
		Name:            "Cancha EV012",
		Type:            "Cancha",
		ReservationMode: "RESERVABLE",
		ImageURL:        "/images/ev012.png",
		Capacity:        &capacity,
		IsActive:        true,
	}

	// Crear y consultar.
	created, err := CreateResource(input)
	if err != nil {
		t.Fatalf("crear recurso: %v", err)
	}
	if created.ID <= 0 || created.VenueID != venueA.ID ||
		!created.IsActive || created.Capacity == nil ||
		*created.Capacity != capacity ||
		created.ImageURL != input.ImageURL {
		t.Fatalf("recurso incorrecto: %+v", created)
	}

	var storedVenueID int
	err = database.DB.QueryRowContext(ctx, `
		SELECT venue_id FROM resources WHERE id = $1
	`, created.ID).Scan(&storedVenueID)
	if err != nil || storedVenueID != venueA.ID {
		t.Fatalf("asociacion de sede incorrecta: %v", err)
	}

	// No incorporar automaticamente a politicas existentes.
	var policyLinks, groupLinks int
	err = database.DB.QueryRowContext(ctx, `
		SELECT
		  (SELECT COUNT(*) FROM reservation_policy_resources
		   WHERE resource_id = $1),
		  (SELECT COUNT(*) FROM reservation_policy_group_resources
		   WHERE resource_id = $1)
	`, created.ID).Scan(&policyLinks, &groupLinks)
	if err != nil {
		t.Fatal(err)
	}
	if policyLinks != 0 || groupLinks != 0 {
		t.Fatal("el recurso fue agregado a una politica automaticamente")
	}

	// Duplicado dentro de la misma sede.
	_, err = CreateResource(input)
	if !errors.Is(err, ErrResourceDuplicate) {
		t.Fatalf("esperaba recurso duplicado: %v", err)
	}

	// El mismo nombre en otra sede es valido.
	input.VenueID = venueB.ID
	input.IsActive = false

	second, err := CreateResource(input)
	if err != nil || second.VenueID != venueB.ID {
		t.Fatalf("mismo nombre en otra sede: %v", err)
	}

	// Rechazar sede inexistente.
	input.VenueID = 99999999
	_, err = CreateResource(input)
	if !errors.Is(err, ErrVenueNotFound) {
		t.Fatalf("esperaba sede inexistente: %v", err)
	}

	// Desactivar sede B: solo tiene recursos inactivos.
	inactive := false
	_, err = UpdateVenue(venueB.ID, VenuePatch{
		IsActive: &inactive,
	})
	if err != nil {
		t.Fatal(err)
	}

	// No crear recursos activos en sede inactiva.
	input.VenueID = venueB.ID
	input.Name = "Nuevo activo EV012"
	input.IsActive = true

	_, err = CreateResource(input)
	if !errors.Is(err, ErrResourceVenueInactive) {
		t.Fatalf("esperaba rechazo por sede inactiva: %v", err)
	}

	var rejectedCount int
	err = database.DB.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM resources
		WHERE venue_id = $1 AND name = $2
	`, venueB.ID, input.Name).Scan(&rejectedCount)
	if err != nil || rejectedCount != 0 {
		t.Fatalf("creacion rechazada dejo datos: %v", err)
	}

	// Un recurso inactivo si puede registrarse.
	input.IsActive = false
	createdInactive, err := CreateResource(input)
	if err != nil || createdInactive.IsActive {
		t.Fatalf("crear recurso inactivo: %v", err)
	}

	t.Log("Alta, duplicados, sedes y politicas: PASS")
}
