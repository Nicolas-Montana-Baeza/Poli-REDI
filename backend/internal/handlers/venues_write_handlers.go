package handlers

import (
	"errors"
	"strings"
	"unicode/utf8"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"

	"github.com/gofiber/fiber/v2"
)

type createAdminVenueRequest struct {
	Name        string   `json:"name"`
	AddressLine *string  `json:"addressLine"`
	Commune     *string  `json:"commune"`
	City        *string  `json:"city"`
	Region      *string  `json:"region"`
	Country     string   `json:"country"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
}

func normalizeVenueText(value *string, maxLength int) (
	*string, bool,
) {
	if value == nil {
		return nil, true
	}

	normalized := strings.TrimSpace(*value)

	if utf8.RuneCountInString(normalized) > maxLength {
		return nil, false
	}

	if normalized == "" {
		return nil, true
	}

	return &normalized, true
}

func CreateAdminVenue(c *fiber.Ctx) error {
	var request createAdminVenueRequest

	if err := decodeStrictJSON(c.Body(), &request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "datos invalidos o campos no permitidos"},
		)
	}

	name := strings.TrimSpace(request.Name)
	country := strings.TrimSpace(request.Country)

	if name == "" || country == "" ||
		utf8.RuneCountInString(name) > 150 ||
		utf8.RuneCountInString(country) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "nombre o pais invalido"},
		)
	}

	address, validAddress := normalizeVenueText(
		request.AddressLine, 200,
	)
	commune, validCommune := normalizeVenueText(
		request.Commune, 100,
	)
	city, validCity := normalizeVenueText(
		request.City, 100,
	)
	region, validRegion := normalizeVenueText(
		request.Region, 100,
	)

	if !validAddress || !validCommune ||
		!validCity || !validRegion {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "ubicacion demasiado larga"},
		)
	}

	if request.Latitude != nil &&
		(*request.Latitude < -90 || *request.Latitude > 90) {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "latitud fuera de rango"},
		)
	}

	if request.Longitude != nil &&
		(*request.Longitude < -180 || *request.Longitude > 180) {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "longitud fuera de rango"},
		)
	}

	venue, err := repositories.CreateVenue(models.Venue{
		Name:        name,
		AddressLine: address,
		Commune:     commune,
		City:        city,
		Region:      region,
		Country:     country,
		Latitude:    request.Latitude,
		Longitude:   request.Longitude,
		IsActive:    true,
	})

	if errors.Is(err, repositories.ErrVenueDuplicate) {
		return c.Status(fiber.StatusConflict).JSON(
			fiber.Map{"error": err.Error()},
		)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "no se pudo crear la sede"},
		)
	}

	return c.Status(fiber.StatusCreated).JSON(venue)
}
