package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/schedulingconflicts"
)

// ============================================================================
// ERRORES DE RESOLUCIÓN
// ============================================================================

var (
	ErrSchedulingConflictNotFound = errors.New(
		"conflicto de programación no encontrado",
	)

	ErrSchedulingConflictResolved = errors.New(
		"el conflicto de programación ya fue resuelto",
	)

	ErrSchedulingConflictItemNotFound = errors.New(
		"elemento del conflicto no encontrado",
	)

	ErrSchedulingConflictItemResolved = errors.New(
		"el elemento del conflicto ya fue resuelto",
	)

	ErrSchedulingResolutionInvalidPlan = errors.New(
		"el plan de resolución mantiene ocupaciones incompatibles",
	)

	ErrSchedulingRescheduleReservation = errors.New(
		"una reserva no puede reprogramarse mediante programación institucional",
	)

	ErrSchedulingOccurrenceAlreadyAdjusted = errors.New(
		"la ocurrencia institucional ya fue modificada previamente",
	)

	ErrSchedulingRescheduleBlocked = errors.New(
		"el nuevo horario se encuentra bloqueado administrativamente",
	)
)

// ============================================================================
// CONSULTA DE CONFLICTO
// ============================================================================

type schedulingConflictQueryer interface {
	QueryRowContext(
		context.Context,
		string,
		...any,
	) *sql.Row

	QueryContext(
		context.Context,
		string,
		...any,
	) (*sql.Rows, error)
}

func GetSchedulingConflictByID(
	conflictID int,
) (models.SchedulingConflict, error) {
	return getSchedulingConflictByID(
		context.Background(),
		database.DB,
		conflictID,
	)
}

func getSchedulingConflictByID(
	ctx context.Context,
	queryer schedulingConflictQueryer,
	conflictID int,
) (models.SchedulingConflict, error) {
	var (
		conflict models.SchedulingConflict

		resolutionSource sql.NullString
		resolvedBy       sql.NullInt64
		resolvedAt       sql.NullTime
	)

	err := queryer.QueryRowContext(
		ctx,
		`
		SELECT
			conflict.id,
			conflict.resource_id,
			conflict.status,
			conflict.resolution_source,
			COALESCE(conflict.resolution_summary, ''),
			conflict.resolved_by_user_id,
			conflict.resolved_at,
			conflict.created_at,
			conflict.updated_at,
			resource.name

		FROM scheduling_conflicts conflict

		INNER JOIN resources resource
			ON resource.id = conflict.resource_id

		WHERE conflict.id = $1
		`,
		conflictID,
	).Scan(
		&conflict.ID,
		&conflict.ResourceID,
		&conflict.Status,
		&resolutionSource,
		&conflict.ResolutionSummary,
		&resolvedBy,
		&resolvedAt,
		&conflict.CreatedAt,
		&conflict.UpdatedAt,
		&conflict.ResourceName,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.SchedulingConflict{},
				ErrSchedulingConflictNotFound
		}

		return models.SchedulingConflict{}, err
	}

	if resolutionSource.Valid {
		value := resolutionSource.String
		conflict.ResolutionSource = &value
	}

	if resolvedBy.Valid {
		value := int(resolvedBy.Int64)
		conflict.ResolvedByUserID = &value
	}

	if resolvedAt.Valid {
		value := resolvedAt.Time.In(
			businessclock.Location(),
		)

		conflict.ResolvedAt = &value
	}

	conflict.CreatedAt =
		conflict.CreatedAt.In(
			businessclock.Location(),
		)

	conflict.UpdatedAt =
		conflict.UpdatedAt.In(
			businessclock.Location(),
		)

	rows, err := queryer.QueryContext(
		ctx,
		`
		SELECT
			item.id,
			item.conflict_id,
			item.institutional_activity_id,
			item.schedule_id,
			item.reservation_id,
			item.occurrence_start,
			item.occurrence_end,
			item.resolution,
			item.resolution_source,
			item.resolved_by_user_id,
			item.resolved_at,
			COALESCE(item.resolution_note, ''),
			item.created_at,
			item.updated_at,

			COALESCE(
				activity.title,
				'Reserva #' || reservation.id::text
			) AS title,

			COALESCE(unit.name, '') AS unit_name

		FROM scheduling_conflict_items item

		LEFT JOIN institutional_activities activity
			ON activity.id =
				item.institutional_activity_id

		LEFT JOIN institutional_units unit
			ON unit.id = activity.unit_id

		LEFT JOIN reservations reservation
			ON reservation.id =
				item.reservation_id

		WHERE item.conflict_id = $1

		ORDER BY
			item.occurrence_start,
			item.id
		`,
		conflictID,
	)

	if err != nil {
		return models.SchedulingConflict{}, err
	}

	defer rows.Close()

	conflict.Items =
		[]models.SchedulingConflictItem{}

	for rows.Next() {
		var (
			item models.SchedulingConflictItem

			activityID    sql.NullInt64
			scheduleID    sql.NullInt64
			reservationID sql.NullInt64

			itemResolutionSource sql.NullString
			itemResolvedBy       sql.NullInt64
			itemResolvedAt       sql.NullTime
		)

		err := rows.Scan(
			&item.ID,
			&item.ConflictID,
			&activityID,
			&scheduleID,
			&reservationID,
			&item.OccurrenceStart,
			&item.OccurrenceEnd,
			&item.Resolution,
			&itemResolutionSource,
			&itemResolvedBy,
			&itemResolvedAt,
			&item.ResolutionNote,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.Title,
			&item.UnitName,
		)

		if err != nil {
			return models.SchedulingConflict{},
				err
		}

		if activityID.Valid {
			value := int(activityID.Int64)

			item.InstitutionalActivityID =
				&value

			item.EntityType =
				models.
					SchedulingConflictItemTypeInstitutionalActivity
		}

		if scheduleID.Valid {
			value := int(scheduleID.Int64)
			item.ScheduleID = &value
		}

		if reservationID.Valid {
			value := int(reservationID.Int64)

			item.ReservationID =
				&value

			item.EntityType =
				models.
					SchedulingConflictItemTypeReservation
		}

		if itemResolutionSource.Valid {
			value :=
				itemResolutionSource.String

			item.ResolutionSource =
				&value
		}

		if itemResolvedBy.Valid {
			value :=
				int(itemResolvedBy.Int64)

			item.ResolvedByUserID =
				&value
		}

		if itemResolvedAt.Valid {
			value :=
				itemResolvedAt.Time.In(
					businessclock.Location(),
				)

			item.ResolvedAt =
				&value
		}

		item.OccurrenceStart =
			item.OccurrenceStart.In(
				businessclock.Location(),
			)

		item.OccurrenceEnd =
			item.OccurrenceEnd.In(
				businessclock.Location(),
			)

		item.CreatedAt =
			item.CreatedAt.In(
				businessclock.Location(),
			)

		item.UpdatedAt =
			item.UpdatedAt.In(
				businessclock.Location(),
			)

		conflict.Items = append(
			conflict.Items,
			item,
		)
	}

	if err := rows.Err(); err != nil {
		return models.SchedulingConflict{}, err
	}

	return conflict, nil
}

