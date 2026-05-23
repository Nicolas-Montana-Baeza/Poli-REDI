package handlers

import (
	"poli-redi-api/internal/repositories"

	"github.com/gofiber/fiber/v2"
)

func GetResources(c *fiber.Ctx) error {
	resources := repositories.GetAllResources()

	return c.JSON(resources)
}
