package repositories

import (
	"context"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func GetActiveScheduledActivities() ([]models.ScheduledActivity, error) {
	rows, err := database.DB.QueryContext(
		context.Background(),
		`
		SELECT
			s.id,
			s.resource_id,
			s.title,
			s.activity_type,
			s.start_time,
			s.end_time,
			res.name,
			s.created_by_user_id
		FROM dbo.scheduled_activities s
		INNER JOIN dbo.resources res
			ON res.id = s.resource_id
		WHERE s.is_active = 1
		ORDER BY s.start_time ASC;
		`,
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

		activities = append(activities, activity)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return activities, nil
}
