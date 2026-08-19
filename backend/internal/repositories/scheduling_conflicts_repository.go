package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/schedulingconflicts"
)

// ============================================================================
// DETECCIÓN Y PERSISTENCIA
// ============================================================================

// DetectAndPersistSchedulingConflictsForActivityTx recalcula los componentes
// de conflicto afectados por una actividad institucional recién programada.
//
// IMPORTANTE:
//
// Esta función debe ejecutarse dentro de la MISMA transacción y advisory lock
// utilizados para crear la actividad.
//
// Así evitamos:
//
//	crear actividad
//	       ↓
//	liberar lock
//	       ↓
//	aparecer otra ocupación
//	       ↓
//	detectar sobre un estado inconsistente
//
// Cada fecha concreta se evalúa independientemente. Una actividad WEEKLY puede
// por tanto generar conflictos diferentes en distintas ocurrencias.
func DetectAndPersistSchedulingConflictsForActivityTx(
	ctx context.Context,
	tx *sql.Tx,
	activityID int,
	resourceID int,
) error {
	occurrenceDates, err :=
		getInstitutionalActivityOccurrenceDatesTx(
			ctx,
			tx,
			activityID,
		)

	if err != nil {
		return err
	}

	for _, occurrenceDate := range occurrenceDates {
		occupancies, err :=
			getSchedulingOccupanciesForDateTx(
				ctx,
				tx,
				resourceID,
				occurrenceDate,
			)

		if err != nil {
			return err
		}

		components, err :=
			schedulingconflicts.DetectConnectedComponents(
				occupancies,
			)

		if err != nil {
			return err
		}

		for _, component := range components {

			// Solo recalculamos componentes afectados por la actividad
			// recién creada. Los demás conflictos del recurso no cambian.
			if !componentContainsInstitutionalActivity(
				component,
				activityID,
			) {
				continue
			}

			if err := persistSchedulingConflictComponentTx(
				ctx,
				tx,
				component,
			); err != nil {
				return err
			}
		}
	}

	return nil
}

// ============================================================================
// FECHAS GENERADAS POR UNA ACTIVIDAD
// ============================================================================

