package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
)

// ============================================================================
// WORKSHOP / CAPACIDAD CONCURRENTE
// ============================================================================
//
// Escenario:
//
//	capacidad = 1
//
//	Usuario A ─┐
//	           ├── inscripción simultánea
//	Usuario B ─┘
//
// Resultado obligatorio:
//
//	exactamente 1 CONFIRMED
//	exactamente 1 ErrInstitutionalWorkshopFull
//
// También validamos:
//
//   - inscripción duplicada;
//   - retiro;
//   - reapertura del cupo;
//   - reactivación de una inscripción CANCELLED;
//   - reutilización de la misma fila activity/user.
//
// El objetivo es demostrar que la capacidad no depende de un COUNT()
// desprotegido y que Poli-REDI no puede sobreinscribir un taller bajo
// concurrencia real.
func TestInstitutionalWorkshopConcurrentCapacityIntegration(
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
		unitID     int
		resourceID int
		workshopID int
		userAID    int
		userBID    int
	)

	// ========================================================================
	// CLEANUP
	// ========================================================================

	defer func() {
		cleanupCtx := context.Background()

		if workshopID > 0 {
			_, _ = database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM institutional_activity_enrollments
				WHERE activity_id = $1
				`,
				workshopID,
			)
		}

		if resourceID > 0 {
			// scheduling_conflict_items mantiene referencias históricas hacia
			// las actividades, por lo que los grupos deben eliminarse primero.
			_, _ = database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM scheduling_conflicts
				WHERE resource_id = $1
				`,
				resourceID,
			)
		}

		if workshopID > 0 {
			_, _ = database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM institutional_activities
				WHERE id = $1
				`,
				workshopID,
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

		if resourceID > 0 {
			_, _ = database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM resources
				WHERE id = $1
				`,
				resourceID,
			)
		}

		if userAID > 0 {
			_, _ = database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM users
				WHERE id = $1
				`,
				userAID,
			)
		}

		if userBID > 0 {
			_, _ = database.DB.ExecContext(
				cleanupCtx,
				`
				DELETE FROM users
				WHERE id = $1
				`,
				userBID,
			)
		}
	}()

	// ========================================================================
	// ADMIN EXISTENTE
	// ========================================================================

	var (
		adminID       int
		adminEmail    string
		adminFullName string
	)

	err := database.DB.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			email,
			full_name

		FROM users

		WHERE is_admin = true
		  AND is_blocked = false

		ORDER BY id

		LIMIT 1
		`,
	).Scan(
		&adminID,
		&adminEmail,
		&adminFullName,
	)

	if err != nil {
		t.Fatalf(
			"load integration admin: %v",
			err,
		)
	}

	admin := models.LocalAuthUser{
		ID:        adminID,
		Email:     adminEmail,
		FullName:  adminFullName,
		IsAdmin:   true,
		IsBlocked: false,
	}

	// ========================================================================
	// USUARIOS QUE COMPETIRÁN POR EL ÚLTIMO CUPO
	// ========================================================================

	createUser := func(
		label string,
	) models.LocalAuthUser {
		t.Helper()

		email := fmt.Sprintf(
			"workshop.%s.%d@test.local",
			label,
			suffix,
		)

		var userID int

		err := database.DB.QueryRowContext(
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
				$2,
				false,
				false
			)
			RETURNING id
			`,
			email,
			"Workshop User "+label,
		).Scan(
			&userID,
		)

		if err != nil {
			t.Fatalf(
				"create user %s: %v",
				label,
				err,
			)
		}

		return models.LocalAuthUser{
			ID:        userID,
			Email:     email,
			FullName:  "Workshop User " + label,
			IsAdmin:   false,
			IsBlocked: false,
		}
	}

	userA := createUser("A")
	userAID = userA.ID

	userB := createUser("B")
	userBID = userB.ID

	// ========================================================================
	// UNIDAD INSTITUCIONAL
	// ========================================================================

	unit, err := CreateInstitutionalUnit(
		admin,
		models.CreateInstitutionalUnitRequest{
			Name: fmt.Sprintf(
				"Workshop Integration %d",
				suffix,
			),
			Code: fmt.Sprintf(
				"WKS-%d",
				suffix,
			),
			UnitType: models.
				InstitutionalUnitTypeSportsUnit,
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
	// RECURSO EXCLUSIVO DEL TEST
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
	).Scan(
		&venueID,
	)

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
			'ADMIN_ONLY',
			30,
			true
		)
		RETURNING id
		`,
		venueID,
		fmt.Sprintf(
			"Workshop Resource %d",
			suffix,
		),
	).Scan(
		&resourceID,
	)

	if err != nil {
		t.Fatalf(
			"create resource: %v",
			err,
		)
	}

	// ========================================================================
	// WORKSHOP CON CAPACIDAD 1
	// ========================================================================

	capacity := 1

	workshopDate :=
		businessclock.Now().
			AddDate(0, 0, 2).
			Format("2006-01-02")

	workshop, err :=
		CreateInstitutionalActivity(
			admin,
			models.CreateInstitutionalActivityRequest{
				UnitID: unitID,

				ResourceID: resourceID,

				ActivityType: models.
					InstitutionalActivityTypeWorkshop,

				Title: fmt.Sprintf(
					"Workshop Concurrent Capacity %d",
					suffix,
				),

				Description: "Integration test de capacidad concurrente",

				RequiresEnrollment: true,

				Capacity: &capacity,

				Schedules: []models.CreateInstitutionalScheduleRequest{
					{
						ScheduleType: models.
							InstitutionalScheduleTypeSingle,

						SpecificDate: &workshopDate,

						StartTime: "10:00",

						EndTime: "11:00",
					},
				},
			},
		)

	if err != nil {
		t.Fatalf(
			"create workshop: %v",
			err,
		)
	}

	workshopID = workshop.ID

	// ========================================================================
	// CARRERA CONCURRENTE POR EL ÚNICO CUPO
	// ========================================================================

	type enrollmentResult struct {
		user models.LocalAuthUser
		err  error
	}

	results :=
		make(
			chan enrollmentResult,
			2,
		)

	start :=
		make(chan struct{})

	var ready sync.WaitGroup

	ready.Add(2)

	enroll := func(
		user models.LocalAuthUser,
	) {
		ready.Done()

		// Barrera para aproximar al máximo el inicio simultáneo.
		<-start

		_, err :=
			EnrollInInstitutionalWorkshop(
				workshopID,
				user,
			)

		results <- enrollmentResult{
			user: user,
			err:  err,
		}
	}

	go enroll(userA)
	go enroll(userB)

	ready.Wait()

	close(start)

	first := <-results
	second := <-results

	resultValues :=
		[]enrollmentResult{
			first,
			second,
		}

	successes := 0
	fullErrors := 0

	var (
		winner models.LocalAuthUser
		loser  models.LocalAuthUser
	)

	for _, result := range resultValues {

		switch {

		case result.err == nil:

			successes++
			winner = result.user

		case errors.Is(
			result.err,
			repositories.
				ErrInstitutionalWorkshopFull,
		):

			fullErrors++
			loser = result.user

		default:

			t.Fatalf(
				"unexpected concurrent enrollment result for user %d: %v",
				result.user.ID,
				result.err,
			)
		}
	}

	if successes != 1 {
		t.Fatalf(
			"expected exactly 1 successful enrollment, got %d",
			successes,
		)
	}

	if fullErrors != 1 {
		t.Fatalf(
			"expected exactly 1 FULL result, got %d",
			fullErrors,
		)
	}

	// ========================================================================
	// INVARIANTE FÍSICA EN POSTGRESQL
	// ========================================================================

	var confirmedCount int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT COUNT(*)::integer

		FROM institutional_activity_enrollments

		WHERE activity_id = $1
		  AND status = 'CONFIRMED'
		`,
		workshopID,
	).Scan(
		&confirmedCount,
	)

	if err != nil {
		t.Fatalf(
			"count confirmed enrollments: %v",
			err,
		)
	}

	if confirmedCount != 1 {
		t.Fatalf(
			"capacity invariant broken: expected 1 CONFIRMED, got %d",
			confirmedCount,
		)
	}

	// ========================================================================
	// VISTA DEL GANADOR Y DEL PERDEDOR
	// ========================================================================

	winnerView, err :=
		GetInstitutionalWorkshop(
			workshopID,
			winner,
		)

	if err != nil {
		t.Fatalf(
			"load winner workshop view: %v",
			err,
		)
	}

	if winnerView.IsEnrolled == nil ||
		!*winnerView.IsEnrolled {

		t.Fatal(
			"winner should appear enrolled",
		)
	}

	if winnerView.AvailableSpots == nil ||
		*winnerView.AvailableSpots != 0 {

		t.Fatalf(
			"expected 0 available spots after winner, got %v",
			winnerView.AvailableSpots,
		)
	}

	loserView, err :=
		GetInstitutionalWorkshop(
			workshopID,
			loser,
		)

	if err != nil {
		t.Fatalf(
			"load loser workshop view: %v",
			err,
		)
	}

	if loserView.IsEnrolled == nil ||
		*loserView.IsEnrolled {

		t.Fatal(
			"loser must not appear enrolled",
		)
	}

	// ========================================================================
	// INSCRIPCIÓN DUPLICADA
	// ========================================================================

	_, err =
		EnrollInInstitutionalWorkshop(
			workshopID,
			winner,
		)

	if !errors.Is(
		err,
		repositories.
			ErrInstitutionalWorkshopAlreadyEnrolled,
	) {
		t.Fatalf(
			"expected already enrolled error, got: %v",
			err,
		)
	}

	// ========================================================================
	// RETIRO DEL GANADOR
	// ========================================================================

	afterLeave, err :=
		LeaveInstitutionalWorkshop(
			workshopID,
			winner,
		)

	if err != nil {
		t.Fatalf(
			"winner leave workshop: %v",
			err,
		)
	}

	if afterLeave.IsEnrolled == nil ||
		*afterLeave.IsEnrolled {

		t.Fatal(
			"winner should no longer appear enrolled",
		)
	}

	if afterLeave.AvailableSpots == nil ||
		*afterLeave.AvailableSpots != 1 {

		t.Fatalf(
			"expected cupo to reopen after leave, got %v",
			afterLeave.AvailableSpots,
		)
	}

	// ========================================================================
	// REACTIVACIÓN DE LA MISMA INSCRIPCIÓN
	// ========================================================================

	_, err =
		EnrollInInstitutionalWorkshop(
			workshopID,
			winner,
		)

	if err != nil {
		t.Fatalf(
			"reactivate winner enrollment: %v",
			err,
		)
	}

	var winnerEnrollmentRows int

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT COUNT(*)::integer

		FROM institutional_activity_enrollments

		WHERE activity_id = $1
		  AND user_id = $2
		`,
		workshopID,
		winner.ID,
	).Scan(
		&winnerEnrollmentRows,
	)

	if err != nil {
		t.Fatalf(
			"count winner enrollment rows: %v",
			err,
		)
	}

	if winnerEnrollmentRows != 1 {
		t.Fatalf(
			"reactivation must reuse enrollment row, got %d rows",
			winnerEnrollmentRows,
		)
	}

	// ========================================================================
	// LIBERAMOS NUEVAMENTE EL CUPO
	// ========================================================================

	_, err =
		LeaveInstitutionalWorkshop(
			workshopID,
			winner,
		)

	if err != nil {
		t.Fatalf(
			"winner second leave: %v",
			err,
		)
	}

	// ========================================================================
	// EL PERDEDOR AHORA PUEDE INSCRIBIRSE
	// ========================================================================

	finalView, err :=
		EnrollInInstitutionalWorkshop(
			workshopID,
			loser,
		)

	if err != nil {
		t.Fatalf(
			"loser enroll after capacity reopened: %v",
			err,
		)
	}

	if finalView.IsEnrolled == nil ||
		!*finalView.IsEnrolled {

		t.Fatal(
			"loser should become enrolled after cupo reopens",
		)
	}

	if finalView.EnrollmentCount != 1 {
		t.Fatalf(
			"expected final enrollment count 1, got %d",
			finalView.EnrollmentCount,
		)
	}

	if finalView.AvailableSpots == nil ||
		*finalView.AvailableSpots != 0 {

		t.Fatalf(
			"expected final available spots 0, got %v",
			finalView.AvailableSpots,
		)
	}
}
