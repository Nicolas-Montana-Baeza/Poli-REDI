package repositories

import (
	"context"
	"database/sql"
	"errors"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrVenueNotFound = errors.New("sede no encontrada")

	ErrVenueHasActiveResources = errors.New(
		"desactiva primero los recursos activos de la sede",
	)

	ErrVenueConcurrentUpdate = errors.New(
		"la sede cambio durante la operacion; vuelve a intentar",
	)
)

type VenuePatch struct {
	Name               *string
	AddressLine        *string
	AddressLinePresent bool
	Commune            *string
	CommunePresent     bool
	City               *string
	CityPresent        bool
	Region             *string
	RegionPresent      bool
	Country            *string
	Latitude           *float64
	LatitudePresent    bool
	Longitude          *float64
	LongitudePresent   bool
	IsActive           *bool
}

func venuePatchDBError(err error) error {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" &&
			pgErr.ConstraintName == "uq_venues_name" {
			return ErrVenueDuplicate
		}

		if pgErr.Code == "40001" || pgErr.Code == "40P01" {
			return ErrVenueConcurrentUpdate
		}
	}

	return err
}

func GetVenueByID(id int) (models.Venue, error) {
	venues, err := GetAllVenues()
	if err != nil {
		return models.Venue{}, err
	}

	for _, venue := range venues {
		if venue.ID == id {
			return venue, nil
		}
	}

	return models.Venue{}, ErrVenueNotFound
}

func UpdateVenue(
	id int,
	patch VenuePatch,
) (models.Venue, error) {
	ctx := context.Background()

	tx, err := database.DB.BeginTx(
		ctx,
		&sql.TxOptions{
			Isolation: sql.LevelSerializable,
		},
	)
	if err != nil {
		return models.Venue{}, venuePatchDBError(err)
	}
	defer tx.Rollback()

	var currentlyActive bool

	err = tx.QueryRowContext(ctx, `
        SELECT is_active
        FROM venues
        WHERE id = $1
        FOR UPDATE
    `, id).Scan(&currentlyActive)

	if errors.Is(err, sql.ErrNoRows) {
		return models.Venue{}, ErrVenueNotFound
	}
	if err != nil {
		return models.Venue{}, venuePatchDBError(err)
	}

	if patch.IsActive != nil && !*patch.IsActive {
		var hasActiveResources bool

		err = tx.QueryRowContext(ctx, `
            SELECT EXISTS (
                SELECT 1
                FROM resources
                WHERE venue_id = $1
                  AND is_active = true
            )
        `, id).Scan(&hasActiveResources)

		if err != nil {
			return models.Venue{}, venuePatchDBError(err)
		}

		if hasActiveResources {
			return models.Venue{}, ErrVenueHasActiveResources
		}
	}

	_, err = tx.ExecContext(ctx, `
        UPDATE venues
        SET
            name = COALESCE($2::text, name),
            address_line = CASE
                WHEN $3::boolean THEN $4::text
                ELSE address_line
            END,
            commune = CASE
                WHEN $5::boolean THEN $6::text
                ELSE commune
            END,
            city = CASE
                WHEN $7::boolean THEN $8::text
                ELSE city
            END,
            region = CASE
                WHEN $9::boolean THEN $10::text
                ELSE region
            END,
            country = COALESCE($11::text, country),
            latitude = CASE
                WHEN $12::boolean THEN $13::numeric
                ELSE latitude
            END,
            longitude = CASE
                WHEN $14::boolean THEN $15::numeric
                ELSE longitude
            END,
            is_active = COALESCE($16::boolean, is_active),
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1
    `,
		id,
		patch.Name,
		patch.AddressLinePresent,
		patch.AddressLine,
		patch.CommunePresent,
		patch.Commune,
		patch.CityPresent,
		patch.City,
		patch.RegionPresent,
		patch.Region,
		patch.Country,
		patch.LatitudePresent,
		patch.Latitude,
		patch.LongitudePresent,
		patch.Longitude,
		patch.IsActive,
	)

	if err != nil {
		return models.Venue{}, venuePatchDBError(err)
	}

	if err := tx.Commit(); err != nil {
		return models.Venue{}, venuePatchDBError(err)
	}

	return GetVenueByID(id)
}
