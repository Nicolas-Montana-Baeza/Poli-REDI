package handlers

import (
	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/services"
	"poli-redi-api/internal/validators"

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

	items, err := services.GetAvailabilityItems()

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  "No se pudo cargar la disponibilidad",
			"detail": err.Error(),
		})
	}

	for index := range items {
		if items[index].UserID == user.ID && items[index].TargetParticipants != nil && items[index].ConfirmationDeadline != nil {
			items[index].CanEditTarget = !businessclock.Now().After(*items[index].ConfirmationDeadline)
		} else {
			items[index].CanEditTarget = false
		}
	}

	if !user.IsAdmin {
		// La disponibilidad es compartida, pero los datos personales de reservas
		// no. Admin ve detalle operacional; usuario normal solo ve ocupacion y
		// bloques institucionales necesarios para decidir disponibilidad.
		for index := range items {
			items[index].UserID = 0
			items[index].UserFullName = ""
			items[index].UserEmail = ""
			items[index].UserRUT = ""

			if items[index].IsScheduledActivity {
				items[index].Title = "Actividad institucional"
			}
		}
	}

	return c.JSON(items)
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

	if err := decodeStrictJSON(c.Body(), &request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Datos inválidos o campos no permitidos",
		})
	}
	if !user.IsAdmin && !validators.HasRUT(user.RUT) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Debes registrar tu RUT antes de crear reservas."})
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
		UserID:             user.ID,
		ResourceID:         request.ResourceID,
		ActivityID:         request.ActivityID,
		StartTime:          startTime,
		DurationMinutes:    request.DurationMinutes,
		TargetParticipants: request.TargetParticipants,
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
