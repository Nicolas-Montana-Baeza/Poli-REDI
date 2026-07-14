package handlers

import "github.com/gofiber/fiber/v2"

func GetRoot(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"service": "Poli-REDI API",
		"status":  "ok",
		"health":  "/api/health",
	})
}

func GetHealth(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "Poli-REDI API funcionando",
	})
}
