package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

// ErrInstitutionalScheduleBlocked representa una colisión con un bloqueo
// administrativo duro.
//
// A diferencia de una reserva o de otra actividad institucional, un
// availability_block no genera un conflicto resoluble: impide programar.
var ErrInstitutionalScheduleBlocked = errors.New(
	"la programación institucional se solapa con un bloqueo administrativo",
)

// ============================================================================
// SCANNER COMÚN
// ============================================================================

// rowScanner permite reutilizar el mismo código tanto con QueryRow como con
// Rows. Todas las consultas de actividades deben mantener el mismo orden de
// columnas definido en scanInstitutionalActivity.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanInstitutionalActivity(
	scanner rowScanner,
) (models.InstitutionalActivity, error) {
	var activity models.InstitutionalActivity
	var capacity sql.NullInt64

	err := scanner.Scan(
		&activity.ID,
		&activity.UnitID,
		&activity.ResourceID,
		&activity.ActivityType,
		&activity.Title,
		&activity.Description,
		&activity.Status,
		&activity.RequiresEnrollment,
		&capacity,
		&activity.CreatedByUserID,
		&activity.CreatedAt,
		&activity.UpdatedAt,
		&activity.UnitName,
		&activity.ResourceName,
		&activity.CreatedBy,
		&activity.EnrollmentCount,
	)

	if err != nil {
		return models.InstitutionalActivity{}, err
	}

	if capacity.Valid {
		value := int(capacity.Int64)
		activity.Capacity = &value
	}

	return activity, nil
}

// ============================================================================
// CONSULTA DE ACTIVIDAD
// ============================================================================

