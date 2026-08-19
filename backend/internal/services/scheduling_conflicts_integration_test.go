package services

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

// TestSchedulingConflictMergeIntegration valida el comportamiento N-elementos
// del módulo de Programación Institucional contra PostgreSQL.
//
// Escenario:
//
//	Grupo inicial 1:
//	    A ↔ B
//
//	Grupo inicial 2:
//	    C ↔ Reserva
//
//	Posteriormente:
//	    D ↔ B
//	    D ↔ C
//
// Resultado esperado:
//
//	[A, B] + D + [C, Reserva]
//	             ↓
//	un único scheduling_conflict
//	con cinco scheduling_conflict_items.
//
// El objetivo es demostrar que Poli-REDI modela componentes conectados de
// conflictos y no una colección de pares independientes.
func TestSchedulingConflictMergeIntegration(
	t *testing.T,
) {
	if os.Getenv("POLIREDI_INTEGRATION") != "1" {
		t.Skip("integration test disabled")
	}

	database.Close()

	if err := database.Connect(); err != nil {
		t.Fatalf("connect postgres: %v", err)
	}

	defer database.Close()

	ctx := context.Background()

	suffix := time.Now().UnixNano()

	var (
		adminID       int
		unitID        int
		resourceID    int
		reservationID int
	)

	// =========================================================================
	// CLEANUP
	// =========================================================================
	//
	// El recurso es exclusivo de este test, por lo que podemos limpiar todos
	// los conflictos asociados a él sin afectar otros escenarios.
	//
	// scheduling_conflicts debe eliminarse antes que actividades y reservas
	// porque scheduling_conflict_items conserva FKs RESTRICT hacia las
	// entidades originales para proteger la trazabilidad histórica.

	defer func() {
		cleanupCtx := context.Background()

		if resourceID > 0 {
			if _, err := database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM scheduling_conflicts
				WHERE resource_id = $1
				`,
				resourceID,
			); err != nil {
				t.Logf(
					"cleanup scheduling conflicts: %v",
					err,
				)
			}
		}

		if unitID > 0 {
			if _, err := database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM institutional_activities
				WHERE unit_id = $1
				`,
				unitID,
			); err != nil {
				t.Logf(
					"cleanup institutional activities: %v",
					err,
				)
			}
		}

		if reservationID > 0 {
			if _, err := database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM reservations
				WHERE id = $1
				`,
				reservationID,
			); err != nil {
				t.Logf(
					"cleanup reservation: %v",
					err,
				)
			}
		}

		if unitID > 0 {
			if _, err := database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM institutional_unit_memberships
				WHERE unit_id = $1
				`,
				unitID,
			); err != nil {
				t.Logf(
					"cleanup institutional memberships: %v",
					err,
				)
			}

			if _, err := database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM institutional_units
				WHERE id = $1
				`,
				unitID,
			); err != nil {
				t.Logf(
					"cleanup institutional unit: %v",
					err,
				)
			}
		}

		if resourceID > 0 {
			if _, err := database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM reservation_policy_resources
				WHERE resource_id = $1
				`,
				resourceID,
			); err != nil {
				t.Logf(
					"cleanup policy resource: %v",
					err,
				)
			}

			if _, err := database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM resources
				WHERE id = $1
				`,
				resourceID,
			); err != nil {
				t.Logf(
					"cleanup resource: %v",
					err,
				)
			}
		}

		if adminID > 0 {
			if _, err := database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM users
				WHERE id = $1
				`,
				adminID,
			); err != nil {
				t.Logf(
					"cleanup admin: %v",
					err,
				)
			}
		}
	}()

	// =========================================================================
	// POLÍTICA VIGENTE
	// =========================================================================

	var (
		policyID             int
		reservableWindowDays int
		openingMinute        int
		closingMinute        int
		slotInterval         int
	)

	err := database.DB.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			reservable_window_days,
			opening_minute,
			closing_minute,
			slot_interval_minutes

		FROM reservation_policies

		WHERE is_published = true
		  AND effective_from <= CURRENT_TIMESTAMP
		  AND (
				effective_to IS NULL
				OR effective_to > CURRENT_TIMESTAMP
		  )

		ORDER BY
			effective_from DESC,
			id DESC

		LIMIT 1
		`,
	).Scan(
		&policyID,
		&reservableWindowDays,
		&openingMinute,
		&closingMinute,
		&slotInterval,
	)

	if err != nil {
		t.Fatalf(
			"load current reservation policy: %v",
			err,
		)
	}

	var durationMinutes int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT duration_minutes

		FROM reservation_policy_durations

		WHERE policy_id = $1

		ORDER BY duration_minutes ASC

		LIMIT 1
		`,
		policyID,
	).Scan(&durationMinutes)

	if err != nil {
		t.Fatalf(
			"load reservation duration: %v",
			err,
		)
	}

	// =========================================================================
	// ADMIN TEMPORAL
	// =========================================================================

	adminEmail := fmt.Sprintf(
		"mvp2.conflicts.admin.%d@test.local",
		suffix,
	)

	err = database.DB.QueryRowContext(
		ctx,
		`
		INSERT INTO users (
			email,
			full_name,
			rut,
			is_admin,
			is_blocked
		)
		VALUES (
			$1,
			'Integration Scheduling Conflict Admin',
			NULL,
			true,
			false
		)
		RETURNING id
		`,
		adminEmail,
	).Scan(&adminID)

	if err != nil {
		t.Fatalf(
			"create integration admin: %v",
			err,
		)
	}

	admin := models.LocalAuthUser{
		ID:        adminID,
		Email:     adminEmail,
		FullName:  "Integration Scheduling Conflict Admin",
		IsAdmin:   true,
		IsBlocked: false,
	}

	// =========================================================================
	// UNIDAD INSTITUCIONAL
	// =========================================================================

	unit, err := CreateInstitutionalUnit(
		admin,
		models.CreateInstitutionalUnitRequest{
			Name: fmt.Sprintf(
				"Scheduling Conflict Integration %d",
				suffix,
			),
			Code: fmt.Sprintf(
				"SCI-%d",
				suffix,
			),
			UnitType: models.InstitutionalUnitTypeAcademicProgram,
		},
	)

	if err != nil {
		t.Fatalf(
			"create institutional unit: %v",
			err,
		)
	}

	unitID = unit.ID

	// =========================================================================
	// RECURSO TEMPORAL
	// =========================================================================
	//
	// Usamos un recurso exclusivo para que ninguna reserva o actividad de otro
	// integration test pueda convertirse accidentalmente en un nodo adicional.

	var venueID int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT id
		FROM venues
		WHERE is_active = true
		ORDER BY id
		LIMIT 1
		`,
	).Scan(&venueID)

	if err != nil {
		t.Fatalf(
			"load active venue: %v",
			err,
		)
	}

	resourceName := fmt.Sprintf(
		"Scheduling Conflict Resource %d",
		suffix,
	)

	err = database.DB.QueryRowContext(
		ctx,
		`
		INSERT INTO resources (
			venue_id,
			name,
			type,
			reservation_mode,
			capacity,
			is_active
		)
		VALUES (
			$1,
			$2,
			'INTEGRATION_TEST',
			'RESERVABLE',
			50,
			true
		)
		RETURNING id
		`,
		venueID,
		resourceName,
	).Scan(&resourceID)

	if err != nil {
		t.Fatalf(
			"create integration resource: %v",
			err,
		)
	}

	// La reserva debe pertenecer al scope de la política vigente.
	_, err = database.DB.ExecContext(
		ctx,
		`
		INSERT INTO reservation_policy_resources (
			policy_id,
			resource_id
		)
		VALUES (
			$1,
			$2
		)
		`,
		policyID,
		resourceID,
	)

	if err != nil {
		t.Fatalf(
			"attach resource to policy: %v",
			err,
		)
	}

	// =========================================================================
	// FECHA Y HORA BASE
	// =========================================================================

	// Elegimos mañana cuando la política lo permite para evitar usar una fecha
	// pasada en relación con la hora actual. Si la ventana fuese de un solo
	// día, utilizamos hoy.
	reservationDate := businessclock.Now()

	if reservableWindowDays > 1 {
		reservationDate =
			reservationDate.AddDate(0, 0, 1)
	}

	reservationDateValue :=
		reservationDate.Format("2006-01-02")

	// La reserva debe comenzar:
	//
	//   - dentro del horario de la política;
	//   - alineada al slot_interval;
	//   - con espacio suficiente para terminar antes del cierre.
	//
	// Además dejamos al menos 60 minutos antes de R para construir A y B.

	minute := openingMinute

	if minute < 60 {
		minute = 60
	}

	reservationStartMinute :=
		((minute + slotInterval - 1) / slotInterval) *
			slotInterval

	if reservationStartMinute+durationMinutes >
		closingMinute {
		t.Fatalf(
			"policy has no suitable test slot: opening=%d closing=%d duration=%d slot=%d",
			openingMinute,
			closingMinute,
			durationMinutes,
			slotInterval,
		)
	}

	reservationStart := time.Date(
		reservationDate.Year(),
		reservationDate.Month(),
		reservationDate.Day(),
		reservationStartMinute/60,
		reservationStartMinute%60,
		0,
		0,
		businessclock.Location(),
	)

	reservationEnd :=
		reservationStart.Add(
			time.Duration(durationMinutes) *
				time.Minute,
		)

	// =========================================================================
	// RESERVA R
	// =========================================================================
	//
	// La reserva existe ANTES de crear las actividades institucionales.
	//
	// No usamos AddReservationWithPolicy porque aquí no estamos validando el
	// flujo de creación de reservas: ese módulo ya posee sus integration tests.
	//
	// Sí dejamos pasar el INSERT por las invariantes reales de PostgreSQL.

	err = database.DB.QueryRowContext(
		ctx,
		`
		INSERT INTO reservations (
			policy_id,
			user_id,
			resource_id,
			start_time,
			end_time,
			duration_minutes,
			status
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			'CONFIRMED'
		)
		RETURNING id
		`,
		policyID,
		adminID,
		resourceID,
		reservationStart,
		reservationEnd,
		durationMinutes,
	).Scan(&reservationID)

	if err != nil {
		t.Fatalf(
			"create integration reservation: %v",
			err,
		)
	}

	// =========================================================================
	// HELPER PARA ACTIVIDADES SINGLE
	// =========================================================================

	createActivity := func(
		title string,
		startMinute int,
		endMinute int,
	) models.InstitutionalActivity {
		t.Helper()

		activity, err := CreateInstitutionalActivity(
			admin,
			models.CreateInstitutionalActivityRequest{
				UnitID:       unitID,
				ResourceID:   resourceID,
				ActivityType: models.InstitutionalActivityTypeAcademicClass,
				Title: fmt.Sprintf(
					"%s %d",
					title,
					suffix,
				),

				RequiresEnrollment: false,

				Schedules: []models.CreateInstitutionalScheduleRequest{
					{
						ScheduleType: models.InstitutionalScheduleTypeSingle,
						SpecificDate: &reservationDateValue,
						StartTime: conflictIntegrationClock(
							startMinute,
						),
						EndTime: conflictIntegrationClock(
							endMinute,
						),
					},
				},
			},
		)

		if err != nil {
			t.Fatalf(
				"create activity %s: %v",
				title,
				err,
			)
		}

		return activity
	}

	// =========================================================================
	// GRUPO 1: A ↔ B
	// =========================================================================
	//
	// Respecto del inicio de R:
	//
	// A = T-60 .. T-30
	// B = T-45 .. T-15
	//
	// A ↔ B
	// B NO toca R.

	activityA := createActivity(
		"Activity A",
		reservationStartMinute-60,
		reservationStartMinute-30,
	)

	// A por sí sola no debe crear conflicto.
	assertPendingConflictShape(
		t,
		ctx,
		resourceID,
		0,
		0,
	)

	activityB := createActivity(
		"Activity B",
		reservationStartMinute-45,
		reservationStartMinute-15,
	)

	// Ahora existe:
	//
	//   [A,B]
	assertPendingConflictShape(
		t,
		ctx,
		resourceID,
		1,
		2,
	)

	// =========================================================================
	// GRUPO 2: C ↔ R
	// =========================================================================
	//
	// C = T-10 .. T+20
	// R = T    .. T+duration
	//
	// C ↔ R
	//
	// C todavía NO toca B porque:
	//
	// B termina T-15
	// C empieza T-10

	activityC := createActivity(
		"Activity C",
		reservationStartMinute-10,
		reservationStartMinute+20,
	)

	// Deben existir dos componentes independientes:
	//
	//   [A,B]
	//   [C,R]
	assertPendingConflictShape(
		t,
		ctx,
		resourceID,
		2,
		4,
	)

	// =========================================================================
	// PUENTE D
	// =========================================================================
	//
	// D = T-20 .. T
	//
	// D ↔ B
	// D ↔ C
	//
	// D y R solamente se tocan exactamente en T.
	//
	// Como usamos intervalos [start,end), D NO se solapa directamente con R.
	//
	// El grafo queda:
	//
	// A ↔ B ↔ D ↔ C ↔ R
	//
	// Los dos conflictos existentes deben fusionarse en uno solo.

	activityD := createActivity(
		"Activity D",
		reservationStartMinute-20,
		reservationStartMinute,
	)

	assertPendingConflictShape(
		t,
		ctx,
		resourceID,
		1,
		5,
	)

	// =========================================================================
	// VALIDACIÓN DEL CONFLICTO FINAL
	// =========================================================================

	var conflictID int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT id

		FROM scheduling_conflicts

		WHERE resource_id = $1
		  AND status = 'PENDING'

		ORDER BY id

		LIMIT 1
		`,
		resourceID,
	).Scan(&conflictID)

	if err != nil {
		t.Fatalf(
			"load final conflict: %v",
			err,
		)
	}

	// Deben existir cuatro ocurrencias institucionales.
	var institutionalItemCount int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT COUNT(*)

		FROM scheduling_conflict_items

		WHERE conflict_id = $1
		  AND institutional_activity_id IS NOT NULL
		  AND reservation_id IS NULL
		`,
		conflictID,
	).Scan(&institutionalItemCount)

	if err != nil {
		t.Fatalf(
			"count institutional conflict items: %v",
			err,
		)
	}

	if institutionalItemCount != 4 {
		t.Fatalf(
			"expected 4 institutional items, got %d",
			institutionalItemCount,
		)
	}

	// Y exactamente una reserva.
	var reservationItemCount int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT COUNT(*)

		FROM scheduling_conflict_items

		WHERE conflict_id = $1
		  AND reservation_id = $2
		`,
		conflictID,
		reservationID,
	).Scan(&reservationItemCount)

	if err != nil {
		t.Fatalf(
			"count reservation conflict item: %v",
			err,
		)
	}

	if reservationItemCount != 1 {
		t.Fatalf(
			"expected reservation exactly once, got %d",
			reservationItemCount,
		)
	}

	// Confirmamos que A/B/C/D son exactamente las actividades almacenadas.
	var matchedActivityCount int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT COUNT(*)

		FROM scheduling_conflict_items

		WHERE conflict_id = $1
		  AND institutional_activity_id = ANY($2::integer[])
		`,
		conflictID,
		fmt.Sprintf(
			"{%d,%d,%d,%d}",
			activityA.ID,
			activityB.ID,
			activityC.ID,
			activityD.ID,
		),
	).Scan(&matchedActivityCount)

	if err != nil {
		t.Fatalf(
			"validate institutional activity IDs: %v",
			err,
		)
	}

	if matchedActivityCount != 4 {
		t.Fatalf(
			"expected A/B/C/D in final conflict, got %d",
			matchedActivityCount,
		)
	}

	// Ningún item ha sido resuelto todavía.
	var pendingItemCount int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT COUNT(*)

		FROM scheduling_conflict_items

		WHERE conflict_id = $1
		  AND resolution = 'PENDING'
		`,
		conflictID,
	).Scan(&pendingItemCount)

	if err != nil {
		t.Fatalf(
			"count pending items: %v",
			err,
		)
	}

	if pendingItemCount != 5 {
		t.Fatalf(
			"expected 5 PENDING items, got %d",
			pendingItemCount,
		)
	}

	// =========================================================================
	// SNAPSHOT / INTERVALO PROTEGIDO
	// =========================================================================

	var (
		protectedStart time.Time
		protectedEnd   time.Time
	)

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT
			MIN(occurrence_start),
			MAX(occurrence_end)

		FROM scheduling_conflict_items

		WHERE conflict_id = $1
		`,
		conflictID,
	).Scan(
		&protectedStart,
		&protectedEnd,
	)

	if err != nil {
		t.Fatalf(
			"load protected interval snapshots: %v",
			err,
		)
	}

	expectedStart :=
		reservationStart.Add(-60 * time.Minute)

	expectedEnd :=
		reservationEnd

	cEnd :=
		reservationStart.Add(20 * time.Minute)

	if cEnd.After(expectedEnd) {
		expectedEnd = cEnd
	}

	if !protectedStart.Equal(expectedStart) {
		t.Fatalf(
			"expected protected start %v, got %v",
			expectedStart,
			protectedStart,
		)
	}

	if !protectedEnd.Equal(expectedEnd) {
		t.Fatalf(
			"expected protected end %v, got %v",
			expectedEnd,
			protectedEnd,
		)
	}

	t.Logf(
		"PASS: conflict=%d resource=%d items=5 [A=%d B=%d D=%d C=%d R=%d] protected=%s..%s",
		conflictID,
		resourceID,
		activityA.ID,
		activityB.ID,
		activityD.ID,
		activityC.ID,
		reservationID,
		protectedStart.In(
			businessclock.Location(),
		).Format("15:04"),
		protectedEnd.In(
			businessclock.Location(),
		).Format("15:04"),
	)
}

// ============================================================================
// HELPERS
// ============================================================================

func assertPendingConflictShape(
	t *testing.T,
	ctx context.Context,
	resourceID int,
	expectedConflicts int,
	expectedItems int,
) {
	t.Helper()

	var conflictCount int

	err := database.DB.QueryRowContext(
		ctx,
		`
		SELECT COUNT(*)

		FROM scheduling_conflicts

		WHERE resource_id = $1
		  AND status = 'PENDING'
		`,
		resourceID,
	).Scan(&conflictCount)

	if err != nil {
		t.Fatalf(
			"count pending conflicts: %v",
			err,
		)
	}

	if conflictCount != expectedConflicts {
		t.Fatalf(
			"expected %d pending conflicts, got %d",
			expectedConflicts,
			conflictCount,
		)
	}

	var itemCount int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT COUNT(*)

		FROM scheduling_conflict_items item

		INNER JOIN scheduling_conflicts conflict
			ON conflict.id = item.conflict_id

		WHERE conflict.resource_id = $1
		  AND conflict.status = 'PENDING'
		`,
		resourceID,
	).Scan(&itemCount)

	if err != nil {
		t.Fatalf(
			"count conflict items: %v",
			err,
		)
	}

	if itemCount != expectedItems {
		t.Fatalf(
			"expected %d conflict items, got %d",
			expectedItems,
			itemCount,
		)
	}
}

func conflictIntegrationClock(
	minute int,
) string {
	return fmt.Sprintf(
		"%02d:%02d",
		minute/60,
		minute%60,
	)
}
