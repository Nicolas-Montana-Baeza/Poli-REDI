package handlers

import (
	"errors"
	"strings"
	"unicode/utf8"

	"poli-redi-api/internal/repositories"

	"github.com/gofiber/fiber/v2"
)

type createAdminResourceRequest struct {
	VenueID         int    `json:"venueId"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	ReservationMode string `json:"reservationMode"`
	ImageURL        string `json:"imageUrl"`
	Capacity        *int   `json:"capacity"`
	IsActive        *bool  `json:"isActive"`
}

func CreateAdminResource(c *fiber.Ctx) error {
	var request createAdminResourceRequest

	if err := decodeStrictJSON(c.Body(), &request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "datos invalidos o campos no permitidos"},
		)
	}

	name := strings.TrimSpace(request.Name)
	resourceType := strings.TrimSpace(request.Type)
	mode := strings.ToUpper(strings.TrimSpace(request.ReservationMode))
	imageURL := strings.TrimSpace(request.ImageURL)

	if request.VenueID <= 0 ||
		name == "" ||
		resourceType == "" ||
		utf8.RuneCountInString(name) > 120 ||
		utf8.RuneCountInString(resourceType) > 80 {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "sede, nombre o tipo invalido"},
		)
	}

	if mode == "" {
		mode = "RESERVABLE"
	}
	if !validResourceReservationMode(mode) {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "modo de reserva invalido"},
		)
	}

	if request.Capacity != nil && *request.Capacity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "capacidad invalida"},
		)
	}

	if len(imageURL) > 500 ||
		(imageURL != "" &&
			!strings.HasPrefix(imageURL, "https://") &&
			!strings.HasPrefix(imageURL, "http://") &&
			!strings.HasPrefix(imageURL, "/")) {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "URL de imagen invalida"},
		)
	}

	active := true
	if request.IsActive != nil {
		active = *request.IsActive
	}

	resource, err := repositories.CreateResource(
		repositories.CreateResourceInput{
			VenueID:         request.VenueID,
			Name:            name,
			Type:            resourceType,
			ReservationMode: mode,
			ImageURL:        imageURL,
			Capacity:        request.Capacity,
			IsActive:        active,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrVenueNotFound):
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": err.Error()},
			)

		case errors.Is(err, repositories.ErrResourceDuplicate),
			errors.Is(err, repositories.ErrResourceVenueInactive),
			errors.Is(err, repositories.ErrResourceConcurrentUpdate):
			return c.Status(fiber.StatusConflict).JSON(
				fiber.Map{"error": err.Error()},
			)

		default:
			return c.Status(fiber.StatusInternalServerError).JSON(
				fiber.Map{"error": "no se pudo crear el recurso"},
			)
		}
	}

	return c.Status(fiber.StatusCreated).JSON(resource)
}
