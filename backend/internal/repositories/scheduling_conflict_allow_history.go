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
// HISTORIAL DE COEXISTENCIAS AUTORIZADAS
// ============================================================================
//
// ALLOW + ALLOW no elimina ninguna ocupación.
//
// Representa una decisión administrativa explícita:
//
//     estas dos ocurrencias concretas pueden coexistir
//
// La autorización pertenece al SNAPSHOT exacto:
//
//     entidad
//     inicio
//     fin
//
// Si posteriormente una programación cambia, la autorización histórica no se
// reutiliza accidentalmente para otra ocurrencia.

func detectSchedulingConflictComponentsTx(
	ctx context.Context,
	tx *sql.Tx,
	occupancies []schedulingconflicts.Occupancy,
) ([]schedulingconflicts.Component, error) {
	if len(occupancies) == 0 {
		return []schedulingconflicts.Component{},
			nil
	}

	allowedPairs, err :=
		loadSchedulingAllowedPairsTx(
			ctx,
			tx,
			occupancies,
		)

	if err != nil {
		return nil, err
	}

	return schedulingconflicts.
		DetectConnectedComponentsWithPredicate(
			occupancies,
			func(
				left schedulingconflicts.Occupancy,
				right schedulingconflicts.Occupancy,
			) bool {
				// Primero aplicamos la semántica estructural:
				//
				// reserva ↔ reserva nunca es arista administrativa.
				if !schedulingconflicts.
					DefaultConflictPredicate(
						left,
						right,
					) {

					return false
				}

				pair :=
					schedulingAllowedPairKey(
						schedulingOccupancySignature(
							left,
						),
						schedulingOccupancySignature(
							right,
						),
					)

				// Una pareja ALLOW + ALLOW ya autorizada deja de ser una
				// incompatibilidad para esa ocurrencia exacta.
				if _, allowed :=
					allowedPairs[pair]; allowed {

					return false
				}

				return true
			},
		)
}

