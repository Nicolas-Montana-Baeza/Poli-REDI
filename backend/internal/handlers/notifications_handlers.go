package handlers

import (
	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/repositories"

	"github.com/gofiber/fiber/v2"
)

func GetNotifications(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)

	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "usuario no autenticado",
		})
	}

	notifications, err := repositories.GetNotificationsByUserID(user.ID)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error obteniendo notificaciones",
		})
	}

	return c.JSON(notifications)
}
