package handlers

import (
	"errors"
	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
	"poli-redi-api/internal/validators"

	"github.com/gofiber/fiber/v2"
)

func GetMe(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "usuario no autenticado",
		})
	}

	return c.JSON(user)
}

func UpdateMeRUT(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "usuario no autenticado",
		})
	}

	var request models.UpdateRUTRequest

	if err := decodeStrictJSON(c.Body(), &request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Datos inválidos",
		})
	}

	normalizedRUT := validators.NormalizeRUT(request.RUT)

	if !validators.IsValidRUT(normalizedRUT) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "RUT inválido",
		})
	}

	updatedUser, err := repositories.UpdateUserRUT(user.ID, normalizedRUT)

	if err != nil {
		if errors.Is(err, repositories.ErrRUTInvalid) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		if errors.Is(err, repositories.ErrRUTAlreadySet) || errors.Is(err, repositories.ErrRUTDuplicate) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no se pudo actualizar el RUT"})
	}

	return c.JSON(updatedUser)
}