// ============================================================================
// RESOLUCIÓN TRANSACCIONAL DE UN ELEMENTO
// ============================================================================

func ResolveSchedulingConflictItem(
	conflictID int,
	itemID int,
	resolvedByUserID int,
	request models.ResolveSchedulingConflictItemRequest,
) (models.SchedulingConflict, error) {
	ctx := context.Background()

	tx, err := database.DB.BeginTx(
		ctx,
		&sql.TxOptions{
			Isolation: sql.LevelSerializable,
		},
	)

	if err != nil {
		return models.SchedulingConflict{}, err
	}

	defer tx.Rollback()

	// ========================================================================
	// CONFLICTO / RESOURCE LOCK
	// ========================================================================

	var (
		resourceID int
		status     string
	)

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT
			resource_id,
			status

		FROM scheduling_conflicts

		WHERE id = $1
		`,
		conflictID,
	).Scan(
		&resourceID,
		&status,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.SchedulingConflict{},
				ErrSchedulingConflictNotFound
		}

		return models.SchedulingConflict{}, err
	}

	// Mantenemos el mismo lock lógico utilizado por reservas,
	// availability_blocks y programación institucional.
	//
	// Toda operación que cambie ocupación del recurso se serializa mediante
	// esta clave antes de aplicar side effects.
	if _, err := tx.ExecContext(
		ctx,
		`
		SELECT pg_advisory_xact_lock(
			73001,
			$1
		)
		`,
		resourceID,
	); err != nil {
		return models.SchedulingConflict{}, err
	}

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT status

		FROM scheduling_conflicts

		WHERE id = $1
		  AND resource_id = $2

		FOR UPDATE
		`,
		conflictID,
		resourceID,
	).Scan(&status)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.SchedulingConflict{},
				ErrSchedulingConflictNotFound
		}

		return models.SchedulingConflict{}, err
	}

	if status !=
		models.SchedulingConflictStatusPending {

		return models.SchedulingConflict{},
			ErrSchedulingConflictResolved
	}

	// ========================================================================
	// ELEMENTO
	// ========================================================================

	var (
		itemResolution string

		activityID    sql.NullInt64
		scheduleID    sql.NullInt64
		reservationID sql.NullInt64

		occurrenceStart time.Time
		occurrenceEnd   time.Time
	)

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT
			resolution,
			institutional_activity_id,
			schedule_id,
			reservation_id,
			occurrence_start,
			occurrence_end

		FROM scheduling_conflict_items

		WHERE id = $1
		  AND conflict_id = $2

		FOR UPDATE
		`,
		itemID,
		conflictID,
	).Scan(
		&itemResolution,
		&activityID,
		&scheduleID,
		&reservationID,
		&occurrenceStart,
		&occurrenceEnd,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.SchedulingConflict{},
				ErrSchedulingConflictItemNotFound
		}

		return models.SchedulingConflict{}, err
	}

	if itemResolution !=
		models.SchedulingItemResolutionPending {

		return models.SchedulingConflict{},
			ErrSchedulingConflictItemResolved
	}

	occurrenceStart =
		occurrenceStart.In(
			businessclock.Location(),
		)

	occurrenceEnd =
		occurrenceEnd.In(
			businessclock.Location(),
		)

	// ========================================================================
	// SIDE EFFECT
	// ========================================================================

	switch request.Resolution {

	case models.SchedulingItemResolutionKeep,
		models.SchedulingItemResolutionAllow:

		// KEEP y ALLOW conservan la ocupación.
		// Su coherencia contra otros elementos resueltos se valida más abajo.

	case models.SchedulingItemResolutionCancel:

		if activityID.Valid &&
			scheduleID.Valid {

			if err := ensureOccurrenceHasNoPreviousAdjustmentTx(
				ctx,
				tx,
				int(scheduleID.Int64),
				occurrenceStart,
			); err != nil {
				return models.SchedulingConflict{},
					err
			}

			_, err = tx.ExecContext(
				ctx,
				`
				INSERT INTO institutional_activity_schedule_exceptions (
					activity_id,
					schedule_id,
					original_date,
					exception_type,
					created_by_user_id,
					reason
				)
				VALUES (
					$1,
					$2,
					$3::date,
					'CANCEL',
					$4,
					$5
				)
				`,
				int(activityID.Int64),
				int(scheduleID.Int64),
				occurrenceStart.Format(
					"2006-01-02",
				),
				resolvedByUserID,
				request.ResolutionNote,
			)

			if err != nil {
				return models.SchedulingConflict{},
					err
			}

		} else if reservationID.Valid {

			result, err := tx.ExecContext(
				ctx,
				`
				UPDATE reservations

				SET
					status = 'CANCELLED',
					cancellation_reason = $2

				WHERE id = $1
				  AND status IN (
						'PENDING',
						'CONFIRMED'
				  )
				`,
				int(reservationID.Int64),
				request.ResolutionNote,
			)

			if err != nil {
				return models.SchedulingConflict{},
					err
			}

			affected, err :=
				result.RowsAffected()

			if err != nil {
				return models.SchedulingConflict{},
					err
			}

			if affected != 1 {
				return models.SchedulingConflict{},
					fmt.Errorf(
						"la reserva del conflicto ya no se encuentra activa",
					)
			}
		}

	case models.SchedulingItemResolutionReschedule:

		if !activityID.Valid ||
			!scheduleID.Valid {

			return models.SchedulingConflict{},
				ErrSchedulingRescheduleReservation
		}

		if err := ensureOccurrenceHasNoPreviousAdjustmentTx(
			ctx,
			tx,
			int(scheduleID.Int64),
			occurrenceStart,
		); err != nil {
			return models.SchedulingConflict{},
				err
		}

		if err := validateRescheduleHardBlocksTx(
			ctx,
			tx,
			resourceID,
			request,
		); err != nil {
			return models.SchedulingConflict{},
				err
		}

		_, err = tx.ExecContext(
			ctx,
			`
			INSERT INTO institutional_activity_schedule_exceptions (
				activity_id,
				schedule_id,
				original_date,
				exception_type,
				new_date,
				new_start_time,
				new_end_time,
				created_by_user_id,
				reason
			)
			VALUES (
				$1,
				$2,
				$3::date,
				'RESCHEDULE',
				$4::date,
				$5::time,
				$6::time,
				$7,
				$8
			)
			`,
			int(activityID.Int64),
			int(scheduleID.Int64),
			occurrenceStart.Format(
				"2006-01-02",
			),
			*request.NewDate,
			*request.NewStartTime,
			*request.NewEndTime,
			resolvedByUserID,
			request.ResolutionNote,
		)

		if err != nil {
			return models.SchedulingConflict{},
				err
		}

	default:

		return models.SchedulingConflict{},
			fmt.Errorf(
				"resolución administrativa no soportada: %s",
				request.Resolution,
			)
	}

	// ========================================================================
	// REGISTRO INMUTABLE DE LA DECISIÓN
	// ========================================================================

	_, err = tx.ExecContext(
		ctx,
		`
		UPDATE scheduling_conflict_items

		SET
			resolution = $2,
			resolution_source = 'MANUAL',
			resolved_by_user_id = $3,
			resolved_at = CURRENT_TIMESTAMP,
			resolution_note = $4

		WHERE id = $1
		  AND conflict_id = $5
		  AND resolution = 'PENDING'
		`,
		itemID,
		request.Resolution,
		resolvedByUserID,
		request.ResolutionNote,
		conflictID,
	)

	if err != nil {
		return models.SchedulingConflict{}, err
	}

	// ========================================================================
	// COHERENCIA DE DECISIONES YA TOMADAS
	// ========================================================================
	//
	// Toda pareja de ocupaciones que permanezca físicamente activa y continúe
	// solapándose debe haber sido autorizada explícitamente:
	//
	//     ALLOW + ALLOW
	//
	// KEEP significa que la ocupación se conserva porque sus competidores
	// incompatibles serán CANCEL o RESCHEDULE.
	//
	// Por eso:
	//
	//     KEEP + KEEP   -> inválido si se solapan
	//     KEEP + ALLOW  -> inválido si se solapan
	//     ALLOW + ALLOW -> válido

	if err := validateResolvedSurvivorPairsTx(
		ctx,
		tx,
		conflictID,
	); err != nil {
		return models.SchedulingConflict{},
			err
	}

	// ========================================================================
	// RESCHEDULE -> NUEVA DETECCIÓN
	// ========================================================================
	//
	// El conflicto actual se excluye como grupo candidato.
	//
	// Si el nuevo destino choca con otra ocupación, se crea un conflicto nuevo
	// en vez de contaminar/reabrir el conflicto histórico que estamos
	// resolviendo.

	if request.Resolution ==
		models.SchedulingItemResolutionReschedule {

		if err := detectAndPersistSchedulingConflictsExcludingTx(
			ctx,
			tx,
			int(activityID.Int64),
			resourceID,
			conflictID,
		); err != nil {
			return models.SchedulingConflict{},
				err
		}
	}

	// ========================================================================
	// CIERRE AUTOMÁTICO
	// ========================================================================

	var pendingCount int

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT COUNT(*)

		FROM scheduling_conflict_items

		WHERE conflict_id = $1
		  AND resolution = 'PENDING'
		`,
		conflictID,
	).Scan(&pendingCount)

	if err != nil {
		return models.SchedulingConflict{}, err
	}

	if pendingCount == 0 {
		_, err = tx.ExecContext(
			ctx,
			`
			UPDATE scheduling_conflicts

			SET
				status = 'RESOLVED',
				resolution_source = 'MANUAL',
				resolution_summary =
					'Resolución administrativa completada',
				resolved_by_user_id = $2,
				resolved_at = CURRENT_TIMESTAMP

			WHERE id = $1
			  AND status = 'PENDING'
			`,
			conflictID,
			resolvedByUserID,
		)

		if err != nil {
			return models.SchedulingConflict{},
				err
		}
	}

	if err := tx.Commit(); err != nil {
		return models.SchedulingConflict{}, err
	}

	return GetSchedulingConflictByID(
		conflictID,
	)
}

