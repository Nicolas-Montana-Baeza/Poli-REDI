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

	"github.com/jackc/pgx/v5/pgconn"
)

// TestInstitutionalAvailabilityIntegration valida que una actividad
// institucional SCHEDULED proteja efectivamente su intervalo.
//
// Se comprueba:
//
//   - reserva que termina exactamente al comenzar la actividad -> permitida;
//   - reserva solapada -> rechazada;
//   - reserva que comienza exactamente al terminar -> permitida;
//   - OPEN_USE también queda protegido;
//   - availability_block posterior sobre la actividad -> rechazado.
//
// Los intervalos utilizan semántica [start,end).
func TestInstitutionalAvailabilityIntegration(
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
		unitID             int
		reservableResource int
		openUseResource    int
		createdUserIDs     []int
	)

	// ========================================================================
	// CLEANUP
	// ========================================================================

	defer func() {
		cleanupCtx := context.Background()

		for _, resourceID := range []int{
			reservableResource,
			openUseResource,
		} {
			if resourceID <= 0 {
				continue
			}

			// Los conflictos conservan FKs hacia sus ocupaciones.
			_, _ = database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM scheduling_conflicts
				WHERE resource_id = $1
				`,
				resourceID,
			)

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
				DELETE FROM availability_blocks
				WHERE resource_id = $1
				`,
				resourceID,
			)

			_, _ = database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM institutional_activities
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

		if unitID > 0 {
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

		for _, userID := range createdUserIDs {
			_, _ = database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM users
				WHERE id = $1
				`,
				userID,
			)
		}
	}()

	// ========================================================================
	// POLÍTICA VIGENTE
	// ========================================================================

	var (
		policyID       int
		windowDays     int
		openingMinute  int
		closingMinute  int
		slotInterval   int
		durationMinute int
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
		ORDER BY effective_from DESC, id DESC
		LIMIT 1
		`,
	).Scan(
		&policyID,
		&windowDays,
		&openingMinute,
		&closingMinute,
		&slotInterval,
	)

	if err != nil {
		t.Fatalf("load reservation policy: %v", err)
	}

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
	).Scan(&durationMinute)

	if err != nil {
		t.Fatalf("load reservation duration: %v", err)
	}

	// Necesitamos espacio para:
	//
	//   reserva anterior
	//   actividad
	//   reserva posterior

	activityStartMinute :=
		((openingMinute +
			durationMinute +
			slotInterval - 1) /
			slotInterval) *
			slotInterval

	activityEndMinute :=
		activityStartMinute +
			durationMinute

	if activityEndMinute+durationMinute >
		closingMinute {
		t.Fatalf(
			"policy does not provide enough room for boundary test",
		)
	}

	// ========================================================================
	// FECHA FUTURA DENTRO DE LA VENTANA
	// ========================================================================

	testDate := businessclock.Now()

	if windowDays > 1 {
		testDate =
			testDate.AddDate(0, 0, 1)
	}

	dateValue :=
		testDate.Format("2006-01-02")

	activityStart := time.Date(
		testDate.Year(),
		testDate.Month(),
		testDate.Day(),
		activityStartMinute/60,
		activityStartMinute%60,
		0,
		0,
		businessclock.Location(),
	)

	activityEnd := time.Date(
		testDate.Year(),
		testDate.Month(),
		testDate.Day(),
		activityEndMinute/60,
		activityEndMinute%60,
		0,
		0,
		businessclock.Location(),
	)

	// ========================================================================
	// USUARIOS TEMPORALES
	// ========================================================================

	createUser := func(label string) int {
		t.Helper()

		var userID int

		err := database.DB.QueryRowContext(
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
				$2,
				NULL,
				true,
				false
			)
			RETURNING id
			`,
			fmt.Sprintf(
				"%s.%d@test.local",
				label,
				suffix,
			),
			"Integration "+label,
		).Scan(&userID)

		if err != nil {
			t.Fatalf(
				"create user %s: %v",
				label,
				err,
			)
		}

		createdUserIDs = append(
			createdUserIDs,
			userID,
		)

		return userID
	}

	adminID :=
		createUser("institutional-availability-admin")

	beforeUserID :=
		createUser("institutional-before")

	overlapUserID :=
		createUser("institutional-overlap")

	afterUserID :=
		createUser("institutional-after")

	openUseUserID :=
		createUser("institutional-open-use")

	admin := models.LocalAuthUser{
		ID:        adminID,
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
				"Institutional Availability Integration %d",
				suffix,
			),
			Code: fmt.Sprintf(
				"IAI-%d",
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

	// ========================================================================
	// RECURSOS TEMPORALES
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
		t.Fatalf("load venue: %v", err)
	}

	createResource := func(
		name string,
		mode string,
	) int {
		t.Helper()

		var resourceID int

		err := database.DB.QueryRowContext(
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
				$3,
				50,
				true
			)
			RETURNING id
			`,
			venueID,
			name,
			mode,
		).Scan(&resourceID)

		if err != nil {
			t.Fatalf(
				"create %s resource: %v",
				mode,
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

		return resourceID
	}

	reservableResource = createResource(
		fmt.Sprintf(
			"Institutional RESERVABLE %d",
			suffix,
		),
		"RESERVABLE",
	)

	openUseResource = createResource(
		fmt.Sprintf(
			"Institutional OPEN_USE %d",
			suffix,
		),
		"OPEN_USE",
	)

	// ========================================================================
	// ACTIVIDADES SCHEDULED
	// ========================================================================

	createInstitutionalOccupation := func(
		resourceID int,
		title string,
	) {
		t.Helper()

		_, err := CreateInstitutionalActivity(
			admin,
			models.CreateInstitutionalActivityRequest{
				UnitID:       unitID,
				ResourceID:   resourceID,
				ActivityType: models.InstitutionalActivityTypeAcademicClass,
				Title:        title,

				RequiresEnrollment: false,

				Schedules: []models.CreateInstitutionalScheduleRequest{
					{
						ScheduleType: models.InstitutionalScheduleTypeSingle,

						SpecificDate: &dateValue,

						StartTime: institutionalAvailabilityClock(
							activityStartMinute,
						),

						EndTime: institutionalAvailabilityClock(
							activityEndMinute,
						),
					},
				},
			},
		)

		if err != nil {
			t.Fatalf(
				"create institutional activity: %v",
				err,
			)
		}
	}

	createInstitutionalOccupation(
		reservableResource,
		"RESERVABLE institutional occupation",
	)

	createInstitutionalOccupation(
		openUseResource,
		"OPEN_USE institutional occupation",
	)

	// ========================================================================
	// RESERVA QUE TERMINA EXACTAMENTE AL COMENZAR
	// ========================================================================

	beforeStart :=
		activityStart.Add(
			-time.Duration(durationMinute) *
				time.Minute,
		)

	_, err = CreateReservation(
		models.Reservation{
			UserID:          beforeUserID,
			ResourceID:      reservableResource,
			StartTime:       beforeStart,
			DurationMinutes: durationMinute,
		},
	)

	if err != nil {
		t.Fatalf(
			"touching-before reservation should succeed: %v",
			err,
		)
	}

	// ========================================================================
	// RESERVA SOLAPADA
	// ========================================================================

	_, err = CreateReservation(
		models.Reservation{
			UserID:          overlapUserID,
			ResourceID:      reservableResource,
			StartTime:       activityStart,
			DurationMinutes: durationMinute,
		},
	)

	if err == nil {
		t.Fatal(
			"overlapping reservation should have been rejected",
		)
	}

	if err.Error() !=
		"el recurso tiene programación institucional en ese horario" {
		t.Fatalf(
			"unexpected overlap error: %v",
			err,
		)
	}

	// ========================================================================
	// RESERVA QUE COMIENZA EXACTAMENTE AL TERMINAR
	// ========================================================================

	_, err = CreateReservation(
		models.Reservation{
			UserID:          afterUserID,
			ResourceID:      reservableResource,
			StartTime:       activityEnd,
			DurationMinutes: durationMinute,
		},
	)

	if err != nil {
		t.Fatalf(
			"touching-after reservation should succeed: %v",
			err,
		)
	}

	// ========================================================================
	// OPEN_USE TAMBIÉN DEBE SER RECHAZADO
	// ========================================================================

	_, err = CreateReservation(
		models.Reservation{
			UserID:          openUseUserID,
			ResourceID:      openUseResource,
			StartTime:       activityStart,
			DurationMinutes: durationMinute,
		},
	)

	if err == nil {
		t.Fatal(
			"OPEN_USE overlap should have been rejected",
		)
	}

	if err.Error() !=
		"el recurso tiene programación institucional en ese horario" {
		t.Fatalf(
			"unexpected OPEN_USE error: %v",
			err,
		)
	}

	// ========================================================================
	// AVAILABILITY BLOCK POSTERIOR
	// ========================================================================

	_, err = database.DB.ExecContext(
		ctx,
		`
		INSERT INTO availability_blocks (
			resource_id,
			created_by_user_id,
			block_type,
			reason,
			start_time,
			end_time,
			is_active
		)
		VALUES (
			$1,
			$2,
			'ADMINISTRATIVE',
			$3,
			$4,
			$5,
			true
		)
		`,
		reservableResource,
		adminID,
		fmt.Sprintf(
			"Institutional overlap %d",
			suffix,
		),
		activityStart,
		activityEnd,
	)

	if err == nil {
		t.Fatal(
			"availability block overlap should have been rejected",
		)
	}

	var pgErr *pgconn.PgError

	if !errors.As(err, &pgErr) {
		t.Fatalf(
			"expected PostgreSQL error for block overlap, got %v",
			err,
		)
	}

	if pgErr.Code != "P1012" {
		t.Fatalf(
			"expected P1012, got %s: %v",
			pgErr.Code,
			err,
		)
	}

	t.Logf(
		"PASS: institutional occupancy=%s-%s; touching boundaries allowed; RESERVABLE/OPEN_USE overlaps rejected; block rejected with P1012",
		activityStart.Format("15:04"),
		activityEnd.Format("15:04"),
	)
}

func institutionalAvailabilityClock(
	minute int,
) string {
	return fmt.Sprintf(
		"%02d:%02d",
		minute/60,
		minute%60,
	)
}
