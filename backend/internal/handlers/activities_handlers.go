package handlers

import (
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func GetActivities(c *fiber.Ctx) error {
	activities, err := repositories.GetActiveActivities()

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  "Error obteniendo actividades",
			"detail": err.Error(),
		})
	}

	return c.JSON(activities)
}

func CreateActivity(c *fiber.Ctx) error {
	var request models.CreateActivityRequest

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Datos invalidos",
		})
	}

	name := strings.TrimSpace(request.Name)

	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Ingresa el nombre de la actividad",
		})
	}

	if len([]rune(name)) > 120 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "La actividad no puede superar 120 caracteres",
		})
	}

	activity, err := repositories.GetOrCreateActivityByName(name, request.Description)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "No se pudo crear la actividad",
			"detail": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(activity)
}
