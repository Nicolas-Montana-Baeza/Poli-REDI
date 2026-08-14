package handlers

import (
	"errors"
	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/services"
	"strconv"
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

	from, to, err := availabilityRange(c.Query("from"), c.Query("to"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	items, err := services.GetAvailabilityItems(from, to, user.ID, user.IsAdmin)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  "No se pudo cargar la disponibilidad",
			"detail": err.Error(),
		})
	}

	if !user.IsAdmin {
		// La disponibilidad es compartida, pero los datos personales de reservas
		// no. Admin ve detalle operacional; usuario normal solo ve ocupacion y
		// bloques institucionales necesarios para decidir disponibilidad.
		for index := range items {
			if items[index].UserID != user.ID {
				items[index].UserID = 0
				items[index].UserFullName = ""
				items[index].UserEmail = ""
				items[index].UserRUT = ""
				if items[index].Type != "blocked" {
					items[index].Title = "Ocupado"
				}
			}
		}
	}

	return c.JSON(items)
}

func availabilityRange(fromValue string, toValue string) (time.Time, time.Time, error) {
	location := businessclock.Location()
	now := businessclock.Now()
	defaultFrom := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	defaultTo := defaultFrom.AddDate(0, 0, 15)

	from, err := parseRangeBoundary(fromValue, defaultFrom, location)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("from debe usar YYYY-MM-DD o RFC3339")
	}
	to, err := parseRangeBoundary(toValue, defaultTo, location)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("to debe usar YYYY-MM-DD o RFC3339")
	}
	if !to.After(from) {
		return time.Time{}, time.Time{}, errors.New("to debe ser posterior a from")
	}
	if to.After(from.AddDate(0, 0, 31)) {
		return time.Time{}, time.Time{}, errors.New("el rango de disponibilidad no puede superar 31 dias")
	}
	return from, to, nil
}

func parseRangeBoundary(value string, fallback time.Time, location *time.Location) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}
	if parsed, err := time.ParseInLocation("2006-01-02", value, location); err == nil {
		return parsed, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.In(location), nil
}

func GetReservationDetail(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "usuario no autenticado"})
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "reserva invalida"})
	}
	reservation, err := services.GetReservationDetail(id, user)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrReservationNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		case errors.Is(err, services.ErrReservationForbidden):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no se pudo cargar la reserva"})
		}
	}
	return c.JSON(reservation)
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

	if err := decodeStrictJSON(c.Body(), &request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Datos inválidos o campos no permitidos",
		})
	}

	startTime, err := businessclock.ParseDateTime(request.StartTime)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Fecha de inicio inválida",
		})
	}

	reservation := models.Reservation{
		// El usuario local autenticado es el unico propietario aceptado. Cualquier
		// userId/status controlado por cliente se rechaza antes de este punto.
		UserID:          user.ID,
		ResourceID:      request.ResourceID,
		ActivityID:      request.ActivityID,
		StartTime:       startTime,
		DurationMinutes: request.DurationMinutes,
	}

	createdReservation, err := services.CreateReservation(reservation)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(201).JSON(createdReservation)
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
