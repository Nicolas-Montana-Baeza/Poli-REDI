package repositories

import (
	"context"
	"database/sql"
	"errors"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

// ============================================================================
// ERRORES DE TALLERES INSTITUCIONALES
// ============================================================================

var (
	ErrInstitutionalWorkshopNotFound = errors.New(
		"taller institucional no encontrado",
	)

	ErrInstitutionalWorkshopUnavailable = errors.New(
		"el taller no se encuentra disponible para inscripción",
	)

	ErrInstitutionalWorkshopAlreadyEnrolled = errors.New(
		"el usuario ya está inscrito en este taller",
	)

	ErrInstitutionalWorkshopNotEnrolled = errors.New(
		"el usuario no está inscrito en este taller",
	)

	ErrInstitutionalWorkshopFull = errors.New(
		"el taller no tiene cupos disponibles",
	)
)

// ============================================================================
// LISTADO DE TALLERES
// ============================================================================

// GetScheduledInstitutionalWorkshopsForUser devuelve únicamente talleres
// actualmente programados.
//
// WORKSHOP no constituye una entidad paralela: es una especialización de
// institutional_activities.
//
// Esto evita volver a mantener dos calendarios, dos sistemas de conflictos y
// dos modelos diferentes de ocupación del recurso.
func GetScheduledInstitutionalWorkshopsForUser(
	userID int,
) ([]models.InstitutionalActivity, error) {
	rows, err := database.DB.QueryContext(
		context.Background(),
		`
		SELECT
			activity.id,
			activity.unit_id,
			activity.resource_id,
			activity.activity_type,
			activity.title,
			COALESCE(activity.description, ''),
			activity.status,
			activity.requires_enrollment,
			activity.capacity,
			activity.created_by_user_id,
			activity.created_at,
			activity.updated_at,

			unit.name,
			resource.name,
			creator.full_name,

			(
				SELECT COUNT(*)::integer

				FROM institutional_activity_enrollments enrollment

				WHERE enrollment.activity_id = activity.id
				  AND enrollment.status = 'CONFIRMED'
			) AS enrollment_count

		FROM institutional_activities activity

		INNER JOIN institutional_units unit
			ON unit.id = activity.unit_id

		INNER JOIN resources resource
			ON resource.id = activity.resource_id

		INNER JOIN users creator
			ON creator.id = activity.created_by_user_id

		WHERE activity.activity_type = 'WORKSHOP'
		  AND activity.status = 'SCHEDULED'
		  AND activity.requires_enrollment = true
		  AND unit.is_active = true
		  AND resource.is_active = true

		ORDER BY
			activity.title,
			activity.id;
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	workshops :=
		[]models.InstitutionalActivity{}

	for rows.Next() {
		workshop, err :=
			scanInstitutionalActivity(
				rows,
			)

		if err != nil {
			return nil, err
		}

		if err :=
			hydrateInstitutionalWorkshopForUser(
				&workshop,
				userID,
			); err != nil {

			return nil, err
		}

		workshops = append(
			workshops,
			workshop,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return workshops, nil
}

// ============================================================================
// DETALLE DE TALLER
// ============================================================================

func GetScheduledInstitutionalWorkshopForUser(
	workshopID int,
	userID int,
) (models.InstitutionalActivity, error) {
	row := database.DB.QueryRowContext(
		context.Background(),
		`
		SELECT
			activity.id,
			activity.unit_id,
			activity.resource_id,
			activity.activity_type,
			activity.title,
			COALESCE(activity.description, ''),
			activity.status,
			activity.requires_enrollment,
			activity.capacity,
			activity.created_by_user_id,
			activity.created_at,
			activity.updated_at,

			unit.name,
			resource.name,
			creator.full_name,

			(
				SELECT COUNT(*)::integer

				FROM institutional_activity_enrollments enrollment

				WHERE enrollment.activity_id = activity.id
				  AND enrollment.status = 'CONFIRMED'
			) AS enrollment_count

		FROM institutional_activities activity

		INNER JOIN institutional_units unit
			ON unit.id = activity.unit_id

		INNER JOIN resources resource
			ON resource.id = activity.resource_id

		INNER JOIN users creator
			ON creator.id = activity.created_by_user_id

		WHERE activity.id = $1
		  AND activity.activity_type = 'WORKSHOP'
		  AND activity.status = 'SCHEDULED'
		  AND activity.requires_enrollment = true
		  AND unit.is_active = true
		  AND resource.is_active = true;
		`,
		workshopID,
	)

	workshop, err :=
		scanInstitutionalActivity(
			row,
		)

	if err != nil {
		if errors.Is(
			err,
			sql.ErrNoRows,
		) {
			return models.InstitutionalActivity{},
				ErrInstitutionalWorkshopNotFound
		}

		return models.InstitutionalActivity{},
			err
	}

	if err :=
		hydrateInstitutionalWorkshopForUser(
			&workshop,
			userID,
		); err != nil {

		return models.InstitutionalActivity{},
			err
	}

	return workshop, nil
}

// ============================================================================
// HIDRATACIÓN DE DATOS DE INSCRIPCIÓN
// ============================================================================

func hydrateInstitutionalWorkshopForUser(
	workshop *models.InstitutionalActivity,
	userID int,
) error {
	schedules, err :=
		GetInstitutionalActivitySchedules(
			workshop.ID,
		)

	if err != nil {
		return err
	}

	// La vista pública solo necesita reglas de horario vigentes.
	activeSchedules :=
		make(
			[]models.InstitutionalActivitySchedule,
			0,
			len(schedules),
		)

	for _, schedule := range schedules {

		if !schedule.IsActive {
			continue
		}

		activeSchedules = append(
			activeSchedules,
			schedule,
		)
	}

	workshop.Schedules =
		activeSchedules

	isEnrolled := false

	if userID > 0 {
		err := database.DB.QueryRowContext(
			context.Background(),
			`
			SELECT EXISTS (
				SELECT 1

				FROM institutional_activity_enrollments

				WHERE activity_id = $1
				  AND user_id = $2
				  AND status = 'CONFIRMED'
			);
			`,
			workshop.ID,
			userID,
		).Scan(
			&isEnrolled,
		)

		if err != nil {
			return err
		}
	}

	availableSpots := 0

	if workshop.Capacity != nil {
		availableSpots =
			*workshop.Capacity -
				workshop.EnrollmentCount

		if availableSpots < 0 {
			// La BD no debería alcanzar este estado porque las altas se
			// serializan por actividad. El clamp evita exponer cupos negativos
			// si se inspeccionan datos históricos inconsistentes.
			availableSpots = 0
		}
	}

	workshop.IsEnrolled =
		&isEnrolled

	workshop.AvailableSpots =
		&availableSpots

	return nil
}

// ============================================================================
// INSCRIPCIÓN TRANSACCIONAL
// ============================================================================

// EnrollUserInInstitutionalWorkshop utiliza un lock de fila sobre la actividad.
//
// Todos los cambios de cupos de un mismo taller pasan primero por ese lock:
//
//	lock activity
//	     ↓
//	contar CONFIRMED
//	     ↓
//	insertar/reactivar
//
// Por lo tanto dos usuarios que intentan tomar simultáneamente el último cupo
// no pueden confirmar ambos.
//
// READ COMMITTED es deliberado.
//
// El FOR UPDATE sobre la actividad serializa todas las modificaciones de cupos.
// Una transacción concurrente espera ese lock y, al continuar, sus siguientes
// statements observan el estado ya confirmado por la transacción anterior.
//
// Esto permite devolver la regla de negocio "sin cupos" en vez de propagar un
// serialization_failure como comportamiento normal del usuario.
func EnrollUserInInstitutionalWorkshop(
	workshopID int,
	userID int,
) (models.InstitutionalActivity, error) {
	ctx :=
		context.Background()

	tx, err := database.DB.BeginTx(
		ctx,
		&sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
		},
	)

	if err != nil {
		return models.InstitutionalActivity{},
			err
	}

	defer tx.Rollback()

	var (
		activityType       string
		status             string
		requiresEnrollment bool
		capacity           sql.NullInt64

		unitActive     bool
		resourceActive bool
	)

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT
			activity.activity_type,
			activity.status,
			activity.requires_enrollment,
			activity.capacity,
			unit.is_active,
			resource.is_active

		FROM institutional_activities activity

		INNER JOIN institutional_units unit
			ON unit.id = activity.unit_id

		INNER JOIN resources resource
			ON resource.id = activity.resource_id

		WHERE activity.id = $1

		FOR UPDATE OF activity;
		`,
		workshopID,
	).Scan(
		&activityType,
		&status,
		&requiresEnrollment,
		&capacity,
		&unitActive,
		&resourceActive,
	)

	if err != nil {
		if errors.Is(
			err,
			sql.ErrNoRows,
		) {
			return models.InstitutionalActivity{},
				ErrInstitutionalWorkshopNotFound
		}

		return models.InstitutionalActivity{},
			err
	}

	if activityType !=
		models.InstitutionalActivityTypeWorkshop {

		return models.InstitutionalActivity{},
			ErrInstitutionalWorkshopNotFound
	}

	if status !=
		models.InstitutionalActivityStatusScheduled ||
		!requiresEnrollment ||
		!capacity.Valid ||
		capacity.Int64 <= 0 ||
		!unitActive ||
		!resourceActive {

		return models.InstitutionalActivity{},
			ErrInstitutionalWorkshopUnavailable
	}

	var existingStatus string

	existingErr :=
		tx.QueryRowContext(
			ctx,
			`
			SELECT status

			FROM institutional_activity_enrollments

			WHERE activity_id = $1
			  AND user_id = $2

			FOR UPDATE;
			`,
			workshopID,
			userID,
		).Scan(
			&existingStatus,
		)

	if existingErr != nil &&
		!errors.Is(
			existingErr,
			sql.ErrNoRows,
		) {

		return models.InstitutionalActivity{},
			existingErr
	}

	if existingErr == nil &&
		existingStatus ==
			models.
				InstitutionalEnrollmentStatusConfirmed {

		return models.InstitutionalActivity{},
			ErrInstitutionalWorkshopAlreadyEnrolled
	}

	var confirmedCount int

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT COUNT(*)::integer

		FROM institutional_activity_enrollments

		WHERE activity_id = $1
		  AND status = 'CONFIRMED';
		`,
		workshopID,
	).Scan(
		&confirmedCount,
	)

	if err != nil {
		return models.InstitutionalActivity{},
			err
	}

	if confirmedCount >=
		int(capacity.Int64) {

		return models.InstitutionalActivity{},
			ErrInstitutionalWorkshopFull
	}

	if existingErr == nil {
		// La combinación activity_id/user_id es única.
		//
		// Una inscripción CANCELLED se reactiva en vez de generar una segunda
		// fila, conservando identidad e historial de la relación.
		_, err = tx.ExecContext(
			ctx,
			`
			UPDATE institutional_activity_enrollments

			SET
				status = 'CONFIRMED',
				updated_at = CURRENT_TIMESTAMP

			WHERE activity_id = $1
			  AND user_id = $2
			  AND status = 'CANCELLED';
			`,
			workshopID,
			userID,
		)
	} else {
		_, err = tx.ExecContext(
			ctx,
			`
			INSERT INTO institutional_activity_enrollments (
				activity_id,
				user_id,
				status
			)
			VALUES (
				$1,
				$2,
				'CONFIRMED'
			);
			`,
			workshopID,
			userID,
		)
	}

	if err != nil {
		return models.InstitutionalActivity{},
			err
	}

	if err := tx.Commit(); err != nil {
		return models.InstitutionalActivity{},
			err
	}

	return GetScheduledInstitutionalWorkshopForUser(
		workshopID,
		userID,
	)
}

// ============================================================================
// RETIRO DE TALLER
// ============================================================================

func CancelUserInstitutionalWorkshopEnrollment(
	workshopID int,
	userID int,
) (models.InstitutionalActivity, error) {
	ctx :=
		context.Background()

	tx, err := database.DB.BeginTx(
		ctx,
		&sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
		},
	)

	if err != nil {
		return models.InstitutionalActivity{},
			err
	}

	defer tx.Rollback()

	var activityType string

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT activity_type

		FROM institutional_activities

		WHERE id = $1

		FOR UPDATE;
		`,
		workshopID,
	).Scan(
		&activityType,
	)

	if err != nil {
		if errors.Is(
			err,
			sql.ErrNoRows,
		) {
			return models.InstitutionalActivity{},
				ErrInstitutionalWorkshopNotFound
		}

		return models.InstitutionalActivity{},
			err
	}

	if activityType !=
		models.InstitutionalActivityTypeWorkshop {

		return models.InstitutionalActivity{},
			ErrInstitutionalWorkshopNotFound
	}

	result, err := tx.ExecContext(
		ctx,
		`
		UPDATE institutional_activity_enrollments

		SET
			status = 'CANCELLED',
			updated_at = CURRENT_TIMESTAMP

		WHERE activity_id = $1
		  AND user_id = $2
		  AND status = 'CONFIRMED';
		`,
		workshopID,
		userID,
	)

	if err != nil {
		return models.InstitutionalActivity{},
			err
	}

	affected, err :=
		result.RowsAffected()

	if err != nil {
		return models.InstitutionalActivity{},
			err
	}

	if affected != 1 {
		return models.InstitutionalActivity{},
			ErrInstitutionalWorkshopNotEnrolled
	}

	if err := tx.Commit(); err != nil {
		return models.InstitutionalActivity{},
			err
	}

	return GetScheduledInstitutionalWorkshopForUser(
		workshopID,
		userID,
	)
}
