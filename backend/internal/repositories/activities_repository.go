package repositories

import (
	"context"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/validators"
)

func GetActiveActivities() ([]models.Activity, error) {
	rows, err := database.DB.QueryContext(
		context.Background(),
		`
		SELECT
			id,
			name,
			COALESCE(description, '') AS description,
			is_active
		FROM activities
		WHERE is_active = true
		ORDER BY name ASC;
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	activities := []models.Activity{}

	for rows.Next() {
		var activity models.Activity

		err := rows.Scan(
			&activity.ID,
			&activity.Name,
			&activity.Description,
			&activity.IsActive,
		)

		if err != nil {
			return nil, err
		}

		if validators.IsValidActivityName(activity.Name) {
			activities = append(activities, activity)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return activities, nil
}
