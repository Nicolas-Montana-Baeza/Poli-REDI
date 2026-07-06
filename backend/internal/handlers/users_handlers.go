package handlers

import (
	"poli-redi-api/internal/repositories"

	"github.com/gofiber/fiber/v2"
)

func GetUsers(c *fiber.Ctx) error {
	users, err := repositories.GetAllUsers()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Error obteniendo usuarios",
			"detail": err.Error(),
		})
	}

	return c.JSON(users)
}