// GetInstitutionalActivityByID devuelve una actividad con información
// enriquecida de unidad, recurso, creador, inscripciones y horarios.
//
// Las actividades CANCELLED/COMPLETED se mantienen consultables porque forman
// parte del historial institucional.
func GetInstitutionalActivityByID(
	activityID int,
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

		WHERE activity.id = $1;
		`,
		activityID,
	)

	activity, err := scanInstitutionalActivity(row)
	if err != nil {
		return models.InstitutionalActivity{}, err
	}

	schedules, err := GetInstitutionalActivitySchedules(activity.ID)
	if err != nil {
		return models.InstitutionalActivity{}, err
	}

	activity.Schedules = schedules

	return activity, nil
}

// ============================================================================
// LISTADO POR UNIDAD
// ============================================================================

// GetInstitutionalActivitiesByUnit devuelve la programación perteneciente a
// una unidad institucional.
//
// No filtra por estado porque la pantalla de gestión necesitará visualizar
// también borradores, actividades canceladas y registros completados.
func GetInstitutionalActivitiesByUnit(
	unitID int,
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

		WHERE activity.unit_id = $1

		ORDER BY
			activity.created_at DESC,
			activity.id DESC;
		`,
		unitID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	activities := []models.InstitutionalActivity{}

	for rows.Next() {
		activity, err := scanInstitutionalActivity(rows)
		if err != nil {
			return nil, err
		}

		schedules, err :=
			GetInstitutionalActivitySchedules(activity.ID)

		if err != nil {
			return nil, err
		}

		activity.Schedules = schedules

		activities = append(activities, activity)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return activities, nil
}

// ============================================================================
// CREACIÓN TRANSACCIONAL
// ============================================================================

// CreateInstitutionalActivityWithSchedules crea una actividad y todas sus
// reglas de programación dentro de una única transacción.
//
// Si falla cualquiera de los horarios, no debe quedar una actividad incompleta.
//
// Utilizamos SERIALIZABLE y el mismo advisory lock por recurso utilizado por
// reservations/availability_blocks. Esto deja preparada la operación para la
// posterior detección transaccional de conflictos sin introducir carreras de
// concurrencia entre distintos tipos de ocupación.
func CreateInstitutionalActivityWithSchedules(
	request models.CreateInstitutionalActivityRequest,
	createdByUserID int,
	status string,
) (models.InstitutionalActivity, error) {
	ctx := context.Background()

	tx, err := database.DB.BeginTx(
		ctx,
		&sql.TxOptions{
			Isolation: sql.LevelSerializable,
		},
	)

	if err != nil {
		return models.InstitutionalActivity{}, err
	}

	defer tx.Rollback()

	// ------------------------------------------------------------
	// Serialización por recurso.
	// ------------------------------------------------------------
	//
	// La clave 73001 es la misma familia utilizada por las invariantes de
	// reservations y availability_blocks.
	//
	// El objetivo no es impedir dos actividades simultáneas —eso está
	// permitido— sino asegurar que la detección/registro de conflictos pueda
	// ejecutarse posteriormente sobre una fotografía coherente.

	_, err = tx.ExecContext(
		ctx,
		`
		SELECT pg_advisory_xact_lock(73001, $1);
		`,
		request.ResourceID,
	)

	if err != nil {
		return models.InstitutionalActivity{}, err
	}

	var activityID int

	err = tx.QueryRowContext(
		ctx,
		`
		INSERT INTO institutional_activities (
			unit_id,
			resource_id,
			activity_type,
			title,
			description,
			status,
			requires_enrollment,
			capacity,
			created_by_user_id
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			NULLIF($5, ''),
			$6,
			$7,
			$8,
			$9
		)
		RETURNING id;
		`,
		request.UnitID,
		request.ResourceID,
		request.ActivityType,
		request.Title,
		request.Description,
		status,
		request.RequiresEnrollment,
		request.Capacity,
		createdByUserID,
	).Scan(&activityID)

	if err != nil {
		return models.InstitutionalActivity{}, err
	}

	// ------------------------------------------------------------
	// Horarios.
	// ------------------------------------------------------------

	for _, schedule := range request.Schedules {

		// ------------------------------------------------------------
		// Bloqueos administrativos.
		// ------------------------------------------------------------
		//
		// Esta validación ocurre después de tomar el advisory lock del
		// recurso y dentro de la misma transacción que crea la actividad.
		//
		// De esta forma un availability_block no puede aparecer entre la
		// comprobación y la persistencia de la programación.

		blocked, err :=
			institutionalScheduleHasAvailabilityBlockTx(
				ctx,
				tx,
				request.ResourceID,
				schedule,
			)

		if err != nil {
			return models.InstitutionalActivity{}, err
		}

		if blocked {
			return models.InstitutionalActivity{},
				ErrInstitutionalScheduleBlocked
		}

		_, err = tx.ExecContext(
			ctx,
			`
			INSERT INTO institutional_activity_schedules (
				activity_id,
				schedule_type,
				specific_date,
				day_of_week,
				start_time,
				end_time,
				valid_from,
				valid_to,
				is_active
			)
			VALUES (
				$1,
				$2,
				$3::date,
				$4,
				$5::time,
				$6::time,
				$7::date,
				$8::date,
				true
			);
			`,
			activityID,
			schedule.ScheduleType,
			nullableString(schedule.SpecificDate),
			nullableInt(schedule.DayOfWeek),
			schedule.StartTime,
			schedule.EndTime,
			nullableString(schedule.ValidFrom),
			nullableString(schedule.ValidTo),
		)

		if err != nil {
			return models.InstitutionalActivity{}, err
		}
	}
	// ------------------------------------------------------------
	// Conflictos de programación.
	// ------------------------------------------------------------
	//
	// Una actividad institucional puede coexistir temporalmente con reservas
	// u otras actividades ya existentes.
	//
	// Esa superposición NO rechaza la creación: se registra como un conflicto
	// administrativo de 2..N elementos.
	//
	// La detección ocurre antes del COMMIT y mientras continúa vigente el
	// advisory lock del recurso.

	if status == models.InstitutionalActivityStatusScheduled {
		if err := DetectAndPersistSchedulingConflictsForActivityTx(
			ctx,
			tx,
			activityID,
			request.ResourceID,
		); err != nil {
			return models.InstitutionalActivity{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return models.InstitutionalActivity{}, err
	}

	return GetInstitutionalActivityByID(activityID)
}

// ============================================================================
// HORARIOS
// ============================================================================

// GetInstitutionalActivitySchedules devuelve las reglas estructuradas de una
// actividad.
//
// Las fechas y horas se convierten explícitamente a representación ISO para
// evitar que detalles del driver PostgreSQL se filtren hacia la API.
func GetInstitutionalActivitySchedules(
	activityID int,
) ([]models.InstitutionalActivitySchedule, error) {
	rows, err := database.DB.QueryContext(
		context.Background(),
		`
		SELECT
			id,
			activity_id,
			schedule_type,

			CASE
				WHEN specific_date IS NULL THEN NULL
				ELSE to_char(specific_date, 'YYYY-MM-DD')
			END,

			day_of_week,

			to_char(start_time, 'HH24:MI'),
			to_char(end_time, 'HH24:MI'),

			CASE
				WHEN valid_from IS NULL THEN NULL
				ELSE to_char(valid_from, 'YYYY-MM-DD')
			END,

			CASE
				WHEN valid_to IS NULL THEN NULL
				ELSE to_char(valid_to, 'YYYY-MM-DD')
			END,

			is_active,
			created_at,
			updated_at

		FROM institutional_activity_schedules

		WHERE activity_id = $1

		ORDER BY
			CASE schedule_type
				WHEN 'SINGLE' THEN 0
				ELSE 1
			END,
			specific_date NULLS LAST,
			day_of_week NULLS LAST,
			start_time,
			id;
		`,
		activityID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	schedules := []models.InstitutionalActivitySchedule{}

	for rows.Next() {
		var schedule models.InstitutionalActivitySchedule

		var specificDate sql.NullString
		var dayOfWeek sql.NullInt64
		var validFrom sql.NullString
		var validTo sql.NullString

		err := rows.Scan(
			&schedule.ID,
			&schedule.ActivityID,
			&schedule.ScheduleType,
			&specificDate,
			&dayOfWeek,
			&schedule.StartTime,
			&schedule.EndTime,
			&validFrom,
			&validTo,
			&schedule.IsActive,
			&schedule.CreatedAt,
			&schedule.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if specificDate.Valid {
			value := specificDate.String
			schedule.SpecificDate = &value
		}

		if dayOfWeek.Valid {
			value := int(dayOfWeek.Int64)
			schedule.DayOfWeek = &value
		}

		if validFrom.Valid {
			value := validFrom.String
			schedule.ValidFrom = &value
		}

		if validTo.Valid {
			value := validTo.String
			schedule.ValidTo = &value
		}

		schedules = append(schedules, schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return schedules, nil
}

// ============================================================================
// BLOQUEOS ADMINISTRATIVOS
// ============================================================================

// institutionalScheduleHasAvailabilityBlockTx comprueba si alguna ocurrencia
// generada por una regla de programación institucional intersecta un bloqueo
// administrativo activo.
//
// SINGLE:
//
//	comprueba una fecha concreta.
//
// WEEKLY:
//
//	expande únicamente los días de semana correspondientes dentro de
//	valid_from / valid_to.
//
// La expansión ocurre en PostgreSQL para que esta comprobación participe en
// la misma transacción y advisory lock utilizados al crear la actividad.
func institutionalScheduleHasAvailabilityBlockTx(
	ctx context.Context,
	tx *sql.Tx,
	resourceID int,
	schedule models.CreateInstitutionalScheduleRequest,
) (bool, error) {
	var blocked bool

	switch schedule.ScheduleType {

	case models.InstitutionalScheduleTypeSingle:

		err := tx.QueryRowContext(
			ctx,
			`
			SELECT EXISTS (
				SELECT 1

				FROM availability_blocks block

				WHERE block.resource_id = $1
				  AND block.is_active = true

				  AND tstzrange(
						block.start_time,
						block.end_time,
						'[)'
				  )
				  &&
				  tstzrange(
						(
							($2::date + $3::time)
							AT TIME ZONE 'America/Santiago'
						),
						(
							($2::date + $4::time)
							AT TIME ZONE 'America/Santiago'
						),
						'[)'
				  )
			);
			`,
			resourceID,
			nullableString(schedule.SpecificDate),
			schedule.StartTime,
			schedule.EndTime,
		).Scan(&blocked)

		if err != nil {
			return false, err
		}

		return blocked, nil

	case models.InstitutionalScheduleTypeWeekly:

		err := tx.QueryRowContext(
			ctx,
			`
			SELECT EXISTS (
				SELECT 1

				FROM generate_series(
					$2::date,
					$3::date,
					interval '1 day'
				) occurrence_date

				INNER JOIN availability_blocks block
					ON block.resource_id = $1
				   AND block.is_active = true

				WHERE extract(
					isodow
					FROM occurrence_date
				)::integer = $4

				  AND tstzrange(
						block.start_time,
						block.end_time,
						'[)'
				  )
				  &&
				  tstzrange(
						(
							(
								occurrence_date::date
								+ $5::time
							)
							AT TIME ZONE 'America/Santiago'
						),
						(
							(
								occurrence_date::date
								+ $6::time
							)
							AT TIME ZONE 'America/Santiago'
						),
						'[)'
				  )
			);
			`,
			resourceID,
			nullableString(schedule.ValidFrom),
			nullableString(schedule.ValidTo),
			nullableInt(schedule.DayOfWeek),
			schedule.StartTime,
			schedule.EndTime,
		).Scan(&blocked)

		if err != nil {
			return false, err
		}

		return blocked, nil

	default:
		// Los tipos inválidos son rechazados por el service antes de llegar
		// al repository. Este retorno mantiene la función defensiva.
		return false, nil
	}
}

// ============================================================================
// DISPONIBILIDAD INSTITUCIONAL
// ============================================================================

// GetScheduledInstitutionalActivitiesForAvailability materializa las
// ocurrencias concretas de actividades institucionales que intersectan el
// rango solicitado por el calendario.
//
// La tabla institutional_activity_schedules almacena reglas:
//
//	SINGLE -> una fecha concreta
//	WEEKLY -> una recurrencia
//
// La disponibilidad, en cambio, necesita ocupaciones concretas:
//
//	2026-08-24 10:00-11:30
//	2026-08-31 10:00-11:30
//
// Por eso WEEKLY se expande solamente dentro del rango consultado.
// No materializamos permanentemente todo el semestre.
func GetScheduledInstitutionalActivitiesForAvailability(
	from time.Time,
	to time.Time,
) ([]models.ScheduledActivity, error) {
	rows, err := database.DB.QueryContext(
		context.Background(),
		`
		WITH original_occurrences AS (

			-- ================================================================
			-- SINGLE ORIGINALES
			-- ================================================================
			--
			-- CANCEL y RESCHEDULE eliminan la ocurrencia original.
			-- RESCHEDULE será materializada posteriormente como una nueva
			-- ocurrencia concreta.

			SELECT
				activity.id,
				activity.resource_id,
				activity.title,
				activity.activity_type,
				activity.created_by_user_id,
				resource.name AS resource_name,

				(
					(
						schedule.specific_date
						+ schedule.start_time
					)
					AT TIME ZONE 'America/Santiago'
				) AS occurrence_start,

				(
					(
						schedule.specific_date
						+ schedule.end_time
					)
					AT TIME ZONE 'America/Santiago'
				) AS occurrence_end

			FROM institutional_activities activity

			INNER JOIN institutional_activity_schedules schedule
				ON schedule.activity_id = activity.id

			INNER JOIN resources resource
				ON resource.id = activity.resource_id

			WHERE activity.status = 'SCHEDULED'
			  AND schedule.is_active = true
			  AND schedule.schedule_type = 'SINGLE'

			  AND NOT EXISTS (
					SELECT 1

					FROM institutional_activity_schedule_exceptions exception

					WHERE exception.schedule_id = schedule.id
					  AND exception.original_date =
							schedule.specific_date
			  )


			UNION ALL


			-- ================================================================
			-- WEEKLY ORIGINALES
			-- ================================================================

			SELECT
				activity.id,
				activity.resource_id,
				activity.title,
				activity.activity_type,
				activity.created_by_user_id,
				resource.name AS resource_name,

				(
					(
						occurrence_date::date
						+ schedule.start_time
					)
					AT TIME ZONE 'America/Santiago'
				) AS occurrence_start,

				(
					(
						occurrence_date::date
						+ schedule.end_time
					)
					AT TIME ZONE 'America/Santiago'
				) AS occurrence_end

			FROM institutional_activities activity

			INNER JOIN institutional_activity_schedules schedule
				ON schedule.activity_id = activity.id

			INNER JOIN resources resource
				ON resource.id = activity.resource_id

			CROSS JOIN LATERAL generate_series(

				GREATEST(
					schedule.valid_from,
					(
						$1::timestamptz
						AT TIME ZONE 'America/Santiago'
					)::date
				)::timestamp,

				LEAST(
					schedule.valid_to,
					(
						(
							$2::timestamptz
							- interval '1 microsecond'
						)
						AT TIME ZONE 'America/Santiago'
					)::date
				)::timestamp,

				interval '1 day'

			) occurrence_date

			WHERE activity.status = 'SCHEDULED'
			  AND schedule.is_active = true
			  AND schedule.schedule_type = 'WEEKLY'

			  AND extract(
					isodow
					FROM occurrence_date
			  )::integer = schedule.day_of_week

			  AND NOT EXISTS (
					SELECT 1

					FROM institutional_activity_schedule_exceptions exception

					WHERE exception.schedule_id = schedule.id
					  AND exception.original_date =
							occurrence_date::date
			  )
		),

		rescheduled_occurrences AS (

			-- ================================================================
			-- EXCEPCIONES RESCHEDULE
			-- ================================================================
			--
			-- La ocurrencia original ya fue excluida.
			-- Aquí materializamos únicamente su reemplazo.

			SELECT
				activity.id,
				activity.resource_id,
				activity.title,
				activity.activity_type,
				activity.created_by_user_id,
				resource.name AS resource_name,

				(
					(
						exception.new_date
						+ exception.new_start_time
					)
					AT TIME ZONE 'America/Santiago'
				) AS occurrence_start,

				(
					(
						exception.new_date
						+ exception.new_end_time
					)
					AT TIME ZONE 'America/Santiago'
				) AS occurrence_end

			FROM institutional_activity_schedule_exceptions exception

			INNER JOIN institutional_activities activity
				ON activity.id = exception.activity_id

			INNER JOIN institutional_activity_schedules schedule
				ON schedule.id = exception.schedule_id
			   AND schedule.activity_id = exception.activity_id

			INNER JOIN resources resource
				ON resource.id = activity.resource_id

			WHERE exception.exception_type = 'RESCHEDULE'
			  AND activity.status = 'SCHEDULED'
			  AND schedule.is_active = true
		),

		all_occurrences AS (

			SELECT *
			FROM original_occurrences

			UNION ALL

			SELECT *
			FROM rescheduled_occurrences
		)

		SELECT
			id,
			resource_id,
			title,
			activity_type,
			occurrence_start,
			occurrence_end,
			resource_name,
			created_by_user_id

		FROM all_occurrences

		WHERE occurrence_start < $2
		  AND occurrence_end > $1

		ORDER BY
			occurrence_start,
			resource_id,
			id;
		`,
		from,
		to,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	activities := []models.ScheduledActivity{}

	for rows.Next() {
		var activity models.ScheduledActivity

		if err := rows.Scan(
			&activity.ID,
			&activity.ResourceID,
			&activity.Title,
			&activity.ActivityType,
			&activity.StartTime,
			&activity.EndTime,
			&activity.ResourceName,
			&activity.CreatedByUserID,
		); err != nil {
			return nil, err
		}

		activity.StartTime =
			activity.StartTime.In(
				businessclock.Location(),
			)

		activity.EndTime =
			activity.EndTime.In(
				businessclock.Location(),
			)

		activities = append(
			activities,
			activity,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return activities, nil
}

// ============================================================================
// HELPERS SQL NULL
// ============================================================================

// Los requests usan punteros para distinguir entre:
//
//   campo ausente -> NULL
//   campo presente -> valor real
//
// Convertimos explícitamente los punteros antes de entregarlos a database/sql
// para mantener predecible el comportamiento del driver.

func nullableString(
	value *string,
) any {
	if value == nil {
		return nil
	}

	return *value
}

func nullableInt(
	value *int,
) any {
	if value == nil {
		return nil
	}

	return *value
}