// getInstitutionalActivityOccurrenceDatesTx materializa solamente las fechas
// en las que realmente ocurre una actividad.
//
// SINGLE:
//
//	specific_date
//
// WEEKLY:
//
//	todas las fechas dentro de valid_from / valid_to cuyo ISO weekday
//	coincide con day_of_week.
//
// Retornamos YYYY-MM-DD para mantener explícita la semántica de fecha local.
func getInstitutionalActivityOccurrenceDatesTx(
	ctx context.Context,
	tx *sql.Tx,
	activityID int,
) ([]string, error) {
	rows, err := tx.QueryContext(
		ctx,
		`
		SELECT DISTINCT
			to_char(occurrence_date, 'YYYY-MM-DD')

		FROM (
			-- --------------------------------------------------------
			-- SINGLE
			-- --------------------------------------------------------

			SELECT
				schedule.specific_date AS occurrence_date

			FROM institutional_activity_schedules schedule

			WHERE schedule.activity_id = $1
			  AND schedule.is_active = true
			  AND schedule.schedule_type = 'SINGLE'

			UNION ALL

			-- --------------------------------------------------------
			-- WEEKLY
			-- --------------------------------------------------------

			SELECT
				generated_date::date AS occurrence_date

			FROM institutional_activity_schedules schedule

			CROSS JOIN LATERAL generate_series(
				schedule.valid_from,
				schedule.valid_to,
				interval '1 day'
			) generated_date

			WHERE schedule.activity_id = $1
			  AND schedule.is_active = true
			  AND schedule.schedule_type = 'WEEKLY'
			  AND extract(
					isodow
					FROM generated_date
			  )::integer = schedule.day_of_week

		) occurrences

		WHERE occurrence_date IS NOT NULL

		ORDER BY 1;
		`,
		activityID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	dates := []string{}

	for rows.Next() {
		var value string

		if err := rows.Scan(&value); err != nil {
			return nil, err
		}

		dates = append(dates, value)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return dates, nil
}

// ============================================================================
// OCUPACIONES DE UN RECURSO PARA UNA FECHA
// ============================================================================

// getSchedulingOccupanciesForDateTx carga todos los nodos que pueden formar
// parte del grafo de conflictos para una fecha concreta:
//
//   - actividades institucionales SCHEDULED;
//   - reservas PENDING;
//   - reservas CONFIRMED.
//
// availability_blocks NO participan en este grafo:
//
//	availability_block = rechazo duro
//	scheduling conflict = incompatibilidad resoluble
func getSchedulingOccupanciesForDateTx(
	ctx context.Context,
	tx *sql.Tx,
	resourceID int,
	occurrenceDate string,
) ([]schedulingconflicts.Occupancy, error) {
	occupancies := []schedulingconflicts.Occupancy{}

	// ========================================================================
	// ACTIVIDADES INSTITUCIONALES
	// ========================================================================

	rows, err := tx.QueryContext(
		ctx,
		`
		SELECT
			activity.id,
			schedule.id,

			(
				($2::date + schedule.start_time)
				AT TIME ZONE 'America/Santiago'
			) AS occurrence_start,

			(
				($2::date + schedule.end_time)
				AT TIME ZONE 'America/Santiago'
			) AS occurrence_end

		FROM institutional_activities activity

		INNER JOIN institutional_activity_schedules schedule
			ON schedule.activity_id = activity.id

		WHERE activity.resource_id = $1
		  AND activity.status = 'SCHEDULED'
		  AND schedule.is_active = true

		  AND (
				(
					schedule.schedule_type = 'SINGLE'
					AND schedule.specific_date = $2::date
				)

				OR

				(
					schedule.schedule_type = 'WEEKLY'
					AND $2::date BETWEEN
						schedule.valid_from
						AND schedule.valid_to

					AND extract(
						isodow
						FROM $2::date
					)::integer = schedule.day_of_week
				)
		  )

		ORDER BY
			occurrence_start,
			occurrence_end,
			activity.id,
			schedule.id;
		`,
		resourceID,
		occurrenceDate,
	)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			activityID int
			scheduleID int

			start time.Time
			end   time.Time
		)

		if err := rows.Scan(
			&activityID,
			&scheduleID,
			&start,
			&end,
		); err != nil {
			rows.Close()
			return nil, err
		}

		start = start.In(businessclock.Location())
		end = end.In(businessclock.Location())

		activityIDValue := activityID
		scheduleIDValue := scheduleID

		occupancies = append(
			occupancies,
			schedulingconflicts.Occupancy{
				Key: fmt.Sprintf(
					"activity:%d:schedule:%d:%s",
					activityID,
					scheduleID,
					start.Format(time.RFC3339Nano),
				),

				ResourceID: resourceID,

				Kind: schedulingconflicts.
					OccupancyKindInstitutionalActivity,

				InstitutionalActivityID: &activityIDValue,
				ScheduleID:              &scheduleIDValue,

				Start: start,
				End:   end,
			},
		)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}

	rows.Close()

	// ========================================================================
	// RESERVAS
	// ========================================================================
	//
	// Las reservas activas se consideran ocupaciones concretas.
	//
	// OPEN_USE también puede aparecer aquí. Una actividad institucional puede
	// requerir posteriormente resolución administrativa frente a usos ya
	// existentes, aunque nuevas reservas deberán ser bloqueadas por la
	// programación institucional una vez integremos disponibilidad.

	dayStartExpression := `
		($2::date::timestamp AT TIME ZONE 'America/Santiago')
	`

	dayEndExpression := `
		(($2::date + 1)::timestamp AT TIME ZONE 'America/Santiago')
	`

	rows, err = tx.QueryContext(
		ctx,
		`
		SELECT
			reservation.id,
			reservation.start_time,
			reservation.end_time

		FROM reservations reservation

		WHERE reservation.resource_id = $1
		  AND reservation.status IN (
				'PENDING',
				'CONFIRMED'
		  )

		  AND reservation.start_time < `+dayEndExpression+`
		  AND reservation.end_time > `+dayStartExpression+`

		ORDER BY
			reservation.start_time,
			reservation.end_time,
			reservation.id;
		`,
		resourceID,
		occurrenceDate,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			reservationID int

			start time.Time
			end   time.Time
		)

		if err := rows.Scan(
			&reservationID,
			&start,
			&end,
		); err != nil {
			return nil, err
		}

		start = start.In(businessclock.Location())
		end = end.In(businessclock.Location())

		reservationIDValue := reservationID

		occupancies = append(
			occupancies,
			schedulingconflicts.Occupancy{
				Key: fmt.Sprintf(
					"reservation:%d",
					reservationID,
				),

				ResourceID: resourceID,

				Kind: schedulingconflicts.
					OccupancyKindReservation,

				ReservationID: &reservationIDValue,

				Start: start,
				End:   end,
			},
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return occupancies, nil
}

// ============================================================================
// COMPONENTE AFECTADO
// ============================================================================

func componentContainsInstitutionalActivity(
	component schedulingconflicts.Component,
	activityID int,
) bool {
	for _, item := range component.Items {
		if item.Kind !=
			schedulingconflicts.
				OccupancyKindInstitutionalActivity {
			continue
		}

		if item.InstitutionalActivityID != nil &&
			*item.InstitutionalActivityID == activityID {
			return true
		}
	}

	return false
}

// ============================================================================
// PERSISTENCIA / FUSIÓN DEL COMPONENTE
// ============================================================================

// persistSchedulingConflictComponentTx garantiza que un componente conectado
// quede representado por UN solo scheduling_conflict.
//
// Si la nueva actividad conecta dos conflictos existentes:
//
//	[A,B]     [C,D]
//	    \       /
//	      NUEVA
//
// los grupos se fusionan.
//
// No modelamos por tanto conflictos como pares independientes.
func persistSchedulingConflictComponentTx(
	ctx context.Context,
	tx *sql.Tx,
	component schedulingconflicts.Component,
) error {
	if len(component.Items) < 2 {
		return nil
	}

	existingConflictIDs, err :=
		findPendingConflictsForComponentTx(
			ctx,
			tx,
			component,
		)

	if err != nil {
		return err
	}

	var conflictID int

	switch len(existingConflictIDs) {

	case 0:

		err := tx.QueryRowContext(
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
			RETURNING id;
			`,
			component.ResourceID,
		).Scan(&conflictID)

		if err != nil {
			return err
		}

	default:

		// El menor ID se conserva como identidad canónica del componente.
		sort.Ints(existingConflictIDs)

		conflictID = existingConflictIDs[0]

		for _, otherConflictID := range existingConflictIDs[1:] {

			if err := mergeSchedulingConflictsTx(
				ctx,
				tx,
				conflictID,
				otherConflictID,
			); err != nil {
				return err
			}
		}
	}

	// Agregamos las ocupaciones detectadas que todavía no pertenezcan al
	// conflicto canónico.
	for _, item := range component.Items {
		if err := insertSchedulingConflictItemIfMissingTx(
			ctx,
			tx,
			conflictID,
			item,
		); err != nil {
			return err
		}
	}

	// El trigger updated_at mantiene la fecha real de última modificación.
	_, err = tx.ExecContext(
		ctx,
		`
		UPDATE scheduling_conflicts
		SET updated_at = CURRENT_TIMESTAMP
		WHERE id = $1;
		`,
		conflictID,
	)

	return err
}

// ============================================================================
// BÚSQUEDA DE CONFLICTOS EXISTENTES
// ============================================================================

// findPendingConflictsForComponentTx busca cualquier conflicto pendiente que
// comparta al menos una ocurrencia con el componente calculado.
//
// Compartir un nodo significa que los grupos pertenecen al mismo componente
// conectado y deben consolidarse.
func findPendingConflictsForComponentTx(
	ctx context.Context,
	tx *sql.Tx,
	component schedulingconflicts.Component,
) ([]int, error) {
	conflictIDs := map[int]struct{}{}

	for _, item := range component.Items {
		var (
			rows *sql.Rows
			err  error
		)

		switch item.Kind {

		case schedulingconflicts.
			OccupancyKindInstitutionalActivity:

			rows, err = tx.QueryContext(
				ctx,
				`
				SELECT DISTINCT conflict.id

				FROM scheduling_conflicts conflict

				INNER JOIN scheduling_conflict_items item
					ON item.conflict_id = conflict.id

				WHERE conflict.resource_id = $1
				  AND conflict.status = 'PENDING'

				  AND item.institutional_activity_id = $2
				  AND item.schedule_id = $3
				  AND item.occurrence_start = $4;
				`,
				component.ResourceID,
				*item.InstitutionalActivityID,
				*item.ScheduleID,
				item.Start,
			)

		case schedulingconflicts.
			OccupancyKindReservation:

			rows, err = tx.QueryContext(
				ctx,
				`
				SELECT DISTINCT conflict.id

				FROM scheduling_conflicts conflict

				INNER JOIN scheduling_conflict_items item
					ON item.conflict_id = conflict.id

				WHERE conflict.resource_id = $1
				  AND conflict.status = 'PENDING'
				  AND item.reservation_id = $2;
				`,
				component.ResourceID,
				*item.ReservationID,
			)

		default:
			continue
		}

		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var conflictID int

			if err := rows.Scan(
				&conflictID,
			); err != nil {
				rows.Close()
				return nil, err
			}

			conflictIDs[conflictID] =
				struct{}{}
		}

		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}

		rows.Close()
	}

	result := make(
		[]int,
		0,
		len(conflictIDs),
	)

	for conflictID := range conflictIDs {
		result = append(
			result,
			conflictID,
		)
	}

	return result, nil
}

// ============================================================================
// FUSIÓN DE CONFLICTOS
// ============================================================================

// mergeSchedulingConflictsTx mueve todos los elementos de sourceConflictID al
// conflicto canónico y elimina posteriormente el grupo redundante.
//
// Los snapshots históricos y cualquier decisión parcial ya registrada se
// conservan.
func mergeSchedulingConflictsTx(
	ctx context.Context,
	tx *sql.Tx,
	targetConflictID int,
	sourceConflictID int,
) error {
	if targetConflictID == sourceConflictID {
		return nil
	}

	rows, err := tx.QueryContext(
		ctx,
		`
		SELECT
			institutional_activity_id,
			schedule_id,
			reservation_id,

			occurrence_start,
			occurrence_end,

			resolution,
			resolution_source,

			resolved_by_user_id,
			resolved_at,
			resolution_note

		FROM scheduling_conflict_items

		WHERE conflict_id = $1

		ORDER BY id;
		`,
		sourceConflictID,
	)

	if err != nil {
		return err
	}

	type storedConflictItem struct {
		activityID    sql.NullInt64
		scheduleID    sql.NullInt64
		reservationID sql.NullInt64

		start time.Time
		end   time.Time

		resolution       string
		resolutionSource sql.NullString

		resolvedBy sql.NullInt64
		resolvedAt sql.NullTime
		note       sql.NullString
	}

	items := []storedConflictItem{}

	for rows.Next() {
		var item storedConflictItem

		if err := rows.Scan(
			&item.activityID,
			&item.scheduleID,
			&item.reservationID,
			&item.start,
			&item.end,
			&item.resolution,
			&item.resolutionSource,
			&item.resolvedBy,
			&item.resolvedAt,
			&item.note,
		); err != nil {
			rows.Close()
			return err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}

	rows.Close()

	for _, item := range items {
		_, err := tx.ExecContext(
			ctx,
			`
			INSERT INTO scheduling_conflict_items (
				conflict_id,

				institutional_activity_id,
				schedule_id,
				reservation_id,

				occurrence_start,
				occurrence_end,

				resolution,
				resolution_source,

				resolved_by_user_id,
				resolved_at,
				resolution_note
			)

			SELECT
				$1,
				$2,
				$3,
				$4,
				$5,
				$6,
				$7,
				$8,
				$9,
				$10,
				$11

			WHERE NOT EXISTS (
				SELECT 1

				FROM scheduling_conflict_items existing

				WHERE existing.conflict_id = $1

				  AND (
						(
							$2::integer IS NOT NULL
							AND existing.institutional_activity_id = $2
							AND existing.schedule_id = $3
							AND existing.occurrence_start = $5
						)

						OR

						(
							$4::integer IS NOT NULL
							AND existing.reservation_id = $4
						)
				  )
			);
			`,
			targetConflictID,

			nullInt64Value(item.activityID),
			nullInt64Value(item.scheduleID),
			nullInt64Value(item.reservationID),

			item.start,
			item.end,

			item.resolution,
			nullStringValue(item.resolutionSource),

			nullInt64Value(item.resolvedBy),
			nullTimeValue(item.resolvedAt),
			nullStringValue(item.note),
		)

		if err != nil {
			return err
		}
	}

	// ON DELETE CASCADE elimina los items restantes del grupo fuente.
	_, err = tx.ExecContext(
		ctx,
		`
		DELETE FROM scheduling_conflicts
		WHERE id = $1;
		`,
		sourceConflictID,
	)

	return err
}

// ============================================================================
// INSERCIÓN DE ITEMS
// ============================================================================

func insertSchedulingConflictItemIfMissingTx(
	ctx context.Context,
	tx *sql.Tx,
	conflictID int,
	item schedulingconflicts.Occupancy,
) error {
	switch item.Kind {

	case schedulingconflicts.
		OccupancyKindInstitutionalActivity:

		_, err := tx.ExecContext(
			ctx,
			`
			INSERT INTO scheduling_conflict_items (
				conflict_id,

				institutional_activity_id,
				schedule_id,

				occurrence_start,
				occurrence_end,

				resolution
			)

			SELECT
				$1,
				$2,
				$3,
				$4,
				$5,
				'PENDING'

			WHERE NOT EXISTS (
				SELECT 1

				FROM scheduling_conflict_items existing

				WHERE existing.conflict_id = $1
				  AND existing.institutional_activity_id = $2
				  AND existing.schedule_id = $3
				  AND existing.occurrence_start = $4
			);
			`,
			conflictID,
			*item.InstitutionalActivityID,
			*item.ScheduleID,
			item.Start,
			item.End,
		)

		return err

	case schedulingconflicts.
		OccupancyKindReservation:

		_, err := tx.ExecContext(
			ctx,
			`
			INSERT INTO scheduling_conflict_items (
				conflict_id,
				reservation_id,

				occurrence_start,
				occurrence_end,

				resolution
			)

			SELECT
				$1,
				$2,
				$3,
				$4,
				'PENDING'

			WHERE NOT EXISTS (
				SELECT 1

				FROM scheduling_conflict_items existing

				WHERE existing.conflict_id = $1
				  AND existing.reservation_id = $2
			);
			`,
			conflictID,
			*item.ReservationID,
			item.Start,
			item.End,
		)

		return err

	default:
		return schedulingconflicts.ErrInvalidOccupancy
	}
}

// ============================================================================
// HELPERS NULL
// ============================================================================

func nullInt64Value(
	value sql.NullInt64,
) any {
	if !value.Valid {
		return nil
	}

	return value.Int64
}

func nullStringValue(
	value sql.NullString,
) any {
	if !value.Valid {
		return nil
	}

	return value.String
}

func nullTimeValue(
	value sql.NullTime,
) any {
	if !value.Valid {
		return nil
	}

	return value.Time
}
