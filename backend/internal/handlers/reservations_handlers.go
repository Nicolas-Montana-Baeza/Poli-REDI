package handlers

import (
	"errors"
	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/services"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetReservations(c *fiber.Ctx) error {
	reservations, err := services.GetReservations()

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  "No se pudieron cargar las reservas",
			"detail": err.Error(),
		})
	}

	return c.JSON(reservations)
}

func GetAvailabilityReservations(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)

	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "usuario no autenticado",
		})
	}

	reservations, err := services.GetReservations()

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  "No se pudo cargar la disponibilidad",
			"detail": err.Error(),
		})
	}

	if !user.IsAdmin {
		for index := range reservations {
			reservations[index].UserID = 0
			reservations[index].UserFullName = ""
			reservations[index].UserEmail = ""
			reservations[index].UserRUT = ""
		}
	}

	return c.JSON(reservations)
}

func GetMyReservations(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)

	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "usuario no autenticado",
		})
	}

	reservations, err := services.GetMyReservations(user.ID)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  "No se pudieron cargar tus reservas",
			"detail": err.Error(),
		})
	}

	return c.JSON(reservations)
}

func CreateReservation(c *fiber.Ctx) error {
	var request models.CreateReservationRequest

	user, ok := middleware.GetLocalUser(c)

	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "usuario no autenticado",
		})
	}

	if !user.IsAdmin && user.RUT == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Debes registrar tu RUT antes de crear reservas.",
		})
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Datos inválidos",
		})
	}

	startTime, err := parseReservationStartTime(request.StartTime)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Fecha de inicio inválida",
		})
	}

	reservation := models.Reservation{
		UserID:          user.ID,
		ResourceID:      request.ResourceID,
		ActivityID:      request.ActivityID,
		StartTime:       startTime,
		DurationMinutes: request.DurationMinutes,
		Status:          request.Status,
	}

	createdReservation, err := services.CreateReservation(reservation)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(201).JSON(createdReservation)
}

func parseReservationStartTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return time.Time{}, errors.New("startTime es obligatorio")
	}

	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return parsed.In(time.Local), nil
	}

	layouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
	}

	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, errors.New("fecha de inicio inválida")
}

func CancelReservation(c *fiber.Ctx) error {
	var request models.CancelReservationRequest

	user, ok := middleware.GetLocalUser(c)

	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "usuario no autenticado",
		})
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Datos inválidos",
		})
	}

	cancelledReservation, err := services.CancelReservation(
		request.ReservationID,
		user,
	)

	if err != nil {
		status := 400

		switch err.Error() {
		case "no tienes permisos para cancelar esta reserva":
			status = 403
		case "reserva no encontrada":
			status = 404
		}

		return c.Status(status).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(cancelledReservation)
}
