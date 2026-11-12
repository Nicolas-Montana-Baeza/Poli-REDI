package handlers

import (
	"poli-redi-api/internal/repositories"

	"github.com/gofiber/fiber/v2"
)

func GetAdminVenues(c *fiber.Ctx) error {
	venues, err := repositories.GetAllVenues()

	if err != nil {
		sec005LogInternalFailure("GetAdminVenues", err)
		return sec005WritePublicError(c, fiber.StatusInternalServerError, "no se pudieron cargar las sedes")
	}

	return c.JSON(venues)
}
