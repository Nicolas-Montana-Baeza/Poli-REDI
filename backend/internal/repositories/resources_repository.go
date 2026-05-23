package repositories

import (
	"context"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func GetAllResources() ([]models.Resource, error) {
	rows, err := database.DB.Query(
		context.Background(),
		`
		SELECT
			id,
			name,
			type,
			reservation_mode,
			is_active
		FROM resources
		ORDER BY id;
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	resources := []models.Resource{}

	for rows.Next() {
		var resource models.Resource

		err := rows.Scan(
			&resource.ID,
			&resource.Name,
			&resource.Type,
			&resource.ReservationMode,
			&resource.IsActive,
		)

		if err != nil {
			return nil, err
		}

		if resource.IsActive {
			resource.Status = "available"
		} else {
			resource.Status = "maintenance"
		}

		resources = append(resources, resource)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return resources, nil
}
