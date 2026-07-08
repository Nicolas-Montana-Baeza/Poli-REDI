package handlers

import (
	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func GetWorkshops(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "usuario no autenticado",
		})
	}

	workshops, err := services.GetWorkshops(user.ID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "No se pudieron cargar los talleres",
			"detail": err.Error(),
		})
	}

	return c.JSON(workshops)
}

func EnrollInWorkshop(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "usuario no autenticado",
		})
	}

	if !user.IsAdmin && user.RUT == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Debes registrar tu RUT antes de inscribirte en talleres.",
		})
	}

	workshopID, err := strconv.Atoi(c.Params("id"))

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Taller invalido",
		})
	}

	workshop, err := services.EnrollInWorkshop(
		workshopID,
		user,
	)

	if err != nil {
		status := fiber.StatusBadRequest

		switch err.Error() {
		case "taller no encontrado o no disponible":
			status = fiber.StatusNotFound
		case "el taller no tiene cupos disponibles":
			status = fiber.StatusConflict
		case "ya estas inscrito en este taller":
			status = fiber.StatusConflict
		}

		return c.Status(status).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(workshop)
}
