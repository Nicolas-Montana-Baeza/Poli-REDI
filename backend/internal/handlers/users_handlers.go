package handlers

import (
	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/repositories"

	"github.com/gofiber/fiber/v2"
)

func GetUsers(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "usuario no autenticado",
		})
	}

	if !user.IsAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "no tienes permisos para ver usuarios",
		})
	}

	users, err := repositories.GetAllUsers()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Error obteniendo usuarios",
			"detail": err.Error(),
		})
	}

	return c.JSON(users)
}
