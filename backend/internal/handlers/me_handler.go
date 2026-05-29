package handlers

import (
	"poli-redi-api/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func GetMe(c *fiber.Ctx) error {
	authUser, ok := middleware.GetAuthUser(c)

	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "usuario no autenticado",
		})
	}

	return c.JSON(fiber.Map{
		"oid":    authUser.OID,
		"name":   authUser.Name,
		"email":  authUser.Email,
		"tenant": authUser.Tenant,
	})
}