// ============================================================================
// VALIDACIÓN DE OCURRENCIA YA MODIFICADA
// ============================================================================
//
// MVP2 mantiene una sola excepción por ocurrencia original.
//
// Si una ocurrencia ya fue RESCHEDULE y aparece en un nuevo conflicto,
// mantenemos su historia inmutable.
//
// En ese segundo conflicto podrá resolverse mediante KEEP / ALLOW o mediante
// decisiones sobre las otras ocupaciones.
//
// Una cadena arbitraria de reprogramaciones queda preparada como evolución
// posterior del modelo de excepciones.

func ensureOccurrenceHasNoPreviousAdjustmentTx(
	ctx context.Context,
	tx *sql.Tx,
	scheduleID int,
	occurrenceStart time.Time,
) error {
	var exists bool

	err := tx.QueryRowContext(
		ctx,
		`
		SELECT EXISTS (

			SELECT 1

			FROM institutional_activity_schedule_exceptions exception

			WHERE exception.schedule_id = $1

			  AND (
					exception.original_date = $2::date

					OR

					(
						exception.exception_type = 'RESCHEDULE'

						AND (
							(
								exception.new_date
								+ exception.new_start_time
							)
							AT TIME ZONE 'America/Santiago'
						) = $3
					)
			  )
		)
		`,
		scheduleID,
		occurrenceStart.Format(
			"2006-01-02",
		),
		occurrenceStart,
	).Scan(&exists)

	if err != nil {
		return err
	}

	if exists {
		return ErrSchedulingOccurrenceAlreadyAdjusted
	}

	return nil
}

