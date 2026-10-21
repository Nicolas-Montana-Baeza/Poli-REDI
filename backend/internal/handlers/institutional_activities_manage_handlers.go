package handlers

import (
	"github.com/gofiber/fiber/v2"
	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/services"
	"strconv"
)

func UpdateInstitutionalActivity(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "usuario no autenticado"})
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "actividad inválida"})
	}
	var request models.CreateInstitutionalActivityRequest
	if err = decodeStrictJSON(c.Body(), &request); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "datos inválidos o campos no permitidos"})
	}
	result, err := services.UpdateInstitutionalActivity(user, id, request)
	if err != nil {
		return institutionalActivityHTTPError(c, err)
	}
	return c.JSON(result)
}

func CancelInstitutionalActivity(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "usuario no autenticado"})
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "actividad inválida"})
	}
	result, err := services.CancelInstitutionalActivity(user, id)
	if err != nil {
		return institutionalActivityHTTPError(c, err)
	}
	return c.JSON(result)
}
