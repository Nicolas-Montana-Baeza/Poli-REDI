package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

var (
	ErrInstitutionalOccurrenceInvalid        = errors.New("sesión semanal o datos de excepción inválidos")
	ErrInstitutionalOccurrenceAlreadyHandled = errors.New("esta sesión ya tiene una excepción registrada")
	ErrInstitutionalOccurrenceEnrolled       = errors.New("la actividad tiene personas inscritas; resuelve sus inscripciones y avisos antes de modificar una sesión")
	ErrInstitutionalOccurrenceBlocked        = errors.New("la nueva sesión intersecta un bloqueo administrativo")
	ErrInstitutionalOccurrenceSelfOverlap    = errors.New("la nueva sesión se solapa con otra sesión de la misma actividad")
)

type InstitutionalOccurrencePatch struct {
	ScheduleID   int
	OriginalDate string
	Action       string
	Reason       string
	NewDate      *string
	NewStartTime *string
	NewEndTime   *string
}

type InstitutionalOccurrenceException struct {
	ID           int                               `json:"id"`
	ActivityID   int                               `json:"activityId"`
	ScheduleID   int                               `json:"scheduleId"`
	OriginalDate string                            `json:"originalDate"`
	Action       string                            `json:"action"`
	Reason       string                            `json:"reason"`
	NewDate      *string                           `json:"newDate,omitempty"`
	NewStartTime *string                           `json:"newStartTime,omitempty"`
	NewEndTime   *string                           `json:"newEndTime,omitempty"`
	CreatedAt    time.Time                         `json:"createdAt"`
	Revisions    []InstitutionalOccurrenceRevision `json:"revisions,omitempty"`
}

type InstitutionalOccurrenceRevision struct {
	ID                int64     `json:"id"`
	PreviousAction    string    `json:"previousAction"`
	NewAction         string    `json:"newAction"`
	PreviousDate      *string   `json:"previousDate,omitempty"`
	PreviousStartTime *string   `json:"previousStartTime,omitempty"`
	PreviousEndTime   *string   `json:"previousEndTime,omitempty"`
	NewDate           *string   `json:"newDate,omitempty"`
	NewStartTime      *string   `json:"newStartTime,omitempty"`
	NewEndTime        *string   `json:"newEndTime,omitempty"`
	Reason            string    `json:"reason"`
	ChangedByUserID   int       `json:"changedByUserId"`
	ChangedAt         time.Time `json:"changedAt"`
}

const exceptionColumns = `id,activity_id,schedule_id,
    to_char(original_date,'YYYY-MM-DD'),exception_type,reason,
    CASE WHEN new_date IS NULL THEN NULL ELSE to_char(new_date,'YYYY-MM-DD') END,
    CASE WHEN new_start_time IS NULL THEN NULL ELSE to_char(new_start_time,'HH24:MI') END,
    CASE WHEN new_end_time IS NULL THEN NULL ELSE to_char(new_end_time,'HH24:MI') END,
    created_at`

func scanOccurrenceException(row interface{ Scan(...any) error }) (InstitutionalOccurrenceException, error) {
	var result InstitutionalOccurrenceException
	var date, start, end sql.NullString
	err := row.Scan(&result.ID, &result.ActivityID, &result.ScheduleID, &result.OriginalDate, &result.Action, &result.Reason, &date, &start, &end, &result.CreatedAt)
	if date.Valid {
		result.NewDate = &date.String
	}
	if start.Valid {
		result.NewStartTime = &start.String
	}
	if end.Valid {
		result.NewEndTime = &end.String
	}
	return result, err
}

