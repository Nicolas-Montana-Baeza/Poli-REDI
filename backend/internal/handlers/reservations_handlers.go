package handlers

import (
	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/services"

	"github.com/gofiber/fiber/v2"
)

func GetReservations(c *fiber.Ctx) error {
	reservations, err := services.GetReservations()

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  "Error obteniendo reservas",
			"detail": err.Error(),
		})
	}

	return c.JSON(reservations)
}

func CreateReservation(c *fiber.Ctx) error {
	var reservation models.Reservation

	user, ok := middleware.GetLocalUser(c)

	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "usuario no autenticado",
		})
	}

	if err := c.BodyParser(&reservation); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Datos invalidos",
		})
	}

	reservation.UserID = user.ID

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

	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Datos invalidos",
		})
	}

	cancelledReservation, err := services.CancelReservation(
		request.ReservationID,
		request.RequestedByUserID,
	)

	if err != nil {
		status := 400

		if err.Error() == "no tienes permisos para cancelar esta reserva" {
			status = 403
		}

		return c.Status(status).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(cancelledReservation)
}
