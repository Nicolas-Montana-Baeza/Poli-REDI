//go:build integration

package services

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func TestPublishPolicyGroupResourcesPostgresIntegration(t *testing.T) {
	if os.Getenv("POLIREDI_3C_ISOLATED") != "1" {
		t.Skip("requiere entorno PostgreSQL efimero 3C")
	}
	if os.Getenv("DATABASE_URL") == "" {
		t.Fatal("DATABASE_URL es obligatorio")
	}

	if err := database.Connect(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(database.Close)

	ctx := context.Background()

	var adminID int
	if err := database.DB.QueryRowContext(
		ctx,
		`SELECT id FROM users
		 WHERE lower(email) = 'admin@poliredi.local'`,
	).Scan(&adminID); err != nil {
		t.Fatalf("admin de prueba: %v", err)
	}

	baseline, err := GetCurrentReservationPolicy()
	if err != nil {
		t.Fatal(err)
	}
	if len(baseline.GroupResources) < 3 {
		t.Fatalf(
			"esperaba 3 recursos grupales iniciales: %+v",
			baseline.GroupResources,
		)
	}

	makeRequest := func() models.PublishReservationPolicyRequest {
		active, activeErr := GetCurrentReservationPolicy()
		if activeErr != nil {
			t.Fatalf("obtener version vigente: %v", activeErr)
		}

		return models.PublishReservationPolicyRequest{
			ExpectedPolicyID:            active.ID,
			ReservableWindowDays:        baseline.ReservableWindowDays,
			RequestFrequencyDays:        baseline.RequestFrequencyDays,
			ConfirmationDeadlineMinutes: baseline.ConfirmationDeadlineMinutes,
			MinimumParticipants:         baseline.MinimumParticipants,
			OpeningMinute:               baseline.OpeningMinute,
			ClosingMinute:               baseline.ClosingMinute,
			SlotIntervalMinutes:         baseline.SlotIntervalMinutes,
			AllowedDurations: append(
				[]int{}, baseline.AllowedDurations...,
			),
			ResourceIDs: append(
				[]int{}, baseline.ResourceIDs...,
			),
		}
	}

	countPolicies := func() int {
		t.Helper()
		var count int
		if err := database.DB.QueryRowContext(
			ctx, `SELECT COUNT(*) FROM reservation_policies`,
		).Scan(&count); err != nil {
			t.Fatal(err)
		}
		return count
	}

	assertCurrent := func(expectedID int) {
		t.Helper()
		current, err := GetCurrentReservationPolicy()
		if err != nil {
			t.Fatal(err)
		}
		if current.ID != expectedID {
			t.Fatalf(
				"politica vigente %d, esperaba %d",
				current.ID, expectedID,
			)
		}
	}

	assertRejected := func(
		request models.PublishReservationPolicyRequest,
		key string,
	) {
		t.Helper()

		before := countPolicies()
		current, err := GetCurrentReservationPolicy()
		if err != nil {
			t.Fatal(err)
		}

		_, _, err = PublishReservationPolicy(
			request, adminID, key,
		)

		var validation ReservationPolicyValidationError
		if !errors.As(err, &validation) {
			t.Fatalf(
				"esperaba rechazo de validacion, obtuve %v",
				err,
			)
		}

		if countPolicies() != before {
			t.Fatal("una publicacion rechazada creo una version")
		}
		assertCurrent(current.ID)
	}

	groupID := baseline.GroupResources[0].ResourceID

	var capacity int
	if err := database.DB.QueryRowContext(
		ctx,
		`SELECT capacity FROM resources WHERE id = $1`,
		groupID,
	).Scan(&capacity); err != nil {
		t.Fatal(err)
	}

	// 1. Rechazar minimo mayor que la capacidad.
	invalidCapacity := makeRequest()
	badRules := []models.PolicyGroupResource{
		{
			ResourceID:          groupID,
			MinimumParticipants: capacity + 1,
		},
	}
	invalidCapacity.GroupResources = &badRules

	assertRejected(
		invalidCapacity,
		"mvp3-3c-test-invalid-capacity",
	)

	// 2. Rechazar clasificacion grupal de un recurso OPEN_USE.
	var openUseID int
	if err := database.DB.QueryRowContext(
		ctx,
		`SELECT id FROM resources
		 WHERE reservation_mode = 'OPEN_USE'
		   AND is_active = true
		 ORDER BY id LIMIT 1`,
	).Scan(&openUseID); err != nil {
		t.Fatal(err)
	}

	invalidMode := makeRequest()
	invalidMode.ResourceIDs = append(
		invalidMode.ResourceIDs, openUseID,
	)

	badModeRules := []models.PolicyGroupResource{
		{
			ResourceID:          openUseID,
			MinimumParticipants: 10,
		},
	}
	invalidMode.GroupResources = &badModeRules

	assertRejected(
		invalidMode,
		"mvp3-3c-test-invalid-mode",
	)

	// 3. Publicacion explicita de minimos grupales.
	explicit := makeRequest()

	explicitRules := append(
		[]models.PolicyGroupResource{},
		baseline.GroupResources...,
	)

	if capacity > explicitRules[0].MinimumParticipants {
		explicitRules[0].MinimumParticipants++
	}
	explicit.GroupResources = &explicitRules

	const explicitKey = "mvp3-3c-test-explicit"

	created, replayed, err := PublishReservationPolicy(
		explicit, adminID, explicitKey,
	)
	if err != nil {
		t.Fatalf("publicacion explicita: %v", err)
	}
	if replayed {
		t.Fatal("primera publicacion no puede ser replay")
	}

	if !reflect.DeepEqual(
		created.GroupResources, explicitRules,
	) {
		t.Fatalf(
			"reglas publicadas incorrectas: %+v",
			created.GroupResources,
		)
	}
	assertCurrent(created.ID)

	// Rechazar una publicación basada en una versión antigua.
	stale := makeRequest()
	stale.ExpectedPolicyID = baseline.ID

	countBeforeStale := countPolicies()

	_, _, staleErr := PublishReservationPolicy(
		stale, adminID, "mvp3-3c-test-stale",
	)

	if !errors.Is(staleErr, ErrReservationPolicyStale) {
		t.Fatalf("esperaba conflicto de version: %v", staleErr)
	}

	if countPolicies() != countBeforeStale {
		t.Fatal("una publicacion obsoleta creo otra version")
	}

	assertCurrent(created.ID)

	// 4. Idempotencia: misma clave y payload.
	countBeforeReplay := countPolicies()

	repeated, replayed, err := PublishReservationPolicy(
		explicit, adminID, explicitKey,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !replayed || repeated.ID != created.ID {
		t.Fatal("replay no recupero la version original")
	}
	if countPolicies() != countBeforeReplay {
		t.Fatal("replay creo una version adicional")
	}

	// 5. Idempotencia: misma clave y payload diferente.
	conflicting := makeRequest()
	conflicting.ReservableWindowDays++

	_, _, err = PublishReservationPolicy(
		conflicting, adminID, explicitKey,
	)
	if !errors.Is(err, ErrReservationPolicyConflict) {
		t.Fatalf("esperaba conflicto de idempotencia: %v", err)
	}
	assertCurrent(created.ID)

	// 6. Campo omitido: herencia filtrada al nuevo alcance.
	inherited := makeRequest()
	inherited.ResourceIDs = []int{groupID}
	inherited.GroupResources = nil

	inheritedPolicy, replayed, err := PublishReservationPolicy(
		inherited,
		adminID,
		"mvp3-3c-test-inherited",
	)
	if err != nil || replayed {
		t.Fatalf(
			"publicacion heredada: replay=%v error=%v",
			replayed, err,
		)
	}

	expectedInherited := []models.PolicyGroupResource{
		explicitRules[0],
	}

	if !reflect.DeepEqual(
		inheritedPolicy.GroupResources,
		expectedInherited,
	) {
		t.Fatalf(
			"herencia incorrecta: %+v",
			inheritedPolicy.GroupResources,
		)
	}
	assertCurrent(inheritedPolicy.ID)

	// 7. Lista vacia explicita: ninguna regla grupal.
	empty := makeRequest()
	emptyRules := []models.PolicyGroupResource{}
	empty.GroupResources = &emptyRules

	emptyPolicy, replayed, err := PublishReservationPolicy(
		empty,
		adminID,
		"mvp3-3c-test-empty",
	)
	if err != nil || replayed {
		t.Fatalf(
			"publicacion vacia: replay=%v error=%v",
			replayed, err,
		)
	}
	if len(emptyPolicy.GroupResources) != 0 {
		t.Fatalf(
			"la version vacia conserva reglas: %+v",
			emptyPolicy.GroupResources,
		)
	}
	assertCurrent(emptyPolicy.ID)

	// 8. Historial: la version original conserva sus reglas.
	history, err := GetReservationPolicyHistory()
	if err != nil {
		t.Fatal(err)
	}

	foundBaseline := false

	for _, policy := range history {
		if policy.ID != baseline.ID {
			continue
		}

		foundBaseline = true

		if policy.EffectiveTo == nil {
			t.Fatal("la politica anterior sigue abierta")
		}

		if !reflect.DeepEqual(
			policy.GroupResources,
			baseline.GroupResources,
		) {
			t.Fatal("las reglas historicas fueron modificadas")
		}
	}

	if !foundBaseline {
		t.Fatal("la politica original desaparecio del historial")
	}

	// 9. Replay de una version historica no cambia la vigente.
	_, replayed, err = PublishReservationPolicy(
		explicit, adminID, explicitKey,
	)
	if err != nil || !replayed {
		t.Fatalf(
			"replay historico: replay=%v error=%v",
			replayed, err,
		)
	}
	assertCurrent(emptyPolicy.ID)
}
