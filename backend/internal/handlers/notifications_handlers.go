package handlers

import (
	"errors"
	"strconv"

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
			"error":  "Error obteniendo notificaciones",
			"detail": err.Error(),
		})
	}

	return c.JSON(notifications)
}

func MarkNotificationRead(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "usuario no autenticado",
		})
	}

	notificationID, err := strconv.Atoi(c.Params("id"))

	if err != nil || notificationID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "notificacion invalida",
		})
	}

	notification, err :=
		repositories.MarkNotificationRead(
			user.ID,
			notificationID,
		)

	if errors.Is(
		err,
		repositories.ErrNotificationNotFound,
	) {
		// No distinguimos entre inexistente y perteneciente a otro usuario.
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "notificacion no encontrada",
		})
	}

	if err != nil {
		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(fiber.Map{
			"error": "no se pudo actualizar la notificacion",
		})
	}

	return c.JSON(notification)
}
