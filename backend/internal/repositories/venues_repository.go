package repositories

import (
	"context"
	"database/sql"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func GetAllVenues() ([]models.Venue, error) {
	rows, err := database.DB.QueryContext(
		context.Background(),
		`
		SELECT
			v.id,
			v.name,
			v.address_line,
			v.commune,
			v.city,
			v.region,
			v.country,
			v.latitude,
			v.longitude,
			v.is_active,
			COUNT(r.id),
			COUNT(r.id) FILTER (WHERE r.is_active)
		FROM venues v
		LEFT JOIN resources r ON r.venue_id = v.id
		GROUP BY v.id
		ORDER BY v.name, v.id
		`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	venues := []models.Venue{}

	for rows.Next() {
		var venue models.Venue
		var address, commune, city, region sql.NullString
		var latitude, longitude sql.NullFloat64

		err := rows.Scan(
			&venue.ID,
			&venue.Name,
			&address,
			&commune,
			&city,
			&region,
			&venue.Country,
			&latitude,
			&longitude,
			&venue.IsActive,
			&venue.ResourceCount,
			&venue.ActiveResourceCount,
		)
		if err != nil {
			return nil, err
		}

		if address.Valid {
			venue.AddressLine = &address.String
		}
		if commune.Valid {
			venue.Commune = &commune.String
		}
		if city.Valid {
			venue.City = &city.String
		}
		if region.Valid {
			venue.Region = &region.String
		}
		if latitude.Valid {
			venue.Latitude = &latitude.Float64
		}
		if longitude.Valid {
			venue.Longitude = &longitude.Float64
		}

		venues = append(venues, venue)
	}

	return venues, rows.Err()
}
