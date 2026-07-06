package handlers

import (
	"poli-redi-api/internal/repositories"

	"github.com/gofiber/fiber/v2"
)

func GetActivities(c *fiber.Ctx) error {
	activities, err := repositories.GetActiveActivities()

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  "Error obteniendo actividades",
			"detail": err.Error(),
		})
	}

	return c.JSON(activities)
}
