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
	"time"

	"github.com/gofiber/fiber/v2"
)

var getAvailabilityItems = services.GetAvailabilityItems
var getAvailabilityItemsForRange = services.GetAvailabilityItemsForRange

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

	fromRaw := strings.TrimSpace(c.Query("from"))
	toRaw := strings.TrimSpace(c.Query("to"))

	var items []models.AvailabilityItem
	var err error
	if fromRaw == "" && toRaw == "" {
		items, err = getAvailabilityItems()
	} else {
		from, toExclusive, validationCode, validationMessage := parseAvailabilityRange(fromRaw, toRaw)
		if validationCode != "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":  validationCode,
				"error": validationMessage,
			})
		}
		items, err = getAvailabilityItemsForRange(from, toExclusive, user.ID)
	}

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
		items = services.SanitizeAvailabilityItemsForAudience(items, user)
	}

	return c.JSON(items)
}

func parseAvailabilityRange(fromRaw, toRaw string) (
	time.Time,
	time.Time,
	string,
	string,
) {
	if fromRaw == "" || toRaw == "" {
		return time.Time{}, time.Time{},
			"AVAILABILITY_RANGE_REQUIRED",
			"Debes indicar from y to juntos."
	}

	const dateLayout = "2006-01-02"
	from, fromErr := time.ParseInLocation(dateLayout, fromRaw, businessclock.Location())
	to, toErr := time.ParseInLocation(dateLayout, toRaw, businessclock.Location())
	if fromErr != nil || toErr != nil || from.After(to) {
		return time.Time{}, time.Time{},
			"AVAILABILITY_RANGE_INVALID",
			"from y to deben usar el formato YYYY-MM-DD y formar un rango valido."
	}

	if to.After(from.AddDate(0, 0, 30)) {
		return time.Time{}, time.Time{},
			"AVAILABILITY_RANGE_TOO_LARGE",
			"El rango de disponibilidad no puede superar 31 dias."
	}

	return from, to.AddDate(0, 0, 1), "", ""
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
