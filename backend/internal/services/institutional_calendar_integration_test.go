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

// TestInstitutionalCalendarIntegration valida la integración final entre
// Programación Institucional y GetAvailabilityItems.
//
// Se comprueba que:
//
//   - SINGLE aparece como una ocupación concreta;
//   - WEEKLY se materializa en múltiples ocurrencias;
//   - DRAFT no aparece;
//   - CANCELLED no aparece;
//   - una actividad fuera del rango consultado no aparece;
//   - las ocurrencias institucionales se identifican mediante
//     IsScheduledActivity=true.
func TestInstitutionalCalendarIntegration(t *testing.T) {
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
		adminID    int
		unitID     int
		resourceID int
	)

	// ========================================================================
	// CLEANUP
	// ========================================================================

	defer func() {
		cleanupCtx := context.Background()

		if resourceID > 0 {
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
				DELETE FROM institutional_activities
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
	// ADMIN
	// ========================================================================

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
			'Institutional Calendar Integration Admin',
			NULL,
			true,
			false
		)
		RETURNING id
		`,
		fmt.Sprintf(
			"calendar.integration.%d@test.local",
			suffix,
		),
	).Scan(&adminID)

	if err != nil {
		t.Fatalf("create admin: %v", err)
	}

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
				"Calendar Integration Unit %d",
				suffix,
			),
			Code: fmt.Sprintf(
				"CAL-%d",
				suffix,
			),
			UnitType: models.InstitutionalUnitTypeAcademicProgram,
		},
	)

	if err != nil {
		t.Fatalf("create unit: %v", err)
	}

	unitID = unit.ID

	// ========================================================================
	// RECURSO EXCLUSIVO
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
			'ADMIN_ONLY',
			50,
			true
		)
		RETURNING id
		`,
		venueID,
		fmt.Sprintf(
			"Institutional Calendar Resource %d",
			suffix,
		),
	).Scan(&resourceID)

	if err != nil {
		t.Fatalf("create resource: %v", err)
	}

	// ========================================================================
	// RANGO DEL CALENDARIO
	// ========================================================================
	//
	// Buscamos un lunes futuro. WEEKLY abarcará tres lunes consecutivos.

	base := businessclock.Now().
		AddDate(0, 1, 0)

	for base.Weekday() != time.Monday {
		base = base.AddDate(0, 0, 1)
	}

	base = time.Date(
		base.Year(),
		base.Month(),
		base.Day(),
		0,
		0,
		0,
		0,
		businessclock.Location(),
	)

	queryFrom := base
	queryTo := base.AddDate(0, 0, 15)

	singleDate := base.
		AddDate(0, 0, 1).
		Format("2006-01-02")

	weeklyFrom := base.Format("2006-01-02")
	weeklyTo := base.
		AddDate(0, 0, 14).
		Format("2006-01-02")

	weeklyDay := 1 // ISO Monday

	draftDate := base.
		AddDate(0, 0, 2).
		Format("2006-01-02")

	cancelledDate := base.
		AddDate(0, 0, 3).
		Format("2006-01-02")

	outsideDate := base.
		AddDate(0, 0, 30).
		Format("2006-01-02")

	// ========================================================================
	// HELPER
	// ========================================================================

	createSingle := func(
		title string,
		date string,
		start string,
		end string,
	) models.InstitutionalActivity {
		t.Helper()

		activity, err := CreateInstitutionalActivity(
			admin,
			models.CreateInstitutionalActivityRequest{
				UnitID:       unitID,
				ResourceID:   resourceID,
				ActivityType: models.InstitutionalActivityTypeEvent,
				Title:        title,

				RequiresEnrollment: false,

				Schedules: []models.CreateInstitutionalScheduleRequest{
					{
						ScheduleType: models.InstitutionalScheduleTypeSingle,

						SpecificDate: &date,
						StartTime:    start,
						EndTime:      end,
					},
				},
			},
		)

		if err != nil {
			t.Fatalf(
				"create activity %q: %v",
				title,
				err,
			)
		}

		return activity
	}

	// ========================================================================
	// SINGLE
	// ========================================================================

	singleTitle := fmt.Sprintf(
		"Calendar SINGLE %d",
		suffix,
	)

	createSingle(
		singleTitle,
		singleDate,
		"10:00",
		"11:00",
	)

	// ========================================================================
	// WEEKLY
	// ========================================================================

	weeklyTitle := fmt.Sprintf(
		"Calendar WEEKLY %d",
		suffix,
	)

	_, err = CreateInstitutionalActivity(
		admin,
		models.CreateInstitutionalActivityRequest{
			UnitID:       unitID,
			ResourceID:   resourceID,
			ActivityType: models.InstitutionalActivityTypeTraining,
			Title:        weeklyTitle,

			RequiresEnrollment: false,

			Schedules: []models.CreateInstitutionalScheduleRequest{
				{
					ScheduleType: models.InstitutionalScheduleTypeWeekly,

					DayOfWeek: &weeklyDay,

					StartTime: "14:00",
					EndTime:   "15:00",

					ValidFrom: &weeklyFrom,
					ValidTo:   &weeklyTo,
				},
			},
		},
	)

	if err != nil {
		t.Fatalf("create WEEKLY activity: %v", err)
	}

	// ========================================================================
	// DRAFT
	// ========================================================================

	draftTitle := fmt.Sprintf(
		"Calendar DRAFT %d",
		suffix,
	)

	draft := createSingle(
		draftTitle,
		draftDate,
		"16:00",
		"17:00",
	)

	_, err = database.DB.ExecContext(
		ctx,
		`
		UPDATE institutional_activities
		SET status = 'DRAFT'
		WHERE id = $1
		`,
		draft.ID,
	)

	if err != nil {
		t.Fatalf("set DRAFT status: %v", err)
	}

	// ========================================================================
	// CANCELLED
	// ========================================================================

	cancelledTitle := fmt.Sprintf(
		"Calendar CANCELLED %d",
		suffix,
	)

	cancelled := createSingle(
		cancelledTitle,
		cancelledDate,
		"17:00",
		"18:00",
	)

	_, err = database.DB.ExecContext(
		ctx,
		`
		UPDATE institutional_activities
		SET status = 'CANCELLED'
		WHERE id = $1
		`,
		cancelled.ID,
	)

	if err != nil {
		t.Fatalf("set CANCELLED status: %v", err)
	}

	// ========================================================================
	// FUERA DEL RANGO
	// ========================================================================

	outsideTitle := fmt.Sprintf(
		"Calendar OUTSIDE %d",
		suffix,
	)

	createSingle(
		outsideTitle,
		outsideDate,
		"12:00",
		"13:00",
	)

	// ========================================================================
	// GET AVAILABILITY
	// ========================================================================

	items, err := GetAvailabilityItems(
		queryFrom,
		queryTo,
		adminID,
		true,
	)

	if err != nil {
		t.Fatalf(
			"GetAvailabilityItems: %v",
			err,
		)
	}

	// Solo analizamos nuestro recurso exclusivo.
	institutionalItems :=
		[]models.AvailabilityItem{}

	for _, item := range items {
		if item.ResourceID != resourceID {
			continue
		}

		if !item.IsScheduledActivity {
			continue
		}

		institutionalItems = append(
			institutionalItems,
			item,
		)
	}

	// SINGLE = 1
	// WEEKLY = 3
	//
	// Total esperado = 4.
	if len(institutionalItems) != 4 {
		t.Fatalf(
			"expected 4 institutional calendar items, got %d",
			len(institutionalItems),
		)
	}

	var (
		singleCount    int
		weeklyCount    int
		draftCount     int
		cancelledCount int
		outsideCount   int
	)

	keys := map[string]struct{}{}

	for _, item := range institutionalItems {
		if !item.IsScheduledActivity {
			t.Fatalf(
				"item %q is not marked as scheduled activity",
				item.Title,
			)
		}

		if item.Type != "scheduled" {
			t.Fatalf(
				"expected type scheduled, got %q",
				item.Type,
			)
		}

		if item.AvailabilityKey == "" {
			t.Fatal(
				"scheduled activity has empty availability key",
			)
		}

		if _, exists :=
			keys[item.AvailabilityKey]; exists {
			t.Fatalf(
				"duplicate availability key: %s",
				item.AvailabilityKey,
			)
		}

		keys[item.AvailabilityKey] =
			struct{}{}

		switch item.Title {
		case singleTitle:
			singleCount++

			if item.ActivityType !=
				models.InstitutionalActivityTypeEvent {
				t.Fatalf(
					"unexpected SINGLE activity type: %s",
					item.ActivityType,
				)
			}

		case weeklyTitle:
			weeklyCount++

			if item.ActivityType !=
				models.InstitutionalActivityTypeTraining {
				t.Fatalf(
					"unexpected WEEKLY activity type: %s",
					item.ActivityType,
				)
			}

		case draftTitle:
			draftCount++

		case cancelledTitle:
			cancelledCount++

		case outsideTitle:
			outsideCount++
		}
	}

	if singleCount != 1 {
		t.Fatalf(
			"expected SINGLE once, got %d",
			singleCount,
		)
	}

	if weeklyCount != 3 {
		t.Fatalf(
			"expected WEEKLY 3 times, got %d",
			weeklyCount,
		)
	}

	if draftCount != 0 {
		t.Fatalf(
			"DRAFT activity appeared %d times",
			draftCount,
		)
	}

	if cancelledCount != 0 {
		t.Fatalf(
			"CANCELLED activity appeared %d times",
			cancelledCount,
		)
	}

	if outsideCount != 0 {
		t.Fatalf(
			"out-of-range activity appeared %d times",
			outsideCount,
		)
	}

	t.Logf(
		"PASS: calendar contains SINGLE=1 WEEKLY=3 DRAFT=0 CANCELLED=0 OUTSIDE=0 resource=%d",
		resourceID,
	)
}