func GetInstitutionalOccurrenceExceptions(activityID int) ([]InstitutionalOccurrenceException, error) {
	rows, err := database.DB.QueryContext(context.Background(),
		"SELECT "+exceptionColumns+" FROM institutional_activity_schedule_exceptions WHERE activity_id=$1 ORDER BY original_date,schedule_id", activityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []InstitutionalOccurrenceException{}
	for rows.Next() {
		item, scanErr := scanOccurrenceException(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}
	for i := range items {
		revisions, loadErr := GetInstitutionalOccurrenceRevisions(items[i].ID)
		if loadErr != nil {
			return nil, loadErr
		}
		items[i].Revisions = revisions
	}
	return items, nil
}

func GetInstitutionalOccurrenceRevisions(exceptionID int) ([]InstitutionalOccurrenceRevision, error) {
	rows, err := database.DB.QueryContext(context.Background(), `SELECT id,previous_exception_type,new_exception_type,
       CASE WHEN previous_date IS NULL THEN NULL ELSE to_char(previous_date,'YYYY-MM-DD') END,
       CASE WHEN previous_start_time IS NULL THEN NULL ELSE to_char(previous_start_time,'HH24:MI') END,
       CASE WHEN previous_end_time IS NULL THEN NULL ELSE to_char(previous_end_time,'HH24:MI') END,
       CASE WHEN new_date IS NULL THEN NULL ELSE to_char(new_date,'YYYY-MM-DD') END,
       CASE WHEN new_start_time IS NULL THEN NULL ELSE to_char(new_start_time,'HH24:MI') END,
       CASE WHEN new_end_time IS NULL THEN NULL ELSE to_char(new_end_time,'HH24:MI') END,
       reason,changed_by_user_id,changed_at
       FROM institutional_activity_schedule_exception_revisions
       WHERE exception_id=$1 ORDER BY id ASC`, exceptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	output := []InstitutionalOccurrenceRevision{}
	for rows.Next() {
		var value InstitutionalOccurrenceRevision
		var pd, ps, pe, nd, ns, ne sql.NullString
		if err = rows.Scan(&value.ID, &value.PreviousAction, &value.NewAction,
			&pd, &ps, &pe, &nd, &ns, &ne, &value.Reason, &value.ChangedByUserID, &value.ChangedAt); err != nil {
			return nil, err
		}
		if pd.Valid {
			value.PreviousDate = &pd.String
		}
		if ps.Valid {
			value.PreviousStartTime = &ps.String
		}
		if pe.Valid {
			value.PreviousEndTime = &pe.String
		}
		if nd.Valid {
			value.NewDate = &nd.String
		}
		if ns.Valid {
			value.NewStartTime = &ns.String
		}
		if ne.Valid {
			value.NewEndTime = &ne.String
		}
		output = append(output, value)
	}
	return output, rows.Err()
}

// La excepción se escribe en una sola transacción. El UNIQUE(schedule_id,original_date)
// evita cancelar y reprogramar a la vez, incluso bajo peticiones concurrentes.
// ApplyInstitutionalOccurrenceException aplica la primera decisión o revisa
// la excepción existente. El trigger de PostgreSQL archiva cada revisión.
// Se mantiene una única fila vigente por (schedule_id,original_date).
func ApplyInstitutionalOccurrenceException(activityID, actorID int, p InstitutionalOccurrencePatch) (InstitutionalOccurrenceException, error) {
	empty := InstitutionalOccurrenceException{}
	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return empty, err
	}
	defer tx.Rollback()

	var resourceID int
	var status string
	err = tx.QueryRowContext(ctx, `SELECT resource_id,status FROM institutional_activities WHERE id=$1 FOR UPDATE`, activityID).Scan(&resourceID, &status)
	if err != nil {
		return empty, err
	}
	if status != models.InstitutionalActivityStatusScheduled {
		return empty, ErrInstitutionalActivityAlreadyCancelled
	}

	var validFrom, validTo, startHH, endHH string
	var weekday int
	err = tx.QueryRowContext(ctx, `
      SELECT to_char(valid_from,'YYYY-MM-DD'),to_char(valid_to,'YYYY-MM-DD'),day_of_week,
       to_char(start_time,'HH24:MI'),to_char(end_time,'HH24:MI')
      FROM institutional_activity_schedules
      WHERE id=$1 AND activity_id=$2 AND is_active=true AND schedule_type='WEEKLY'
      FOR UPDATE`, p.ScheduleID, activityID).Scan(&validFrom, &validTo, &weekday, &startHH, &endHH)
	if errors.Is(err, sql.ErrNoRows) {
		return empty, ErrInstitutionalOccurrenceInvalid
	}
	if err != nil {
		return empty, err
	}

	original, err := time.ParseInLocation("2006-01-02", p.OriginalDate, businessclock.Location())
	if err != nil {
		return empty, ErrInstitutionalOccurrenceInvalid
	}
	isoDay := int(original.Weekday())
	if isoDay == 0 {
		isoDay = 7
	}
	if p.OriginalDate < validFrom || p.OriginalDate > validTo || weekday != isoDay {
		return empty, ErrInstitutionalOccurrenceInvalid
	}
	startClock, _ := time.Parse("15:04", startHH)
	originalStart := time.Date(original.Year(), original.Month(), original.Day(), startClock.Hour(), startClock.Minute(), 0, 0, businessclock.Location())

	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(73001,$1)`, resourceID); err != nil {
		return empty, err
	}
	var enrolled bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM institutional_activity_enrollments WHERE activity_id=$1 AND status='CONFIRMED')`, activityID).Scan(&enrolled)
	if err != nil {
		return empty, err
	}
	if enrolled {
		return empty, ErrInstitutionalOccurrenceEnrolled
	}

	// Bloqueo de fila para que dos administradores no revisen a la vez.
	var existingID int
	var oldAction string
	var oldDate, oldStart, oldEnd sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT id,exception_type,
       CASE WHEN new_date IS NULL THEN NULL ELSE to_char(new_date,'YYYY-MM-DD') END,
       CASE WHEN new_start_time IS NULL THEN NULL ELSE to_char(new_start_time,'HH24:MI') END,
       CASE WHEN new_end_time IS NULL THEN NULL ELSE to_char(new_end_time,'HH24:MI') END
       FROM institutional_activity_schedule_exceptions
       WHERE activity_id=$1 AND schedule_id=$2 AND original_date=$3::date FOR UPDATE`, activityID, p.ScheduleID, p.OriginalDate).
		Scan(&existingID, &oldAction, &oldDate, &oldStart, &oldEnd)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return empty, err
	}
	existing := err == nil

	// La fecha que manda es la sesión VIGENTE, no necesariamente la original.
	currentStart := originalStart
	currentDate, currentHH := p.OriginalDate, startHH
	if existing && oldAction == "RESCHEDULE" && oldDate.Valid && oldStart.Valid {
		d, parseErr := time.ParseInLocation("2006-01-02", oldDate.String, businessclock.Location())
		if parseErr != nil {
			return empty, ErrInstitutionalOccurrenceInvalid
		}
		h, parseErr := time.Parse("15:04", oldStart.String)
		if parseErr != nil {
			return empty, ErrInstitutionalOccurrenceInvalid
		}
		currentStart = time.Date(d.Year(), d.Month(), d.Day(), h.Hour(), h.Minute(), 0, 0, businessclock.Location())
		currentDate, currentHH = oldDate.String, oldStart.String
	}
	if !currentStart.After(businessclock.Now()) {
		return empty, ErrInstitutionalOccurrenceInvalid
	}

	if p.Action == "RESCHEDULE" {
		candidate := models.CreateInstitutionalScheduleRequest{
			ScheduleType: models.InstitutionalScheduleTypeSingle, SpecificDate: p.NewDate,
			StartTime: *p.NewStartTime, EndTime: *p.NewEndTime,
		}
		blocked, checkErr := institutionalScheduleHasAvailabilityBlockTx(ctx, tx, resourceID, candidate)
		if checkErr != nil {
			return empty, checkErr
		}
		if blocked {
			return empty, ErrInstitutionalOccurrenceBlocked
		}

		// No considerar la ocupación anterior de ESTA misma sesión.
		var overlaps bool
		err = tx.QueryRowContext(ctx, `
          WITH other AS (
            SELECT s.start_time AS start_at,s.end_time AS end_at
            FROM institutional_activity_schedules s
            WHERE s.activity_id=$1 AND s.is_active=true
              AND ((s.schedule_type='SINGLE' AND s.specific_date=$3::date)
                OR (s.schedule_type='WEEKLY' AND $3::date BETWEEN s.valid_from AND s.valid_to
                  AND extract(isodow from $3::date)::integer=s.day_of_week))
              AND NOT (s.id=$2 AND $3::date=$4::date)
              AND NOT EXISTS (SELECT 1 FROM institutional_activity_schedule_exceptions e
                WHERE e.schedule_id=s.id AND e.original_date=$3::date)
            UNION ALL
            SELECT e.new_start_time,e.new_end_time
            FROM institutional_activity_schedule_exceptions e
            JOIN institutional_activity_schedules s ON s.id=e.schedule_id AND s.is_active=true
            WHERE e.activity_id=$1 AND e.exception_type='RESCHEDULE' AND e.new_date=$3::date
              AND NOT (e.schedule_id=$2 AND e.original_date=$4::date)
          ) SELECT EXISTS (SELECT 1 FROM other WHERE start_at<$6::time AND end_at>$5::time)`,
			activityID, p.ScheduleID, *p.NewDate, p.OriginalDate, *p.NewStartTime, *p.NewEndTime).Scan(&overlaps)
		if err != nil {
			return empty, err
		}
		if overlaps {
			return empty, ErrInstitutionalOccurrenceSelfOverlap
		}
	}

	var result InstitutionalOccurrenceException
	if existing {
		// La migración hace que el trigger registre la revisión anterior.
		if _, err = tx.ExecContext(ctx, `SELECT set_config('poliredi.exception_revision_actor_id',$1,true)`, strconv.Itoa(actorID)); err != nil {
			return empty, err
		}
		result, err = scanOccurrenceException(tx.QueryRowContext(ctx, `
          UPDATE institutional_activity_schedule_exceptions SET
            exception_type=$2,new_date=$3::date,new_start_time=$4::time,
            new_end_time=$5::time,reason=$6,updated_at=CURRENT_TIMESTAMP
          WHERE id=$1 RETURNING `+exceptionColumns,
			existingID, p.Action, p.NewDate, p.NewStartTime, p.NewEndTime, p.Reason))
	} else {
		result, err = scanOccurrenceException(tx.QueryRowContext(ctx,
			"INSERT INTO institutional_activity_schedule_exceptions "+
				"(activity_id,schedule_id,original_date,exception_type,new_date,new_start_time,new_end_time,created_by_user_id,reason) "+
				"VALUES($1,$2,$3::date,$4,$5::date,$6::time,$7::time,$8,$9) RETURNING "+exceptionColumns,
			activityID, p.ScheduleID, p.OriginalDate, p.Action, p.NewDate, p.NewStartTime, p.NewEndTime, actorID, p.Reason))
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return empty, ErrInstitutionalOccurrenceAlreadyHandled
		}
		return empty, err
	}

	// Cerrar conflictos de la ocupación VIGENTE anterior, no solo original.
	note := "Sesión gestionada: " + p.Reason
	_, err = tx.ExecContext(ctx, `
      UPDATE scheduling_conflict_items SET resolution='CANCEL',resolution_source='MANUAL',
        resolved_by_user_id=$3,resolved_at=CURRENT_TIMESTAMP,resolution_note=$4,
        updated_at=CURRENT_TIMESTAMP
      WHERE institutional_activity_id=$1 AND schedule_id=$2 AND resolution='PENDING'
        AND occurrence_start=(($5::date+$6::time) AT TIME ZONE 'America/Santiago')`,
		activityID, p.ScheduleID, actorID, note, currentDate, currentHH)
	if err != nil {
		return empty, err
	}
	_, err = tx.ExecContext(ctx, `
      UPDATE scheduling_conflicts c SET status='RESOLVED',resolution_source='MANUAL',
        resolution_summary=$3,resolved_by_user_id=$2,resolved_at=CURRENT_TIMESTAMP,updated_at=CURRENT_TIMESTAMP
      WHERE c.status='PENDING'
        AND EXISTS (SELECT 1 FROM scheduling_conflict_items i
          WHERE i.conflict_id=c.id AND i.institutional_activity_id=$1
            AND i.schedule_id=$4 AND i.occurrence_start=(($5::date+$6::time) AT TIME ZONE 'America/Santiago'))
        AND NOT EXISTS (SELECT 1 FROM scheduling_conflict_items pending
          WHERE pending.conflict_id=c.id AND pending.resolution='PENDING')`,
		activityID, actorID, note, p.ScheduleID, currentDate, currentHH)
	if err != nil {
		return empty, err
	}

	if p.Action == "RESCHEDULE" {
		if err = DetectAndPersistSchedulingConflictsForActivityTx(ctx, tx, activityID, resourceID); err != nil {
			return empty, err
		}
	}
	if err = tx.Commit(); err != nil {
		return empty, fmt.Errorf("confirmar revisión de sesión: %w", err)
	}
	return result, nil
}