func loadSchedulingAllowedPairsTx(
	ctx context.Context,
	tx *sql.Tx,
	occupancies []schedulingconflicts.Occupancy,
) (map[string]struct{}, error) {
	result :=
		map[string]struct{}{}

	if len(occupancies) == 0 {
		return result, nil
	}

	resourceID :=
		occupancies[0].ResourceID

	minStart :=
		occupancies[0].Start

	maxEnd :=
		occupancies[0].End

	for _, occupancy := range occupancies[1:] {

		if occupancy.ResourceID !=
			resourceID {

			// Las llamadas actuales trabajan por recurso y fecha.
			// Si esto cambia en el futuro, es mejor fallar explícitamente
			// que aplicar autorizaciones cruzadas.
			return nil, fmt.Errorf(
				"occupancies de múltiples recursos en historial ALLOW",
			)
		}

		if occupancy.Start.Before(
			minStart,
		) {
			minStart = occupancy.Start
		}

		if occupancy.End.After(
			maxEnd,
		) {
			maxEnd = occupancy.End
		}
	}

	rows, err := tx.QueryContext(
		ctx,
		`
		SELECT
			left_item.institutional_activity_id,
			left_item.schedule_id,
			left_item.reservation_id,
			left_item.occurrence_start,
			left_item.occurrence_end,

			right_item.institutional_activity_id,
			right_item.schedule_id,
			right_item.reservation_id,
			right_item.occurrence_start,
			right_item.occurrence_end

		FROM scheduling_conflict_items left_item

		INNER JOIN scheduling_conflict_items right_item
			ON right_item.conflict_id =
				left_item.conflict_id
		   AND right_item.id >
				left_item.id

		INNER JOIN scheduling_conflicts conflict
			ON conflict.id =
				left_item.conflict_id

		WHERE conflict.resource_id = $1

		  AND left_item.resolution = 'ALLOW'
		  AND right_item.resolution = 'ALLOW'

		  -- Basta consultar snapshots que intersectan la ventana actual.
		  AND left_item.occurrence_start < $3
		  AND left_item.occurrence_end > $2

		  AND right_item.occurrence_start < $3
		  AND right_item.occurrence_end > $2
		`,
		resourceID,
		minStart,
		maxEnd,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			leftActivity    sql.NullInt64
			leftSchedule    sql.NullInt64
			leftReservation sql.NullInt64
			leftStart       time.Time
			leftEnd         time.Time

			rightActivity    sql.NullInt64
			rightSchedule    sql.NullInt64
			rightReservation sql.NullInt64
			rightStart       time.Time
			rightEnd         time.Time
		)

		if err := rows.Scan(
			&leftActivity,
			&leftSchedule,
			&leftReservation,
			&leftStart,
			&leftEnd,
			&rightActivity,
			&rightSchedule,
			&rightReservation,
			&rightStart,
			&rightEnd,
		); err != nil {

			return nil, err
		}

		leftSignature, err :=
			schedulingConflictItemSignature(
				leftActivity,
				leftSchedule,
				leftReservation,
				leftStart,
				leftEnd,
			)

		if err != nil {
			return nil, err
		}

		rightSignature, err :=
			schedulingConflictItemSignature(
				rightActivity,
				rightSchedule,
				rightReservation,
				rightStart,
				rightEnd,
			)

		if err != nil {
			return nil, err
		}

		result[schedulingAllowedPairKey(
			leftSignature,
			rightSignature,
		)] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func schedulingOccupancySignature(
	occupancy schedulingconflicts.Occupancy,
) string {
	start :=
		occupancy.Start.
			In(
				businessclock.Location(),
			).
			Format(
				time.RFC3339Nano,
			)

	end :=
		occupancy.End.
			In(
				businessclock.Location(),
			).
			Format(
				time.RFC3339Nano,
			)

	if occupancy.Kind ==
		schedulingconflicts.
			OccupancyKindInstitutionalActivity &&
		occupancy.InstitutionalActivityID != nil &&
		occupancy.ScheduleID != nil {

		return fmt.Sprintf(
			"activity:%d:schedule:%d:%s:%s",
			*occupancy.InstitutionalActivityID,
			*occupancy.ScheduleID,
			start,
			end,
		)
	}

	if occupancy.Kind ==
		schedulingconflicts.
			OccupancyKindReservation &&
		occupancy.ReservationID != nil {

		return fmt.Sprintf(
			"reservation:%d:%s:%s",
			*occupancy.ReservationID,
			start,
			end,
		)
	}

	return "invalid"
}

func schedulingConflictItemSignature(
	activityID sql.NullInt64,
	scheduleID sql.NullInt64,
	reservationID sql.NullInt64,
	start time.Time,
	end time.Time,
) (string, error) {
	startValue :=
		start.
			In(
				businessclock.Location(),
			).
			Format(
				time.RFC3339Nano,
			)

	endValue :=
		end.
			In(
				businessclock.Location(),
			).
			Format(
				time.RFC3339Nano,
			)

	if activityID.Valid &&
		scheduleID.Valid &&
		!reservationID.Valid {

		return fmt.Sprintf(
			"activity:%d:schedule:%d:%s:%s",
			activityID.Int64,
			scheduleID.Int64,
			startValue,
			endValue,
		), nil
	}

	if reservationID.Valid &&
		!activityID.Valid &&
		!scheduleID.Valid {

		return fmt.Sprintf(
			"reservation:%d:%s:%s",
			reservationID.Int64,
			startValue,
			endValue,
		), nil
	}

	return "", fmt.Errorf(
		"snapshot de scheduling_conflict_item inválido",
	)
}

func schedulingAllowedPairKey(
	left string,
	right string,
) string {
	values := []string{
		left,
		right,
	}

	sort.Strings(values)

	return values[0] +
		"||" +
		values[1]
}
