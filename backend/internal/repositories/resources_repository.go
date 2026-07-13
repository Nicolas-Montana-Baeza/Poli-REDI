package repositories

import (
	"context"
	"database/sql"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func GetAllResources() ([]models.Resource, error) {
	rows, err := database.DB.QueryContext(
		context.Background(),
		`
		SELECT
			id,
			name,
			type,
			reservation_mode,
			COALESCE(image_url, '') AS image_url,
			capacity,
			is_active
		FROM dbo.resources
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
		var capacity sql.NullInt64

		err := rows.Scan(
			&resource.ID,
			&resource.Name,
			&resource.Type,
			&resource.ReservationMode,
			&resource.ImageURL,
			&capacity,
			&resource.IsActive,
		)

		if err != nil {
			return nil, err
		}

		if capacity.Valid {
			value := int(capacity.Int64)
			resource.Capacity = &value
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

func GetResourceByID(id int) (models.Resource, error) {
	row := database.DB.QueryRowContext(
		context.Background(),
		`
		SELECT
			id,
			name,
			type,
			reservation_mode,
			COALESCE(image_url, '') AS image_url,
			capacity,
			is_active
		FROM dbo.resources
		WHERE id = @p1;
		`,
		id,
	)

	var resource models.Resource
	var capacity sql.NullInt64

	err := row.Scan(
		&resource.ID,
		&resource.Name,
		&resource.Type,
		&resource.ReservationMode,
		&resource.ImageURL,
		&capacity,
		&resource.IsActive,
	)

	if err != nil {
		return models.Resource{}, err
	}

	if capacity.Valid {
		value := int(capacity.Int64)
		resource.Capacity = &value
	}

	if resource.IsActive {
		resource.Status = "available"
	} else {
		resource.Status = "maintenance"
	}

	return resource, nil
}

func UpdateResourceImageURL(id int, imageURL string) (models.Resource, error) {
	_, err := database.DB.ExecContext(
		context.Background(),
		`
		UPDATE dbo.resources
		SET
			image_url = NULLIF(@p2, ''),
			updated_at = SYSUTCDATETIME()
		WHERE id = @p1;
		`,
		id,
		imageURL,
	)

	if err != nil {
		return models.Resource{}, err
	}

	return GetResourceByID(id)
}
