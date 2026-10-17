package repositories

import (
	"context"
	"errors"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"

	"github.com/jackc/pgx/v5/pgconn"
)

var ErrVenueDuplicate = errors.New(
	"ya existe una sede con ese nombre",
)

func CreateVenue(venue models.Venue) (models.Venue, error) {
	err := database.DB.QueryRowContext(
		context.Background(),
		`
        INSERT INTO venues (
            name,
            address_line,
            commune,
            city,
            region,
            country,
            latitude,
            longitude,
            is_active
        )
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,true)
        RETURNING id
        `,
		venue.Name,
		venue.AddressLine,
		venue.Commune,
		venue.City,
		venue.Region,
		venue.Country,
		venue.Latitude,
		venue.Longitude,
	).Scan(&venue.ID)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "uq_venues_name" {
			return models.Venue{}, ErrVenueDuplicate
		}

		return models.Venue{}, err
	}

	venue.IsActive = true
	return venue, nil
}
