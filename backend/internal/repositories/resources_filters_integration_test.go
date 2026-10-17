//go:build integration

package repositories

import (
	"os"
	"testing"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func TestResourceFiltersPostgresIntegration(t *testing.T) {
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

	venueA, err := CreateVenue(models.Venue{
		Name:    "Sede EV012 Filtros A",
		Country: "Chile",
	})
	if err != nil {
		t.Fatal(err)
	}

	venueB, err := CreateVenue(models.Venue{
		Name:    "Sede EV012 Filtros B",
		Country: "Chile",
	})
	if err != nil {
		t.Fatal(err)
	}

	capacity := 20

	create := func(venueID int, name, resourceType string, active bool) int {
		t.Helper()

		resource, err := CreateResource(CreateResourceInput{
			VenueID:         venueID,
			Name:            name,
			Type:            resourceType,
			ReservationMode: "RESERVABLE",
			Capacity:        &capacity,
			IsActive:        active,
		})
		if err != nil {
			t.Fatalf("crear %s: %v", name, err)
		}

		return resource.ID
	}

	activeA := create(venueA.ID, "Cancha Filtro A1", "Cancha", true)
	inactiveA := create(venueA.ID, "Cancha Filtro A2", "Cancha", false)
	activeB := create(venueB.ID, "Sala Filtro B1", "Sala", true)

	contains := func(resources []models.Resource, id int) bool {
		for _, resource := range resources {
			if resource.ID == id {
				return true
			}
		}
		return false
	}

	expectIDs := func(resources []models.Resource, expected ...int) {
		t.Helper()

		if len(resources) != len(expected) {
			t.Fatalf(
				"esperaba %d recursos, obtuvo %d: %+v",
				len(expected), len(resources), resources,
			)
		}

		for _, id := range expected {
			if !contains(resources, id) {
				t.Fatalf("recurso %d no encontrado: %+v", id, resources)
			}
		}
	}

	// Sin filtros: conservar el comportamiento anterior.
	all, err := GetAllResources()
	if err != nil {
		t.Fatal(err)
	}

	unfiltered, err := GetResourcesFiltered(ResourceFilters{})
	if err != nil {
		t.Fatal(err)
	}

	if len(all) != len(unfiltered) {
		t.Fatal("el contrato sin filtros cambió")
	}

	for i := range all {
		if all[i].ID != unfiltered[i].ID {
			t.Fatal("el orden o contenido sin filtros cambió")
		}
	}

	for _, id := range []int{activeA, inactiveA, activeB} {
		if !contains(all, id) {
			t.Fatalf("inventario completo no contiene %d", id)
		}
	}

	// Filtro por sede.
	byVenue, err := GetResourcesFiltered(ResourceFilters{
		VenueID: &venueA.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	expectIDs(byVenue, activeA, inactiveA)

	// Filtro por tipo.
	cancha := "Cancha"
	byType, err := GetResourcesFiltered(ResourceFilters{
		Type: &cancha,
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, resource := range byType {
		if resource.Type != cancha {
			t.Fatalf("tipo incorrecto: %+v", resource)
		}
	}
	if !contains(byType, activeA) || !contains(byType, inactiveA) {
		t.Fatal("el filtro por tipo omitió recursos esperados")
	}

	// Filtro por estado.
	inactive := false
	byStatus, err := GetResourcesFiltered(ResourceFilters{
		IsActive: &inactive,
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, resource := range byStatus {
		if resource.IsActive {
			t.Fatalf("recurso activo en filtro inactivo: %+v", resource)
		}
	}
	if !contains(byStatus, inactiveA) {
		t.Fatal("no se encontró el recurso inactivo")
	}

	// Los tres filtros combinados.
	active := true
	combined, err := GetResourcesFiltered(ResourceFilters{
		VenueID:  &venueA.ID,
		Type:     &cancha,
		IsActive: &active,
	})
	if err != nil {
		t.Fatal(err)
	}
	expectIDs(combined, activeA)

	sala := "Sala"
	combinedB, err := GetResourcesFiltered(ResourceFilters{
		VenueID:  &venueB.ID,
		Type:     &sala,
		IsActive: &active,
	})
	if err != nil {
		t.Fatal(err)
	}
	expectIDs(combinedB, activeB)

	// Sin coincidencias: devolver lista vacía sin error.
	empty, err := GetResourcesFiltered(ResourceFilters{
		VenueID:  &venueA.ID,
		Type:     &sala,
		IsActive: &active,
	})
	if err != nil {
		t.Fatal(err)
	}
	expectIDs(empty)

	t.Log("Filtros individuales, combinados y compatibilidad: PASS")
}
