package repositories

import (
	"context"
	"database/sql"
	"strings"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
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
		FROM dbo.activities
		WHERE is_active = 1
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

		activities = append(activities, activity)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return activities, nil
}

func GetOrCreateActivityByName(name string, description string) (*models.Activity, error) {
	ctx := context.Background()
	normalizedName := strings.TrimSpace(name)
	normalizedDescription := strings.TrimSpace(description)

	activity, err := getActivityByName(ctx, normalizedName)

	if err == nil {
		return activity, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	_, err = database.DB.ExecContext(
		ctx,
		`
		INSERT INTO dbo.activities (
			name,
			description,
			is_active
		)
		VALUES (@p1, NULLIF(@p2, ''), 1);
		`,
		normalizedName,
		normalizedDescription,
	)

	if err != nil {
		if existingActivity, selectErr := getActivityByName(ctx, normalizedName); selectErr == nil {
			return existingActivity, nil
		}

		return nil, err
	}

	return getActivityByName(ctx, normalizedName)
}

func getActivityByName(ctx context.Context, name string) (*models.Activity, error) {
	var activity models.Activity

	err := database.DB.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			name,
			COALESCE(description, '') AS description,
			is_active
		FROM dbo.activities
		WHERE LOWER(name) = LOWER(@p1);
		`,
		name,
	).Scan(
		&activity.ID,
		&activity.Name,
		&activity.Description,
		&activity.IsActive,
	)

	if err != nil {
		return nil, err
	}

	return &activity, nil
}
