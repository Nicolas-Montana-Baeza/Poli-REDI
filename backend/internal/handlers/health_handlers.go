package handlers

import "github.com/gofiber/fiber/v2"

func GetHealth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "Poli-REDI API funcionando",
	})
}
