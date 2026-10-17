package handlers

import (
	"poli-redi-api/internal/repositories"

	"github.com/gofiber/fiber/v2"
)

func GetAdminVenues(c *fiber.Ctx) error {
	venues, err := repositories.GetAllVenues()

	if err != nil {
		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(fiber.Map{
			"error": "no se pudieron cargar las sedes",
		})
	}

	return c.JSON(venues)
}
