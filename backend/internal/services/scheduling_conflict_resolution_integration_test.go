package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
)

// ============================================================================
// RESOLUCIÓN ADMINISTRATIVA DE CONFLICTOS
// ============================================================================
//
// Este integration test valida el ciclo completo:
//
//   conflicto N
//       ↓
//   decisiones parciales
//       ↓
//   side effects reales
//       ↓
//   cierre RESOLVED
//
// También valida:
//
//   KEEP + KEEP solapados   -> rollback
//   CANCEL actividad        -> excepción CANCEL
//   CANCEL reserva          -> reservation.status = CANCELLED
//   RESCHEDULE actividad    -> excepción RESCHEDULE
//   nuevo choque al mover   -> nuevo scheduling_conflict
//   ALLOW + ALLOW           -> coexistencia autorizada
//
// El detector inicial posee sus propios integration tests. Aquí creamos
// deliberadamente los grupos de conflicto para aislar y probar el resolver.
// ============================================================================

func TestSchedulingConflictResolutionIntegration(
	t *testing.T,
) {
	if os.Getenv("POLIREDI_INTEGRATION") != "1" {
		t.Skip("integration test disabled")
	}

	database.Close()

	if err := database.Connect(); err != nil {
		t.Fatalf(
			"connect postgres: %v",
			err,
		)
	}

	defer database.Close()

	ctx := context.Background()

	suffix := time.Now().UnixNano()

	var (
		adminID    int
		resourceID int
		unitID     int
	)

	// ========================================================================
	// CLEANUP
	// ========================================================================

	defer func() {
		cleanupCtx := context.Background()

		if resourceID > 0 {
			// Los items poseen FK RESTRICT hacia actividades y reservas.
			// Los conflictos deben desaparecer primero.
			_, _ = database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM scheduling_conflicts
				WHERE resource_id = $1
				`,
				resourceID,
			)
		}

		if unitID > 0 {
			// schedule_exceptions desaparece mediante el cascade del schedule.
			_, _ = database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM institutional_activities
				WHERE unit_id = $1
				`,
				unitID,
			)

			_, _ = database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM institutional_unit_memberships
				WHERE unit_id = $1
				`,
				unitID,
			)

			_, _ = database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM institutional_units
				WHERE id = $1
				`,
				unitID,
			)
		}

		if resourceID > 0 {
			_, _ = database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM reservations
				WHERE resource_id = $1
				`,
				resourceID,
			)

			_, _ = database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM reservation_policy_resources
				WHERE resource_id = $1
				`,
				resourceID,
			)

			_, _ = database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM resources
				WHERE id = $1
				`,
				resourceID,
			)
		}

		if adminID > 0 {
			_, _ = database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM users
				WHERE id = $1
				`,
				adminID,
			)
		}
	}()

	// ========================================================================
	// POLÍTICA DE RESERVAS
	// ========================================================================

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
			"load reservation policy: %v",
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

		ORDER BY duration_minutes

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

	// ========================================================================
	// ADMIN
	// ========================================================================

	adminEmail := fmt.Sprintf(
		"resolver.integration.%d@test.local",
		suffix,
	)

	err = database.DB.QueryRowContext(
		ctx,
		`
		INSERT INTO users (
			email,
			full_name,
			is_admin,
			is_blocked
		)
		VALUES (
			$1,
			'Resolver Integration Admin',
			true,
			false
		)
		RETURNING id
		`,
		adminEmail,
	).Scan(&adminID)

	if err != nil {
		t.Fatalf(
			"create admin: %v",
			err,
		)
	}

	admin := models.LocalAuthUser{
		ID:        adminID,
		Email:     adminEmail,
		FullName:  "Resolver Integration Admin",
		IsAdmin:   true,
		IsBlocked: false,
	}

	// ========================================================================
	// UNIDAD
	// ========================================================================

	unit, err := CreateInstitutionalUnit(
		admin,
		models.CreateInstitutionalUnitRequest{
			Name: fmt.Sprintf(
				"Resolver Integration %d",
				suffix,
			),
			Code: fmt.Sprintf(
				"RES-%d",
				suffix,
			),
			UnitType: models.
				InstitutionalUnitTypeAcademicProgram,
		},
	)

	if err != nil {
		t.Fatalf(
			"create institutional unit: %v",
			err,
		)
	}

	unitID = unit.ID

	// ========================================================================
	// RECURSO TEMPORAL
	// ========================================================================

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
			"load venue: %v",
			err,
		)
	}

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
		fmt.Sprintf(
			"Resolver Resource %d",
			suffix,
		),
	).Scan(&resourceID)

	if err != nil {
		t.Fatalf(
			"create resource: %v",
			err,
		)
	}

	_, err = database.DB.ExecContext(
		ctx,
		`
		INSERT INTO reservation_policy_resources (
			policy_id,
			resource_id
		)
		VALUES ($1, $2)
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

	// ========================================================================
	// HORARIO DE RESERVA VÁLIDO
	// ========================================================================

	now := businessclock.Now()

	reservationDate := now

	var startMinute int

	if reservableWindowDays > 1 {
		reservationDate =
			now.AddDate(0, 0, 1)

		startMinute = openingMinute
	} else {
		// Si la política solo permite el día actual, elegimos el siguiente
		// slot futuro disponible.
		currentMinute :=
			now.Hour()*60 +
				now.Minute() +
				slotInterval

		startMinute = currentMinute

		if startMinute < openingMinute {
			startMinute = openingMinute
		}
	}

	startMinute =
		((startMinute + slotInterval - 1) /
			slotInterval) *
			slotInterval

	if startMinute+durationMinutes >
		closingMinute {

		t.Fatalf(
			"policy has no future slot suitable for resolver test",
		)
	}

	reservationStart := time.Date(
		reservationDate.Year(),
		reservationDate.Month(),
		reservationDate.Day(),
		startMinute/60,
		startMinute%60,
		0,
		0,
		businessclock.Location(),
	)

	reservationEnd :=
		reservationStart.Add(
			time.Duration(durationMinutes) *
				time.Minute,
		)

	// ========================================================================
	// RESERVA R
	// ========================================================================
	//
	// Se inserta antes de las actividades porque una reserva posterior a una
	// actividad institucional sería correctamente rechazada por PG16_0006.

	var reservationID int

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
			"create reservation: %v",
			err,
		)
	}

	// ========================================================================
	// HELPERS
	// ========================================================================

	type activityOccurrence struct {
		activityID int
		scheduleID int
		start      time.Time
		end        time.Time
	}

	createActivity := func(
		title string,
		start time.Time,
		end time.Time,
	) activityOccurrence {
		t.Helper()

		var activityID int

		err := database.DB.QueryRowContext(
			ctx,
			`
			INSERT INTO institutional_activities (
				unit_id,
				resource_id,
				activity_type,
				title,
				status,
				requires_enrollment,
				capacity,
				created_by_user_id
			)
			VALUES (
				$1,
				$2,
				'ACADEMIC_CLASS',
				$3,
				'SCHEDULED',
				false,
				NULL,
				$4
			)
			RETURNING id
			`,
			unitID,
			resourceID,
			title,
			adminID,
		).Scan(&activityID)

		if err != nil {
			t.Fatalf(
				"create activity %s: %v",
				title,
				err,
			)
		}

		var scheduleID int

		err = database.DB.QueryRowContext(
			ctx,
			`
			INSERT INTO institutional_activity_schedules (
				activity_id,
				schedule_type,
				specific_date,
				start_time,
				end_time,
				is_active
			)
			VALUES (
				$1,
				'SINGLE',
				$2::date,
				$3::time,
				$4::time,
				true
			)
			RETURNING id
			`,
			activityID,
			start.Format("2006-01-02"),
			start.Format("15:04"),
			end.Format("15:04"),
		).Scan(&scheduleID)

		if err != nil {
			t.Fatalf(
				"create schedule %s: %v",
				title,
				err,
			)
		}

		return activityOccurrence{
			activityID: activityID,
			scheduleID: scheduleID,
			start:      start,
			end:        end,
		}
	}

	createConflict := func() int {
		t.Helper()

		var conflictID int

		err := database.DB.QueryRowContext(
			ctx,
			`
			INSERT INTO scheduling_conflicts (
				resource_id,
				status
			)
			VALUES (
				$1,
				'PENDING'
			)
			RETURNING id
			`,
			resourceID,
		).Scan(&conflictID)

		if err != nil {
			t.Fatalf(
				"create conflict: %v",
				err,
			)
		}

		return conflictID
	}

	addActivityItem := func(
		conflictID int,
		occ activityOccurrence,
	) int {
		t.Helper()

		var itemID int

		err := database.DB.QueryRowContext(
			ctx,
			`
			INSERT INTO scheduling_conflict_items (
				conflict_id,
				institutional_activity_id,
				schedule_id,
				occurrence_start,
				occurrence_end
			)
			VALUES (
				$1,
				$2,
				$3,
				$4,
				$5
			)
			RETURNING id
			`,
			conflictID,
			occ.activityID,
			occ.scheduleID,
			occ.start,
			occ.end,
		).Scan(&itemID)

		if err != nil {
			t.Fatalf(
				"create activity conflict item: %v",
				err,
			)
		}

		return itemID
	}

	addReservationItem := func(
		conflictID int,
	) int {
		t.Helper()

		var itemID int

		err := database.DB.QueryRowContext(
			ctx,
			`
			INSERT INTO scheduling_conflict_items (
				conflict_id,
				reservation_id,
				occurrence_start,
				occurrence_end
			)
			VALUES (
				$1,
				$2,
				$3,
				$4
			)
			RETURNING id
			`,
			conflictID,
			reservationID,
			reservationStart,
			reservationEnd,
		).Scan(&itemID)

		if err != nil {
			t.Fatalf(
				"create reservation conflict item: %v",
				err,
			)
		}

		return itemID
	}

	// ========================================================================
	// ESCENARIO 1
	// KEEP + KEEP INVÁLIDO / CANCEL REAL
	// ========================================================================

	activityA :=
		createActivity(
			fmt.Sprintf(
				"Resolver A %d",
				suffix,
			),
			reservationStart,
			reservationEnd,
		)

	activityB :=
		createActivity(
			fmt.Sprintf(
				"Resolver B %d",
				suffix,
			),
			reservationStart,
			reservationEnd,
		)

	conflict1 := createConflict()

	itemA :=
		addActivityItem(
			conflict1,
			activityA,
		)

	itemB :=
		addActivityItem(
			conflict1,
			activityB,
		)

	itemReservation :=
		addReservationItem(
			conflict1,
		)

	// A gana provisionalmente.
	resolved, err :=
		ResolveSchedulingConflictItem(
			conflict1,
			itemA,
			admin,
			models.ResolveSchedulingConflictItemRequest{
				Resolution: models.
					SchedulingItemResolutionKeep,

				ResolutionNote: "Actividad institucional prioritaria",
			},
		)

	if err != nil {
		t.Fatalf(
			"resolve A KEEP: %v",
			err,
		)
	}

	if resolved.Status !=
		models.SchedulingConflictStatusPending {

		t.Fatalf(
			"conflict should remain pending after partial resolution",
		)
	}

	// B no puede quedar KEEP porque A sigue ocupando exactamente el mismo
	// intervalo.
	_, err =
		ResolveSchedulingConflictItem(
			conflict1,
			itemB,
			admin,
			models.ResolveSchedulingConflictItemRequest{
				Resolution: models.
					SchedulingItemResolutionKeep,

				ResolutionNote: "Intento inválido KEEP + KEEP",
			},
		)

	if !errors.Is(
		err,
		repositories.
			ErrSchedulingResolutionInvalidPlan,
	) {
		t.Fatalf(
			"expected invalid plan error, got: %v",
			err,
		)
	}

	// La transacción inválida debe hacer rollback completo.
	var itemBResolution string

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT resolution
		FROM scheduling_conflict_items
		WHERE id = $1
		`,
		itemB,
	).Scan(&itemBResolution)

	if err != nil {
		t.Fatalf(
			"load B resolution: %v",
			err,
		)
	}

	if itemBResolution !=
		models.SchedulingItemResolutionPending {

		t.Fatalf(
			"B must remain PENDING after rollback, got %s",
			itemBResolution,
		)
	}

	// Cancelamos únicamente la ocurrencia B.
	_, err =
		ResolveSchedulingConflictItem(
			conflict1,
			itemB,
			admin,
			models.ResolveSchedulingConflictItemRequest{
				Resolution: models.
					SchedulingItemResolutionCancel,

				ResolutionNote: "Se mantiene actividad A",
			},
		)

	if err != nil {
		t.Fatalf(
			"resolve B CANCEL: %v",
			err,
		)
	}

	var cancelExceptionType string

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT exception_type

		FROM institutional_activity_schedule_exceptions

		WHERE schedule_id = $1
		  AND original_date = $2::date
		`,
		activityB.scheduleID,
		activityB.start.Format(
			"2006-01-02",
		),
	).Scan(&cancelExceptionType)

	if err != nil {
		t.Fatalf(
			"load CANCEL exception: %v",
			err,
		)
	}

	if cancelExceptionType != "CANCEL" {
		t.Fatalf(
			"expected CANCEL exception, got %s",
			cancelExceptionType,
		)
	}

	// Cancelamos la reserva.
	resolved, err =
		ResolveSchedulingConflictItem(
			conflict1,
			itemReservation,
			admin,
			models.ResolveSchedulingConflictItemRequest{
				Resolution: models.
					SchedulingItemResolutionCancel,

				ResolutionNote: "Reserva desplazada por programación institucional",
			},
		)

	if err != nil {
		t.Fatalf(
			"resolve reservation CANCEL: %v",
			err,
		)
	}

	if resolved.Status !=
		models.SchedulingConflictStatusResolved {

		t.Fatalf(
			"conflict 1 should be RESOLVED, got %s",
			resolved.Status,
		)
	}

	var (
		reservationStatus  string
		cancellationReason string
	)

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT
			status,
			COALESCE(cancellation_reason, '')

		FROM reservations

		WHERE id = $1
		`,
		reservationID,
	).Scan(
		&reservationStatus,
		&cancellationReason,
	)

	if err != nil {
		t.Fatalf(
			"load cancelled reservation: %v",
			err,
		)
	}

	if reservationStatus !=
		models.ReservationStatusCancelled {

		t.Fatalf(
			"reservation should be CANCELLED, got %s",
			reservationStatus,
		)
	}

	if cancellationReason == "" {
		t.Fatal(
			"reservation cancellation reason must be persisted",
		)
	}

	// ========================================================================
	// ESCENARIO 2
	// RESCHEDULE -> NUEVO CONFLICTO
	// ========================================================================

	originalDate :=
		reservationDate.AddDate(
			0,
			0,
			2,
		)

	originalStart := time.Date(
		originalDate.Year(),
		originalDate.Month(),
		originalDate.Day(),
		10,
		0,
		0,
		0,
		businessclock.Location(),
	)

	originalEnd :=
		originalStart.Add(time.Hour)

	targetDate :=
		originalDate.AddDate(
			0,
			0,
			1,
		)

	targetStart := time.Date(
		targetDate.Year(),
		targetDate.Month(),
		targetDate.Day(),
		14,
		0,
		0,
		0,
		businessclock.Location(),
	)

	targetEnd :=
		targetStart.Add(time.Hour)

	activityC :=
		createActivity(
			fmt.Sprintf(
				"Resolver C %d",
				suffix,
			),
			originalStart,
			originalEnd,
		)

	activityD :=
		createActivity(
			fmt.Sprintf(
				"Resolver D %d",
				suffix,
			),
			originalStart,
			originalEnd,
		)

	// E ocupa el destino al que D será trasladada.
	//
	// RESCHEDULE no debe ser rechazado: actividad ↔ actividad es un conflicto
	// administrativo resoluble, no un hard block.
	activityE :=
		createActivity(
			fmt.Sprintf(
				"Resolver E %d",
				suffix,
			),
			targetStart,
			targetEnd,
		)

	conflict2 := createConflict()

	itemC :=
		addActivityItem(
			conflict2,
			activityC,
		)

	itemD :=
		addActivityItem(
			conflict2,
			activityD,
		)

	_, err =
		ResolveSchedulingConflictItem(
			conflict2,
			itemC,
			admin,
			models.ResolveSchedulingConflictItemRequest{
				Resolution: models.
					SchedulingItemResolutionKeep,

				ResolutionNote: "Se conserva C en horario original",
			},
		)

	if err != nil {
		t.Fatalf(
			"resolve C KEEP: %v",
			err,
		)
	}

	newDateValue :=
		targetStart.Format(
			"2006-01-02",
		)

	newStartValue :=
		targetStart.Format(
			"15:04",
		)

	newEndValue :=
		targetEnd.Format(
			"15:04",
		)

	resolved, err =
		ResolveSchedulingConflictItem(
			conflict2,
			itemD,
			admin,
			models.ResolveSchedulingConflictItemRequest{
				Resolution: models.
					SchedulingItemResolutionReschedule,

				ResolutionNote: "D se traslada a un nuevo horario",

				NewDate: &newDateValue,

				NewStartTime: &newStartValue,

				NewEndTime: &newEndValue,
			},
		)

	if err != nil {
		t.Fatalf(
			"resolve D RESCHEDULE: %v",
			err,
		)
	}

	if resolved.Status !=
		models.SchedulingConflictStatusResolved {

		t.Fatalf(
			"original conflict should be RESOLVED after RESCHEDULE",
		)
	}

	var (
		exceptionType string
		newDate       string
	)

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT
			exception_type,
			to_char(new_date, 'YYYY-MM-DD')

		FROM institutional_activity_schedule_exceptions

		WHERE schedule_id = $1
		  AND original_date = $2::date
		`,
		activityD.scheduleID,
		originalStart.Format(
			"2006-01-02",
		),
	).Scan(
		&exceptionType,
		&newDate,
	)

	if err != nil {
		t.Fatalf(
			"load RESCHEDULE exception: %v",
			err,
		)
	}

	if exceptionType != "RESCHEDULE" ||
		newDate != newDateValue {

		t.Fatalf(
			"unexpected RESCHEDULE exception: type=%s date=%s",
			exceptionType,
			newDate,
		)
	}

	// ========================================================================
	// EL DESTINO DE D DEBE HABER GENERADO OTRO CONFLICTO CON E
	// ========================================================================

	var newConflictID int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT conflict.id

		FROM scheduling_conflicts conflict

		WHERE conflict.resource_id = $1
		  AND conflict.status = 'PENDING'
		  AND conflict.id <> $2

		  AND EXISTS (
				SELECT 1

				FROM scheduling_conflict_items item

				WHERE item.conflict_id = conflict.id
				  AND item.institutional_activity_id = $3
		  )

		  AND EXISTS (
				SELECT 1

				FROM scheduling_conflict_items item

				WHERE item.conflict_id = conflict.id
				  AND item.institutional_activity_id = $4
		  )

		ORDER BY conflict.id DESC

		LIMIT 1
		`,
		resourceID,
		conflict2,
		activityD.activityID,
		activityE.activityID,
	).Scan(&newConflictID)

	if err != nil {
		t.Fatalf(
			"expected new conflict after RESCHEDULE: %v",
			err,
		)
	}

	var (
		newItemD int
		newItemE int
	)

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT id

		FROM scheduling_conflict_items

		WHERE conflict_id = $1
		  AND institutional_activity_id = $2
		`,
		newConflictID,
		activityD.activityID,
	).Scan(&newItemD)

	if err != nil {
		t.Fatalf(
			"load rescheduled D item: %v",
			err,
		)
	}

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT id

		FROM scheduling_conflict_items

		WHERE conflict_id = $1
		  AND institutional_activity_id = $2
		`,
		newConflictID,
		activityE.activityID,
	).Scan(&newItemE)

	if err != nil {
		t.Fatalf(
			"load E conflict item: %v",
			err,
		)
	}

	// ========================================================================
	// ALLOW + ALLOW
	// ========================================================================

	_, err =
		ResolveSchedulingConflictItem(
			newConflictID,
			newItemD,
			admin,
			models.ResolveSchedulingConflictItemRequest{
				Resolution: models.
					SchedulingItemResolutionAllow,

				ResolutionNote: "Coexistencia administrativa autorizada",
			},
		)

	if err != nil {
		t.Fatalf(
			"resolve D ALLOW: %v",
			err,
		)
	}

	resolved, err =
		ResolveSchedulingConflictItem(
			newConflictID,
			newItemE,
			admin,
			models.ResolveSchedulingConflictItemRequest{
				Resolution: models.
					SchedulingItemResolutionAllow,

				ResolutionNote: "Coexistencia administrativa autorizada",
			},
		)

	if err != nil {
		t.Fatalf(
			"resolve E ALLOW: %v",
			err,
		)
	}

	if resolved.Status !=
		models.SchedulingConflictStatusResolved {

		t.Fatalf(
			"ALLOW + ALLOW conflict should be RESOLVED",
		)
	}

	// ========================================================================
	// DISPONIBILIDAD: D YA NO DEBE EXISTIR EN SU HORARIO ORIGINAL
	// ========================================================================

	originalActivities, err :=
		repositories.
			GetScheduledInstitutionalActivitiesForAvailability(
				originalStart.Add(-time.Minute),
				originalEnd.Add(time.Minute),
			)

	if err != nil {
		t.Fatalf(
			"load original availability: %v",
			err,
		)
	}

	for _, activity := range originalActivities {

		if activity.ID ==
			activityD.activityID {

			t.Fatal(
				"rescheduled activity D still appears at original occurrence",
			)
		}
	}

	// ========================================================================
	// DISPONIBILIDAD: D DEBE APARECER EN EL DESTINO
	// ========================================================================

	targetActivities, err :=
		repositories.
			GetScheduledInstitutionalActivitiesForAvailability(
				targetStart.Add(-time.Minute),
				targetEnd.Add(time.Minute),
			)

	if err != nil {
		t.Fatalf(
			"load target availability: %v",
			err,
		)
	}

	foundD := false

	for _, activity := range targetActivities {

		if activity.ID !=
			activityD.activityID {

			continue
		}

		if !activity.StartTime.Equal(
			targetStart,
		) {
			t.Fatalf(
				"unexpected rescheduled start: %v",
				activity.StartTime,
			)
		}

		foundD = true
	}

	if !foundD {
		t.Fatal(
			"rescheduled activity D does not appear at target occurrence",
		)
	}
}
