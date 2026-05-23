package handlers

import (
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/services"

	"github.com/gofiber/fiber/v2"
)

func GetReservations(c *fiber.Ctx) error {
	reservations := services.GetReservations()

	return c.JSON(reservations)
}

func CreateReservation(c *fiber.Ctx) error {
	var reservation models.Reservation

	if err := c.BodyParser(&reservation); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Datos inválidos",
		})
	}

	createdReservation, err := services.CreateReservation(reservation)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(201).JSON(createdReservation)
}
