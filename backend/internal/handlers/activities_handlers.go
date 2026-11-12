package handlers

import (
	"poli-redi-api/internal/repositories"

	"github.com/gofiber/fiber/v2"
)

func GetActivities(c *fiber.Ctx) error {
	activities, err := repositories.GetActiveActivities()

	if err != nil {
		sec005LogInternalFailure("list_activities", err)
		return sec005WritePublicError(c, fiber.StatusInternalServerError, "Error obteniendo actividades")
	}

	return c.JSON(activities)
}
