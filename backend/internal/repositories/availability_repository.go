package repositories

import (
	"context"
	"database/sql"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func GetAvailabilityReservationsForRange(
	from time.Time,
	toExclusive time.Time,
	userID int,
) ([]models.AvailabilityReservation, error) {
	if err := ExpirePendingGroupReservations(businessclock.Now()); err != nil {
		return nil, err
	}

	rows, err := database.DB.QueryContext(context.Background(), `
		SELECT
			r.id, r.policy_id, r.user_id, r.resource_id, r.activity_id,
			r.start_time, r.duration_minutes, r.status, r.created_at, r.updated_at,
			COALESCE(a.name, 'Reserva'), res.name,
			COALESCE(u.full_name, ''), COALESCE(u.email, ''), COALESCE(u.rut, ''),
			COALESCE(r.target_participants, r.group_capacity_snapshot),
			r.group_capacity_snapshot, p.minimum_participants,
			(SELECT COUNT(*) FROM dbo.participants pa
			 WHERE pa.reservation_id = r.id AND pa.status = 'CONFIRMED'),
			p.confirmation_deadline_minutes,
			res.reservation_mode,
			CASE WHEN r.user_id = @p3 OR EXISTS (
				SELECT 1 FROM dbo.participants current_pa
				WHERE current_pa.reservation_id = r.id
				  AND current_pa.user_id = @p3
				  AND current_pa.status = 'CONFIRMED'
			) THEN CAST(1 AS BIT) ELSE CAST(0 AS BIT) END
		FROM dbo.reservations r
		INNER JOIN dbo.resources res ON res.id = r.resource_id
		INNER JOIN dbo.users u ON u.id = r.user_id
		INNER JOIN dbo.reservation_policies p ON p.id = r.policy_id
		LEFT JOIN dbo.activities a ON a.id = r.activity_id
		WHERE r.status IN ('PENDING', 'CONFIRMED')
		  AND r.start_time < @p2
		  AND DATEADD(MINUTE, r.duration_minutes, r.start_time) > @p1
		ORDER BY r.start_time ASC, r.id ASC;`,
		businessclock.ToDatabaseWallTime(from),
		businessclock.ToDatabaseWallTime(toExclusive),
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.AvailabilityReservation{}
	for rows.Next() {
		var item models.AvailabilityReservation
		var activityName, resourceName, userFullName, userEmail, userRUT string
		var target, capacity, minimum, count, deadlineMinutes sql.NullInt64
		if err := rows.Scan(
			&item.ID, &item.PolicyID, &item.UserID, &item.ResourceID, &item.ActivityID,
			&item.StartTime, &item.DurationMinutes, &item.Status, &item.CreatedAt, &item.UpdatedAt,
			&activityName, &resourceName, &userFullName, &userEmail, &userRUT,
			&target, &capacity, &minimum, &count, &deadlineMinutes,
			&item.ReservationMode, &item.IsCurrentUserConflict,
		); err != nil {
			return nil, err
		}

		item.StartTime = businessclock.FromDatabaseWallTime(item.StartTime)
		item.Hour = item.StartTime.Format("15:04")
		item.Title = activityName
		item.Type = mapReservationType(item.Status)
		item.ResourceName = resourceName
		item.UserFullName = userFullName
		item.UserEmail = userEmail
		item.UserRUT = userRUT
		if target.Valid {
			value := int(target.Int64)
			item.TargetParticipants = &value
		}
		if capacity.Valid {
			value := int(capacity.Int64)
			item.Capacity = &value
		}
		if minimum.Valid {
			item.MinimumParticipants = int(minimum.Int64)
		}
		if count.Valid {
			item.ParticipantCount = int(count.Int64)
		}
		if deadlineMinutes.Valid && target.Valid {
			deadline := businessclock.ConfirmationDeadline(item.StartTime, int(deadlineMinutes.Int64))
			item.ConfirmationDeadline = &deadline
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func GetScheduledActivitiesForRange(from, toExclusive time.Time) ([]models.ScheduledActivity, error) {
	rows, err := database.DB.QueryContext(context.Background(), `
		SELECT s.id, s.resource_id, s.title, s.activity_type, s.start_time, s.end_time,
		       res.name, res.reservation_mode, s.created_by_user_id
		FROM dbo.scheduled_activities s
		INNER JOIN dbo.resources res ON res.id = s.resource_id
		WHERE s.is_active = 1
		  AND s.start_time < @p2 AND s.end_time > @p1
		ORDER BY s.start_time ASC, s.id ASC;`,
		businessclock.ToDatabaseWallTime(from),
		businessclock.ToDatabaseWallTime(toExclusive),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.ScheduledActivity{}
	for rows.Next() {
		var item models.ScheduledActivity
		if err := rows.Scan(
			&item.ID, &item.ResourceID, &item.Title, &item.ActivityType,
			&item.StartTime, &item.EndTime, &item.ResourceName,
			&item.ReservationMode, &item.CreatedByUserID,
		); err != nil {
			return nil, err
		}
		item.StartTime = businessclock.FromDatabaseWallTime(item.StartTime)
		item.EndTime = businessclock.FromDatabaseWallTime(item.EndTime)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func GetAvailabilityBlocksForRange(from, toExclusive time.Time) ([]models.AvailabilityBlock, error) {
	rows, err := database.DB.QueryContext(context.Background(), `
		SELECT b.id, b.resource_id, b.created_by_user_id, b.block_type,
		       COALESCE(b.reason, ''), b.start_time, b.end_time, res.name
		FROM dbo.availability_blocks b
		INNER JOIN dbo.resources res ON res.id = b.resource_id
		WHERE b.is_active = 1
		  AND b.start_time < @p2 AND b.end_time > @p1
		ORDER BY b.start_time ASC, b.id ASC;`,
		businessclock.ToDatabaseWallTime(from),
		businessclock.ToDatabaseWallTime(toExclusive),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.AvailabilityBlock{}
	for rows.Next() {
		var item models.AvailabilityBlock
		if err := rows.Scan(
			&item.ID, &item.ResourceID, &item.CreatedByUserID, &item.BlockType,
			&item.Reason, &item.StartTime, &item.EndTime, &item.ResourceName,
		); err != nil {
			return nil, err
		}
		item.StartTime = businessclock.FromDatabaseWallTime(item.StartTime)
		item.EndTime = businessclock.FromDatabaseWallTime(item.EndTime)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func GetWorkshopOccurrencesForRange(from, toExclusive time.Time) ([]models.WorkshopAvailabilityOccurrence, error) {
	rows, err := database.DB.QueryContext(context.Background(), `
		WITH calendar_dates AS (
			SELECT CAST(@p1 AS DATE) AS calendar_date
			UNION ALL
			SELECT DATEADD(DAY, 1, calendar_date)
			FROM calendar_dates
			WHERE DATEADD(DAY, 1, calendar_date) < CAST(@p2 AS DATE)
		)
		SELECT wo.id, w.id, w.resource_id, w.title,
		       DATEADD(MINUTE, wo.start_minute, CAST(d.calendar_date AS DATETIME2)),
		       DATEADD(MINUTE, wo.end_minute, CAST(d.calendar_date AS DATETIME2)),
		       res.name, res.reservation_mode
		FROM calendar_dates d
		INNER JOIN dbo.workshop_occurrences wo
		  ON wo.weekday_iso = (DATEDIFF(DAY, CONVERT(DATE, '19000101', 112), d.calendar_date) % 7) + 1
		INNER JOIN dbo.workshops w ON w.id = wo.workshop_id AND w.is_active = 1
		INNER JOIN dbo.resources res ON res.id = w.resource_id
		ORDER BY d.calendar_date ASC, wo.start_minute ASC, wo.id ASC
		OPTION (MAXRECURSION 31);`,
		businessclock.ToDatabaseWallTime(from),
		businessclock.ToDatabaseWallTime(toExclusive),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.WorkshopAvailabilityOccurrence{}
	for rows.Next() {
		var item models.WorkshopAvailabilityOccurrence
		if err := rows.Scan(
			&item.ID, &item.WorkshopID, &item.ResourceID, &item.Title,
			&item.StartTime, &item.EndTime, &item.ResourceName, &item.ReservationMode,
		); err != nil {
			return nil, err
		}
		item.StartTime = businessclock.FromDatabaseWallTime(item.StartTime)
		item.EndTime = businessclock.FromDatabaseWallTime(item.EndTime)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func HasWorkshopAvailabilityConflict(resourceID int, startTime, endTime time.Time) (bool, error) {
	weekdayISO := int(startTime.Weekday())
	if weekdayISO == 0 {
		weekdayISO = 7
	}
	startMinute := startTime.Hour()*60 + startTime.Minute()
	endMinute := endTime.Hour()*60 + endTime.Minute()

	var conflict bool
	err := database.DB.QueryRowContext(context.Background(), `
		SELECT CASE WHEN EXISTS (
			SELECT 1
			FROM dbo.workshop_occurrences wo
			INNER JOIN dbo.workshops w ON w.id = wo.workshop_id
			WHERE w.is_active = 1
			  AND w.resource_id = @p1
			  AND wo.weekday_iso = @p2
			  AND wo.start_minute < @p4
			  AND @p3 < wo.end_minute
		) THEN CAST(1 AS BIT) ELSE CAST(0 AS BIT) END;`,
		resourceID, weekdayISO, startMinute, endMinute,
	).Scan(&conflict)
	return conflict, err
}
