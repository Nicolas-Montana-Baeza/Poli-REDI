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
)

// TestInstitutionalActivitiesIntegration valida la programación institucional
// básica contra PostgreSQL.
//
// El escenario demuestra que:
//
//   - una actividad SINGLE válida puede programarse;
//   - una actividad WEEKLY válida puede programarse;
//   - sus horarios quedan estructurados y recuperables;
//   - un availability_block representa indisponibilidad dura;
//   - una actividad que intersecta ese bloqueo es rechazada;
//   - el rechazo ocurre transaccionalmente y no deja actividad parcial.
//
// Todavía NO valida conflictos contra reservas u otras actividades.
// Esos solapamientos están permitidos y serán tratados por el detector
// scheduling_conflicts en el siguiente bloque del MVP2.
func TestInstitutionalActivitiesIntegration(t *testing.T) {
	if os.Getenv("POLIREDI_INTEGRATION") != "1" {
		t.Skip("integration test disabled")
	}

	database.Close()

	if err := database.Connect(); err != nil {
		t.Fatalf("connect postgres: %v", err)
	}

	defer database.Close()

	ctx := context.Background()

	// =========================================================================
	// DATOS CREADOS POR EL TEST
	// =========================================================================

	var (
		createdActivityIDs []int
		createdBlockIDs    []int
		createdUnitIDs     []int
		createdUserIDs     []int
	)

	// La limpieza se registra antes de crear datos para que también se ejecute
	// si una validación intermedia falla.
	defer func() {
		for _, activityID := range createdActivityIDs {
			if _, err := database.DB.ExecContext(
				context.Background(),
				`
				DELETE FROM institutional_activities
				WHERE id = $1
				`,
				activityID,
			); err != nil {
				t.Logf(
					"cleanup institutional activity %d: %v",
					activityID,
					err,
				)
			}
		}

		for _, blockID := range createdBlockIDs {
			if _, err := database.DB.ExecContext(
				context.Background(),
				`
				DELETE FROM availability_blocks
				WHERE id = $1
				`,
				blockID,
			); err != nil {
				t.Logf(
					"cleanup availability block %d: %v",
					blockID,
					err,
				)
			}
		}

		for _, unitID := range createdUnitIDs {
			if _, err := database.DB.ExecContext(
				context.Background(),
				`
				DELETE FROM institutional_unit_memberships
				WHERE unit_id = $1
				`,
				unitID,
			); err != nil {
				t.Logf(
					"cleanup memberships for unit %d: %v",
					unitID,
					err,
				)
			}

			if _, err := database.DB.ExecContext(
				context.Background(),
				`
				DELETE FROM institutional_units
				WHERE id = $1
				`,
				unitID,
			); err != nil {
				t.Logf(
					"cleanup institutional unit %d: %v",
					unitID,
					err,
				)
			}
		}

		for _, userID := range createdUserIDs {
			if _, err := database.DB.ExecContext(
				context.Background(),
				`
				DELETE FROM users
				WHERE id = $1
				`,
				userID,
			); err != nil {
				t.Logf(
					"cleanup user %d: %v",
					userID,
					err,
				)
			}
		}
	}()

	suffix := time.Now().UnixNano()

	// =========================================================================
	// ADMINISTRADOR TEMPORAL
	// =========================================================================

	var adminID int

	adminEmail := fmt.Sprintf(
		"mvp2.activity.admin.%d@test.local",
		suffix,
	)

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
			'Integration Institutional Activity Admin',
			NULL,
			true,
			false
		)
		RETURNING id
		`,
		adminEmail,
	).Scan(&adminID)

	if err != nil {
		t.Fatalf("create integration admin: %v", err)
	}

	createdUserIDs = append(
		createdUserIDs,
		adminID,
	)

	admin := models.LocalAuthUser{
		ID:        adminID,
		Email:     adminEmail,
		FullName:  "Integration Institutional Activity Admin",
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
				"Educación Física Activities Integration %d",
				suffix,
			),
			Code: fmt.Sprintf(
				"EFI-ACT-%d",
				suffix,
			),
			UnitType: models.InstitutionalUnitTypeAcademicProgram,
		},
	)

	if err != nil {
		t.Fatalf("create institutional unit: %v", err)
	}

	createdUnitIDs = append(
		createdUnitIDs,
		unit.ID,
	)

	// =========================================================================
	// RECURSO
	// =========================================================================
	//
	// No hardcodeamos un ID. Para programación institucional basta que el
	// recurso exista y se encuentre activo.
	//
	// En etapas posteriores validaremos la interacción específica entre
	// RESERVABLE, OPEN_USE y programación institucional.

	var resourceID int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT id
		FROM resources
		WHERE is_active = true
		ORDER BY id
		LIMIT 1
		`,
	).Scan(&resourceID)

	if err != nil {
		t.Fatalf("load active resource: %v", err)
	}

	// =========================================================================
	// FECHAS
	// =========================================================================
	//
	// Utilizamos fechas aproximadamente dos años hacia el futuro para evitar
	// interferencia con reservas de integración pertenecientes a otros tests.

	baseDate := businessclock.Now().
		AddDate(2, 0, 0)

	singleDate := baseDate.
		AddDate(0, 0, 7)

	weeklyStart := baseDate.
		AddDate(0, 0, 21)

	weeklyEnd := weeklyStart.
		AddDate(0, 0, 14)

	blockDate := baseDate.
		AddDate(0, 0, 49)

	singleDateValue :=
		singleDate.Format("2006-01-02")

	weeklyStartValue :=
		weeklyStart.Format("2006-01-02")

	weeklyEndValue :=
		weeklyEnd.Format("2006-01-02")

	blockDateValue :=
		blockDate.Format("2006-01-02")

	weeklyDay := isoWeekday(weeklyStart)

	// =========================================================================
	// 1. ACTIVIDAD SINGLE
	// =========================================================================

	singleActivity, err := CreateInstitutionalActivity(
		admin,
		models.CreateInstitutionalActivityRequest{
			UnitID:       unit.ID,
			ResourceID:   resourceID,
			ActivityType: models.InstitutionalActivityTypeAcademicClass,
			Title:        "Básquetbol I",
			Description:  "Clase puntual de integración",

			RequiresEnrollment: false,
			Capacity:           nil,

			Schedules: []models.CreateInstitutionalScheduleRequest{
				{
					ScheduleType: models.InstitutionalScheduleTypeSingle,
					SpecificDate: &singleDateValue,
					StartTime:    "10:00",
					EndTime:      "11:30",
				},
			},
		},
	)

	if err != nil {
		t.Fatalf(
			"create SINGLE institutional activity: %v",
			err,
		)
	}

	createdActivityIDs = append(
		createdActivityIDs,
		singleActivity.ID,
	)

	if singleActivity.Status !=
		models.InstitutionalActivityStatusScheduled {
		t.Fatalf(
			"expected SINGLE activity SCHEDULED, got %s",
			singleActivity.Status,
		)
	}

	if len(singleActivity.Schedules) != 1 {
		t.Fatalf(
			"expected 1 SINGLE schedule, got %d",
			len(singleActivity.Schedules),
		)
	}

	singleSchedule :=
		singleActivity.Schedules[0]

	if singleSchedule.ScheduleType !=
		models.InstitutionalScheduleTypeSingle {
		t.Fatalf(
			"expected SINGLE schedule type, got %s",
			singleSchedule.ScheduleType,
		)
	}

	if singleSchedule.SpecificDate == nil ||
		*singleSchedule.SpecificDate != singleDateValue {
		t.Fatalf(
			"unexpected SINGLE date: %v",
			singleSchedule.SpecificDate,
		)
	}

	if singleSchedule.StartTime != "10:00" ||
		singleSchedule.EndTime != "11:30" {
		t.Fatalf(
			"unexpected SINGLE schedule: %s-%s",
			singleSchedule.StartTime,
			singleSchedule.EndTime,
		)
	}

	// =========================================================================
	// 2. ACTIVIDAD WEEKLY
	// =========================================================================

	weeklyActivity, err := CreateInstitutionalActivity(
		admin,
		models.CreateInstitutionalActivityRequest{
			UnitID:       unit.ID,
			ResourceID:   resourceID,
			ActivityType: models.InstitutionalActivityTypeTraining,
			Title:        "Entrenamiento semanal",
			Description:  "Programación recurrente de integración",

			RequiresEnrollment: false,
			Capacity:           nil,

			Schedules: []models.CreateInstitutionalScheduleRequest{
				{
					ScheduleType: models.InstitutionalScheduleTypeWeekly,

					DayOfWeek: &weeklyDay,

					StartTime: "14:00",
					EndTime:   "15:30",

					ValidFrom: &weeklyStartValue,
					ValidTo:   &weeklyEndValue,
				},
			},
		},
	)

	if err != nil {
		t.Fatalf(
			"create WEEKLY institutional activity: %v",
			err,
		)
	}

	createdActivityIDs = append(
		createdActivityIDs,
		weeklyActivity.ID,
	)

	if weeklyActivity.Status !=
		models.InstitutionalActivityStatusScheduled {
		t.Fatalf(
			"expected WEEKLY activity SCHEDULED, got %s",
			weeklyActivity.Status,
		)
	}

	if len(weeklyActivity.Schedules) != 1 {
		t.Fatalf(
			"expected 1 WEEKLY schedule, got %d",
			len(weeklyActivity.Schedules),
		)
	}

	weeklySchedule :=
		weeklyActivity.Schedules[0]

	if weeklySchedule.ScheduleType !=
		models.InstitutionalScheduleTypeWeekly {
		t.Fatalf(
			"expected WEEKLY schedule type, got %s",
			weeklySchedule.ScheduleType,
		)
	}

	if weeklySchedule.DayOfWeek == nil ||
		*weeklySchedule.DayOfWeek != weeklyDay {
		t.Fatalf(
			"unexpected WEEKLY day: %v",
			weeklySchedule.DayOfWeek,
		)
	}

	if weeklySchedule.ValidFrom == nil ||
		*weeklySchedule.ValidFrom != weeklyStartValue {
		t.Fatalf(
			"unexpected WEEKLY validFrom: %v",
			weeklySchedule.ValidFrom,
		)
	}

	if weeklySchedule.ValidTo == nil ||
		*weeklySchedule.ValidTo != weeklyEndValue {
		t.Fatalf(
			"unexpected WEEKLY validTo: %v",
			weeklySchedule.ValidTo,
		)
	}

	// =========================================================================
	// 3. BLOQUEO ADMINISTRATIVO DURO
	// =========================================================================
	//
	// Insertamos primero el bloqueo.
	//
	// Después intentaremos programar una actividad sobre el mismo intervalo.
	// La actividad debe ser rechazada en lugar de generar scheduling_conflict.

	blockStart := time.Date(
		blockDate.Year(),
		blockDate.Month(),
		blockDate.Day(),
		10,
		0,
		0,
		0,
		businessclock.Location(),
	)

	blockEnd := time.Date(
		blockDate.Year(),
		blockDate.Month(),
		blockDate.Day(),
		12,
		0,
		0,
		0,
		businessclock.Location(),
	)

	var blockID int

	err = database.DB.QueryRowContext(
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
			'Integration institutional hard block',
			$3,
			$4,
			true
		)
		RETURNING id
		`,
		resourceID,
		admin.ID,
		blockStart,
		blockEnd,
	).Scan(&blockID)

	if err != nil {
		t.Fatalf(
			"create availability block: %v",
			err,
		)
	}

	createdBlockIDs = append(
		createdBlockIDs,
		blockID,
	)

	blockedTitle := fmt.Sprintf(
		"Actividad bloqueada %d",
		suffix,
	)

	_, err = CreateInstitutionalActivity(
		admin,
		models.CreateInstitutionalActivityRequest{
			UnitID:       unit.ID,
			ResourceID:   resourceID,
			ActivityType: models.InstitutionalActivityTypeAcademicClass,
			Title:        blockedTitle,

			RequiresEnrollment: false,
			Capacity:           nil,

			Schedules: []models.CreateInstitutionalScheduleRequest{
				{
					ScheduleType: models.InstitutionalScheduleTypeSingle,
					SpecificDate: &blockDateValue,
					StartTime:    "10:30",
					EndTime:      "11:30",
				},
			},
		},
	)

	if !errors.Is(
		err,
		ErrInstitutionalActivityBlocked,
	) {
		t.Fatalf(
			"expected hard block rejection, got %v",
			err,
		)
	}

	// =========================================================================
	// 4. EL RECHAZO DEBE HACER ROLLBACK COMPLETO
	// =========================================================================

	var blockedActivityCount int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT COUNT(*)
		FROM institutional_activities
		WHERE title = $1
		  AND unit_id = $2
		`,
		blockedTitle,
		unit.ID,
	).Scan(&blockedActivityCount)

	if err != nil {
		t.Fatalf(
			"check blocked activity rollback: %v",
			err,
		)
	}

	if blockedActivityCount != 0 {
		t.Fatalf(
			"blocked activity was persisted: count=%d",
			blockedActivityCount,
		)
	}

	// =========================================================================
	// RESULTADO
	// =========================================================================

	t.Logf(
		"PASS: SINGLE=%d WEEKLY=%d resource=%d block=%d; hard block rejected and transaction rolled back",
		singleActivity.ID,
		weeklyActivity.ID,
		resourceID,
		blockID,
	)
}
