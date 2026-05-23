package handlers

import (
	"poli-redi-api/internal/repositories"

	"github.com/gofiber/fiber/v2"
)

func GetResources(c *fiber.Ctx) error {
	resources, err := repositories.GetAllResources()

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  "Error obteniendo recursos",
			"detail": err.Error(),
		})
	}

	return c.JSON(resources)
}
