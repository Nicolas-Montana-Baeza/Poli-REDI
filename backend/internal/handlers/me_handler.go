package handlers

import (
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

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Datos invalidos",
		})
	}

	normalizedRUT := validators.NormalizeRUT(request.RUT)

	if normalizedRUT == "" && user.IsAdmin {
		updatedUser, err := repositories.UpdateUserRUT(user.ID, "")

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "no se pudo actualizar el RUT",
			})
		}

		return c.JSON(updatedUser)
	}

	if !validators.IsValidRUT(normalizedRUT) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "RUT invalido",
		})
	}

	updatedUser, err := repositories.UpdateUserRUT(user.ID, normalizedRUT)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no se pudo actualizar el RUT",
		})
	}

	return c.JSON(updatedUser)
}
