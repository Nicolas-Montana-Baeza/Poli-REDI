package repositories

import (
	"context"
	"database/sql"
	"errors"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

var (
	ErrResourceNotFound = errors.New("recurso no encontrado")

	ErrGroupResourceMustRemainReservable = errors.New(
		"un recurso con confirmacion grupal vigente debe permanecer RESERVABLE",
	)

	ErrGroupResourceCapacityInvalid = errors.New(
		"la capacidad del recurso no puede ser menor al minimo grupal vigente",
	)
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
		FROM resources
		WHERE id = $1;
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
	result, err := database.DB.ExecContext(
		context.Background(),
		`
		UPDATE resources
		SET
			image_url = NULLIF($2, ''),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1;
		`,
		id,
		imageURL,
	)

	if err != nil {
		return models.Resource{}, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return models.Resource{}, err
	}

	if affected == 0 {
		return models.Resource{}, ErrResourceNotFound
	}

	return GetResourceByID(id)
}

func validateResourceGroupInvariant(
	groupMinimum sql.NullInt64,
	reservationMode string,
	capacity sql.NullInt64,
) error {
	if !groupMinimum.Valid {
		return nil
	}

	if reservationMode != "RESERVABLE" {
		return ErrGroupResourceMustRemainReservable
	}

	if !capacity.Valid ||
		capacity.Int64 < groupMinimum.Int64 {

		return ErrGroupResourceCapacityInvalid
	}

	return nil
}

func UpdateResourceOperationalSettings(
	id int,
	reservationMode *string,
	capacityPresent bool,
	capacity *int,
	isActive *bool,
) (models.Resource, error) {
	ctx := context.Background()

	tx, err := database.DB.BeginTx(
		ctx,
		&sql.TxOptions{
			Isolation: sql.LevelSerializable,
		},
	)
	if err != nil {
		return models.Resource{}, err
	}

	defer tx.Rollback()

	var (
		currentMode     string
		currentCapacity sql.NullInt64
	)

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT
			reservation_mode,
			capacity
		FROM resources
		WHERE id = $1
		FOR UPDATE
		`,
		id,
	).Scan(
		&currentMode,
		&currentCapacity,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return models.Resource{}, ErrResourceNotFound
	}

	if err != nil {
		return models.Resource{}, err
	}

	effectiveMode := currentMode

	if reservationMode != nil {
		effectiveMode = *reservationMode
	}

	effectiveCapacity := currentCapacity

	if capacityPresent {
		if capacity == nil {
			effectiveCapacity = sql.NullInt64{}
		} else {
			effectiveCapacity = sql.NullInt64{
				Int64: int64(*capacity),
				Valid: true,
			}
		}
	}

	// ------------------------------------------------------------
	// Política grupal vigente.
	// ------------------------------------------------------------
	//
	// Una reserva grupal utiliza resources.capacity como límite y snapshot.
	// Por eso la administración no puede dejar el recurso en un estado que
	// vuelva inválidas las nuevas solicitudes de la política vigente.

	var groupMinimum sql.NullInt64

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT
			group_resource.minimum_participants
		FROM reservation_policies policy
		INNER JOIN reservation_policy_group_resources group_resource
			ON group_resource.policy_id = policy.id
		WHERE policy.is_published = true
		  AND policy.effective_from <= CURRENT_TIMESTAMP
		  AND (
			policy.effective_to IS NULL
			OR policy.effective_to > CURRENT_TIMESTAMP
		  )
		  AND group_resource.resource_id = $1
		ORDER BY
			policy.effective_from DESC,
			policy.id DESC
		LIMIT 1
		`,
		id,
	).Scan(&groupMinimum)

	if err != nil &&
		!errors.Is(err, sql.ErrNoRows) {

		return models.Resource{}, err
	}

	if err := validateResourceGroupInvariant(
		groupMinimum,
		effectiveMode,
		effectiveCapacity,
	); err != nil {
		return models.Resource{}, err
	}

	modeValue := ""
	if reservationMode != nil {
		modeValue = *reservationMode
	}

	activeValue := false
	if isActive != nil {
		activeValue = *isActive
	}

	var capacityValue any
	if capacity != nil {
		capacityValue = *capacity
	}

	result, err := tx.ExecContext(
		ctx,
		`
		UPDATE resources
		SET
			reservation_mode = CASE
				WHEN $2 THEN $3
				ELSE reservation_mode
			END,
			capacity = CASE
				WHEN $4 THEN $5
				ELSE capacity
			END,
			is_active = CASE
				WHEN $6 THEN $7
				ELSE is_active
			END,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		`,
		id,
		reservationMode != nil,
		modeValue,
		capacityPresent,
		capacityValue,
		isActive != nil,
		activeValue,
	)

	if err != nil {
		return models.Resource{}, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return models.Resource{}, err
	}

	if affected != 1 {
		return models.Resource{}, ErrResourceNotFound
	}

	if err := tx.Commit(); err != nil {
		return models.Resource{}, err
	}

	return GetResourceByID(id)
}
