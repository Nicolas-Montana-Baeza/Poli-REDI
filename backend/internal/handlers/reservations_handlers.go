package handlers

import (
	"errors"
	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
	"poli-redi-api/internal/services"
	"poli-redi-api/internal/validators"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func GetReservations(c *fiber.Ctx) error {
	reservations, err := services.GetReservations()

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "No se pudieron cargar las reservas",
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
			"error": "No se pudo cargar la disponibilidad",
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
			"error": "No se pudieron cargar tus reservas",
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
		status, message := createReservationErrorResponse(err)
		return c.Status(status).JSON(fiber.Map{
			"error": message,
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
		status, message := cancelReservationErrorResponse(err)
		return c.Status(status).JSON(fiber.Map{
			"error": message,
		})
	}

	return c.JSON(cancelledReservation)
}

func createReservationErrorResponse(err error) (int, string) {
	switch {
	case errors.Is(err, repositories.ErrParticipantConflict):
		return fiber.StatusConflict, repositories.ErrParticipantConflict.Error()
	case errors.Is(err, repositories.ErrResourceNotAllowedByPolicy):
		return fiber.StatusBadRequest, "la instalacion no esta permitida por la politica vigente"
	case errors.Is(err, repositories.ErrTargetForNonGroup):
		return fiber.StatusBadRequest, "el objetivo de participantes solo se permite en reservas grupales"
	case errors.Is(err, repositories.ErrInvalidTargetParticipants):
		return fiber.StatusBadRequest, "el objetivo de participantes debe estar entre el minimo y la capacidad"
	}

	message := err.Error()
	for _, prefix := range []string{
		"selecciona ",
		"no puedes ",
		"el recurso ",
		"la capacidad ",
		"la hora ",
		"la jornada ",
		"la reserva debe ",
		"la fecha debe ",
		"A\u00fan no puedes ",
		"no existe una pol\u00edtica ",
		"debes registrar tu RUT ",
		"el usuario se encuentra bloqueado ",
		"usuario, recurso o actividad ",
		"ya existe un registro ",
	} {
		if strings.HasPrefix(message, prefix) {
			return fiber.StatusBadRequest, message
		}
	}
	return fiber.StatusInternalServerError, "No se pudo crear la reserva"
}

func cancelReservationErrorResponse(err error) (int, string) {
	message := err.Error()
	switch message {
	case "no tienes permisos para cancelar esta reserva":
		return fiber.StatusForbidden, message
	case "reserva no encontrada":
		return fiber.StatusNotFound, message
	case "no se pudo identificar la reserva",
		"usuario autenticado es obligatorio",
		"la reserva ya no se puede cancelar",
		"la reserva ya est\u00e1 cancelada",
		"la reserva en este estado no se puede cancelar",
		"no puedes cancelar una reserva finalizada":
		return fiber.StatusBadRequest, message
	default:
		return fiber.StatusInternalServerError, "No se pudo cancelar la reserva"
	}
}
