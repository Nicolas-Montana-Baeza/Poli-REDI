package repositories

import (
	"context"
	"database/sql"
	"errors"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

// La edición de recurso/unidad y de actividades con inscripción activa
// requiere otros flujos: evitar cambiar implícitamente inscripciones o permisos.
var ErrInstitutionalActivityNotEditable = errors.New("la actividad no admite esta modificación")
var ErrInstitutionalActivityResourceChange = errors.New("el cambio de recurso o unidad no está disponible en esta edición")
var ErrInstitutionalActivityAlreadyCancelled = errors.New("la actividad ya no está programada")

// settleActivityConflictItemsTx conserva los snapshots históricos de los
// conflictos. Al retirar una ocupación, sus items PENDING se cierran como
// CANCEL; no se modifica ninguna reserva particular.
func settleActivityConflictItemsTx(ctx context.Context, tx *sql.Tx, activityID, actorID int, note string) error {
	_, err := tx.ExecContext(ctx, `
        UPDATE scheduling_conflict_items
        SET resolution = 'CANCEL',
            resolution_source = 'MANUAL',
            resolved_by_user_id = $2,
            resolved_at = CURRENT_TIMESTAMP,
            resolution_note = $3,
            updated_at = CURRENT_TIMESTAMP
        WHERE institutional_activity_id = $1
          AND resolution = 'PENDING'
    `, activityID, actorID, note)
	if err != nil {
		return err
	}

	// Si no queda ningún item PENDING, resolver el encabezado. Los conflictos
	// que aún contienen decisiones pendientes se conservan para el admin.
	_, err = tx.ExecContext(ctx, `
        UPDATE scheduling_conflicts conflict
        SET status='RESOLVED',
            resolution_source='MANUAL',
            resolution_summary=$3,
            resolved_by_user_id=$2,
            resolved_at=CURRENT_TIMESTAMP,
            updated_at=CURRENT_TIMESTAMP
        WHERE conflict.status='PENDING'
          AND EXISTS (
              SELECT 1 FROM scheduling_conflict_items item
              WHERE item.conflict_id=conflict.id
                AND item.institutional_activity_id=$1
          )
          AND NOT EXISTS (
              SELECT 1 FROM scheduling_conflict_items pending
              WHERE pending.conflict_id=conflict.id
                AND pending.resolution='PENDING'
          )
    `, activityID, actorID, note)
	return err
}

// CancelInstitutionalActivity mantiene actividad, inscripciones y horarios en
// el historial. El calendario solo proyecta actividades SCHEDULED.
func CancelInstitutionalActivity(activityID, actorID int) (models.InstitutionalActivity, error) {
	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return models.InstitutionalActivity{}, err
	}
	defer tx.Rollback()

	var resourceID int
	var status string
	err = tx.QueryRowContext(ctx, `
        SELECT resource_id, status FROM institutional_activities
        WHERE id=$1 FOR UPDATE
    `, activityID).Scan(&resourceID, &status)
	if err != nil {
		return models.InstitutionalActivity{}, err
	}
	if status != models.InstitutionalActivityStatusScheduled {
		return models.InstitutionalActivity{}, ErrInstitutionalActivityAlreadyCancelled
	}
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(73001, $1)`, resourceID); err != nil {
		return models.InstitutionalActivity{}, err
	}
	if _, err = tx.ExecContext(ctx, `
        UPDATE institutional_activities
        SET status='CANCELLED',updated_at=CURRENT_TIMESTAMP
        WHERE id=$1 AND status='SCHEDULED'
    `, activityID); err != nil {
		return models.InstitutionalActivity{}, err
	}

	// Si el prototipo tenía asistentes inscritos, cancelar su inscripción.
	if _, err = tx.ExecContext(ctx, `
        UPDATE institutional_activity_enrollments
        SET status='CANCELLED',updated_at=CURRENT_TIMESTAMP
        WHERE activity_id=$1 AND status='CONFIRMED'
    `, activityID); err != nil {
		return models.InstitutionalActivity{}, err
	}

	if err = settleActivityConflictItemsTx(ctx, tx, activityID, actorID, "Actividad institucional cancelada por el gestor"); err != nil {
		return models.InstitutionalActivity{}, err
	}
	if err = tx.Commit(); err != nil {
		return models.InstitutionalActivity{}, err
	}
	return GetInstitutionalActivityByID(activityID)
}

// UpdateInstitutionalActivityWithSchedules reemplaza programaciones activas
// sin borrar filas históricas. Recalcula conflictos para nuevos horarios.
// MVP3: unidad/recurso son inmutables y se rechaza editar actividades con
// inscripciones activas (el cambio de contrato de taller se hará aparte).
func UpdateInstitutionalActivityWithSchedules(activityID, actorID int, request models.CreateInstitutionalActivityRequest) (models.InstitutionalActivity, error) {
	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return models.InstitutionalActivity{}, err
	}
	defer tx.Rollback()

	var unitID, resourceID int
	var status string
	err = tx.QueryRowContext(ctx, `
        SELECT unit_id,resource_id,status FROM institutional_activities
        WHERE id=$1 FOR UPDATE
    `, activityID).Scan(&unitID, &resourceID, &status)
	if err != nil {
		return models.InstitutionalActivity{}, err
	}
	if status != models.InstitutionalActivityStatusScheduled {
		return models.InstitutionalActivity{}, ErrInstitutionalActivityAlreadyCancelled
	}
	if unitID != request.UnitID || resourceID != request.ResourceID {
		return models.InstitutionalActivity{}, ErrInstitutionalActivityResourceChange
	}

	var enrolled bool
	if err = tx.QueryRowContext(ctx, `
        SELECT EXISTS(SELECT 1 FROM institutional_activity_enrollments
          WHERE activity_id=$1 AND status='CONFIRMED')
    `, activityID).Scan(&enrolled); err != nil {
		return models.InstitutionalActivity{}, err
	}
	if enrolled {
		return models.InstitutionalActivity{}, ErrInstitutionalActivityNotEditable
	}

	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(73001,$1)`, resourceID); err != nil {
		return models.InstitutionalActivity{}, err
	}
	for _, schedule := range request.Schedules {
		blocked, checkErr := institutionalScheduleHasAvailabilityBlockTx(ctx, tx, resourceID, schedule)
		if checkErr != nil {
			return models.InstitutionalActivity{}, checkErr
		}
		if blocked {
			return models.InstitutionalActivity{}, ErrInstitutionalScheduleBlocked
		}
	}

	_, err = tx.ExecContext(ctx, `
        UPDATE institutional_activities SET
          title=$2, description=NULLIF($3,''), activity_type=$4,
          requires_enrollment=$5, capacity=$6, updated_at=CURRENT_TIMESTAMP
        WHERE id=$1
    `, activityID, request.Title, request.Description, request.ActivityType,
		request.RequiresEnrollment, request.Capacity)
	if err != nil {
		return models.InstitutionalActivity{}, err
	}

	// Primero invalidamos horarios antiguos; los IDs quedan para auditoría.
	if _, err = tx.ExecContext(ctx, `
        UPDATE institutional_activity_schedules
        SET is_active=false,updated_at=CURRENT_TIMESTAMP
        WHERE activity_id=$1 AND is_active=true
    `, activityID); err != nil {
		return models.InstitutionalActivity{}, err
	}

	for _, schedule := range request.Schedules {
		_, err = tx.ExecContext(ctx, `
          INSERT INTO institutional_activity_schedules
            (activity_id,schedule_type,specific_date,day_of_week,
             start_time,end_time,valid_from,valid_to,is_active)
          VALUES ($1,$2,$3::date,$4,$5::time,$6::time,$7::date,$8::date,true)
        `, activityID, schedule.ScheduleType, nullableString(schedule.SpecificDate),
			nullableInt(schedule.DayOfWeek), schedule.StartTime, schedule.EndTime,
			nullableString(schedule.ValidFrom), nullableString(schedule.ValidTo))
		if err != nil {
			return models.InstitutionalActivity{}, err
		}
	}

	if err = settleActivityConflictItemsTx(ctx, tx, activityID, actorID, "Programación institucional reemplazada por edición"); err != nil {
		return models.InstitutionalActivity{}, err
	}
	if err = DetectAndPersistSchedulingConflictsForActivityTx(ctx, tx, activityID, resourceID); err != nil {
		return models.InstitutionalActivity{}, err
	}
	if err = tx.Commit(); err != nil {
		return models.InstitutionalActivity{}, err
	}
	return GetInstitutionalActivityByID(activityID)
}