// ============================================================================
// HARD BLOCK PARA RESCHEDULE
// ============================================================================

func validateRescheduleHardBlocksTx(
	ctx context.Context,
	tx *sql.Tx,
	resourceID int,
	request models.ResolveSchedulingConflictItemRequest,
) error {
	var blocked bool

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
						(
							$2::date
							+ $3::time
						)
						AT TIME ZONE 'America/Santiago'
					),
					(
						(
							$2::date
							+ $4::time
						)
						AT TIME ZONE 'America/Santiago'
					),
					'[)'
			  )
		)
		`,
		resourceID,
		*request.NewDate,
		*request.NewStartTime,
		*request.NewEndTime,
	).Scan(&blocked)

	if err != nil {
		return err
	}

	if blocked {
		return ErrSchedulingRescheduleBlocked
	}

	return nil
}

// ============================================================================
// VALIDACIÓN DEL PLAN PARCIAL / FINAL
// ============================================================================

func validateResolvedSurvivorPairsTx(
	ctx context.Context,
	tx *sql.Tx,
	conflictID int,
) error {
	rows, err := tx.QueryContext(
		ctx,
		`
		SELECT
			id,
			resolution,
			institutional_activity_id,
			reservation_id,
			occurrence_start,
			occurrence_end

		FROM scheduling_conflict_items

		WHERE conflict_id = $1
		  AND resolution IN (
				'KEEP',
				'ALLOW'
		  )

		ORDER BY
			occurrence_start,
			id
		`,
		conflictID,
	)

	if err != nil {
		return err
	}

	defer rows.Close()

	type survivor struct {
		id         int
		resolution string

		activityID    sql.NullInt64
		reservationID sql.NullInt64

		start time.Time
		end   time.Time
	}

	survivors := []survivor{}

	for rows.Next() {
		var item survivor

		if err := rows.Scan(
			&item.id,
			&item.resolution,
			&item.activityID,
			&item.reservationID,
			&item.start,
			&item.end,
		); err != nil {

			return err
		}

		survivors = append(
			survivors,
			item,
		)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	for i := 0; i < len(survivors); i++ {

		for j := i + 1; j < len(survivors); j++ {

			left := survivors[i]
			right := survivors[j]

			overlaps :=
				left.start.Before(
					right.end,
				) &&
					right.start.Before(
						left.end,
					)

			if !overlaps {
				continue
			}

			// ================================================================
			// RESERVA ↔ RESERVA
			// ================================================================
			//
			// No constituye una incompatibilidad de scheduling.
			//
			// Esto permite, por ejemplo, conservar mediante KEEP dos usos
			// OPEN_USE compatibles aun cuando ambos hayan participado en un
			// conflicto debido a una actividad institucional común.

			if left.reservationID.Valid &&
				right.reservationID.Valid {

				continue
			}

			// ================================================================
			// ACTIVIDAD ↔ ACTIVIDAD / ACTIVIDAD ↔ RESERVA
			// ================================================================
			//
			// Si ambas ocupaciones sobreviven físicamente solapadas, la
			// coexistencia debe estar autorizada explícitamente.

			if left.resolution ==
				models.SchedulingItemResolutionAllow &&
				right.resolution ==
					models.SchedulingItemResolutionAllow {

				continue
			}

			return fmt.Errorf(
				"%w: items %d y %d continúan siendo incompatibles; ambos deben usar ALLOW o uno debe CANCEL/RESCHEDULE",
				ErrSchedulingResolutionInvalidPlan,
				left.id,
				right.id,
			)
		}
	}

	return nil
}

// ============================================================================
// NUEVA DETECCIÓN EXCLUYENDO EL CONFLICTO DE ORIGEN
// ============================================================================

func detectAndPersistSchedulingConflictsExcludingTx(
	ctx context.Context,
	tx *sql.Tx,
	activityID int,
	resourceID int,
	excludedConflictID int,
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
			detectSchedulingConflictComponentsTx(ctx, tx, occupancies)

		if err != nil {
			return err
		}

		for _, component := range components {

			if !componentContainsInstitutionalActivity(
				component,
				activityID,
			) {
				continue
			}

			if err :=
				persistSchedulingConflictComponentExcludingTx(
					ctx,
					tx,
					component,
					excludedConflictID,
				); err != nil {

				return err
			}
		}
	}

	return nil
}

func persistSchedulingConflictComponentExcludingTx(
	ctx context.Context,
	tx *sql.Tx,
	component schedulingconflicts.Component,
	excludedConflictID int,
) error {
	if len(component.Items) < 2 {
		return nil
	}

	existingIDs, err :=
		findPendingConflictsForComponentTx(
			ctx,
			tx,
			component,
		)

	if err != nil {
		return err
	}

	filteredIDs := []int{}

	for _, id := range existingIDs {
		if id == excludedConflictID {
			continue
		}

		filteredIDs = append(
			filteredIDs,
			id,
		)
	}

	var conflictID int

	if len(filteredIDs) == 0 {
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
			RETURNING id
			`,
			component.ResourceID,
		).Scan(&conflictID)

		if err != nil {
			return err
		}
	} else {
		sort.Ints(filteredIDs)

		conflictID =
			filteredIDs[0]

		for _, otherID := range filteredIDs[1:] {

			if err :=
				mergeSchedulingConflictsTx(
					ctx,
					tx,
					conflictID,
					otherID,
				); err != nil {

				return err
			}
		}
	}

	for _, item := range component.Items {
		if err :=
			insertSchedulingConflictItemIfMissingTx(
				ctx,
				tx,
				conflictID,
				item,
			); err != nil {

			return err
		}
	}

	_, err = tx.ExecContext(
		ctx,
		`
		UPDATE scheduling_conflicts

		SET updated_at = CURRENT_TIMESTAMP

		WHERE id = $1
		`,
		conflictID,
	)

	return err
}

