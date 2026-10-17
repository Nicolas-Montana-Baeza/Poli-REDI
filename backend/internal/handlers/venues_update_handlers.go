package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"

	"poli-redi-api/internal/repositories"

	"github.com/gofiber/fiber/v2"
)

type venuePatchField[T any] struct {
	Set   bool
	Value *T
}

func (f *venuePatchField[T]) UnmarshalJSON(data []byte) error {
	f.Set = true

	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		f.Value = nil
		return nil
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	f.Value = &value
	return nil
}

type updateAdminVenueRequest struct {
	Name        venuePatchField[string]  `json:"name"`
	AddressLine venuePatchField[string]  `json:"addressLine"`
	Commune     venuePatchField[string]  `json:"commune"`
	City        venuePatchField[string]  `json:"city"`
	Region      venuePatchField[string]  `json:"region"`
	Country     venuePatchField[string]  `json:"country"`
	Latitude    venuePatchField[float64] `json:"latitude"`
	Longitude   venuePatchField[float64] `json:"longitude"`
	IsActive    venuePatchField[bool]    `json:"isActive"`
}

func requiredVenuePatchText(
	field venuePatchField[string],
	maxLength int,
) (*string, bool) {
	if !field.Set {
		return nil, true
	}

	if field.Value == nil {
		return nil, false
	}

	value := strings.TrimSpace(*field.Value)

	if value == "" ||
		utf8.RuneCountInString(value) > maxLength {
		return nil, false
	}

	return &value, true
}

func UpdateAdminVenue(c *fiber.Ctx) error {
	venueID, err := strconv.Atoi(c.Params("id"))

	if err != nil || venueID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "sede invalida"},
		)
	}

	var request updateAdminVenueRequest

	if err := decodeStrictJSON(c.Body(), &request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "datos invalidos o campos no permitidos"},
		)
	}

	if !request.Name.Set &&
		!request.AddressLine.Set &&
		!request.Commune.Set &&
		!request.City.Set &&
		!request.Region.Set &&
		!request.Country.Set &&
		!request.Latitude.Set &&
		!request.Longitude.Set &&
		!request.IsActive.Set {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "indica al menos un campo"},
		)
	}

	name, validName := requiredVenuePatchText(request.Name, 150)
	country, validCountry := requiredVenuePatchText(
		request.Country, 100,
	)

	address, validAddress := normalizeVenueText(
		request.AddressLine.Value, 200,
	)
	commune, validCommune := normalizeVenueText(
		request.Commune.Value, 100,
	)
	city, validCity := normalizeVenueText(
		request.City.Value, 100,
	)
	region, validRegion := normalizeVenueText(
		request.Region.Value, 100,
	)

	if !validName || !validCountry ||
		!validAddress || !validCommune ||
		!validCity || !validRegion {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "datos de sede invalidos"},
		)
	}

	if request.IsActive.Set && request.IsActive.Value == nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "isActive debe ser booleano"},
		)
	}

	if request.Latitude.Value != nil {
		latitude := *request.Latitude.Value
		if latitude < -90 || latitude > 90 {
			return c.Status(fiber.StatusBadRequest).JSON(
				fiber.Map{"error": "latitud fuera de rango"},
			)
		}
	}

	if request.Longitude.Value != nil {
		longitude := *request.Longitude.Value
		if longitude < -180 || longitude > 180 {
			return c.Status(fiber.StatusBadRequest).JSON(
				fiber.Map{"error": "longitud fuera de rango"},
			)
		}
	}

	venue, err := repositories.UpdateVenue(
		venueID,
		repositories.VenuePatch{
			Name:               name,
			AddressLine:        address,
			AddressLinePresent: request.AddressLine.Set,
			Commune:            commune,
			CommunePresent:     request.Commune.Set,
			City:               city,
			CityPresent:        request.City.Set,
			Region:             region,
			RegionPresent:      request.Region.Set,
			Country:            country,
			Latitude:           request.Latitude.Value,
			LatitudePresent:    request.Latitude.Set,
			Longitude:          request.Longitude.Value,
			LongitudePresent:   request.Longitude.Set,
			IsActive:           request.IsActive.Value,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrVenueNotFound):
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": err.Error()},
			)

		case errors.Is(err, repositories.ErrVenueDuplicate),
			errors.Is(err, repositories.ErrVenueHasActiveResources),
			errors.Is(err, repositories.ErrVenueConcurrentUpdate):
			return c.Status(fiber.StatusConflict).JSON(
				fiber.Map{"error": err.Error()},
			)

		default:
			return c.Status(
				fiber.StatusInternalServerError,
			).JSON(fiber.Map{
				"error": "no se pudo actualizar la sede",
			})
		}
	}

	return c.JSON(venue)
}