// ============================================================================
// LISTADO ADMINISTRATIVO DE CONFLICTOS
// ============================================================================
//
// status vacío significa "todos".
//
// El listado reutiliza getSchedulingConflictByID para mantener una única
// representación enriquecida del conflicto y de sus N elementos.
//
// Para el volumen esperado en MVP2 esta estrategia prioriza claridad y evita
// duplicar la lógica de hidratación. Si el volumen crece, podremos introducir
// una proyección resumida/paginada sin cambiar el contrato del detalle.

func GetSchedulingConflicts(
	status string,
) ([]models.SchedulingConflict, error) {
	ctx := context.Background()

	args := []any{}

	query := `
		SELECT id

		FROM scheduling_conflicts
	`

	if status != "" {
		query += `
		WHERE status = $1
		`

		args = append(
			args,
			status,
		)
	}

	query += `
		ORDER BY
			CASE
				WHEN status = 'PENDING' THEN 0
				ELSE 1
			END,
			created_at DESC,
			id DESC
	`

	rows, err := database.DB.QueryContext(
		ctx,
		query,
		args...,
	)

	if err != nil {
		return nil, err
	}

	conflictIDs := []int{}

	for rows.Next() {
		var conflictID int

		if err := rows.Scan(
			&conflictID,
		); err != nil {
			rows.Close()
			return nil, err
		}

		conflictIDs = append(
			conflictIDs,
			conflictID,
		)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}

	// Cerramos el cursor antes de cargar detalles individuales.
	// Esto evita mantener una consulta abierta mientras realizamos las
	// siguientes lecturas sobre el pool de conexiones.
	rows.Close()

	conflicts := make(
		[]models.SchedulingConflict,
		0,
		len(conflictIDs),
	)

	for _, conflictID := range conflictIDs {

		conflict, err :=
			getSchedulingConflictByID(
				ctx,
				database.DB,
				conflictID,
			)

		if err != nil {
			return nil, err
		}

		conflicts = append(
			conflicts,
			conflict,
		)
	}

	return conflicts, nil
}
