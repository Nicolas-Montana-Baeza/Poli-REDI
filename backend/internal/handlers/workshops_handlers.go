package handlers

import (
	"errors"
	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/repositories"
	"poli-redi-api/internal/services"
	"poli-redi-api/internal/validators"
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
			"error": "No se pudieron cargar los talleres",
		})
	}

	return c.JSON(workshops)
}

func GetMyWorkshopEnrollments(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "usuario no autenticado",
		})
	}

	enrollments, err := services.GetMyWorkshopEnrollments(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "No se pudo cargar tu historial de talleres.",
		})
	}
	return c.JSON(enrollments)
}

func EnrollInWorkshop(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "usuario no autenticado",
		})
	}

	if !user.IsAdmin && !validators.HasRUT(user.RUT) {
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

	workshop, created, err := services.EnrollInWorkshop(
		workshopID,
		user,
	)

	if err != nil {
		status := fiber.StatusInternalServerError
		message := "No se pudo procesar la inscripción al taller."

		switch {
		case errors.Is(err, repositories.ErrWorkshopNotFound):
			status = fiber.StatusNotFound
			message = "Taller no encontrado o no disponible."
		case errors.Is(err, repositories.ErrWorkshopCapacity):
			status = fiber.StatusConflict
			message = "El taller no tiene cupos disponibles."
		case errors.Is(err, repositories.ErrWorkshopScheduleInvalid):
			status = fiber.StatusConflict
			message = "El taller no tiene un horario válido."
		}

		var conflict *repositories.WorkshopScheduleConflictError
		if errors.As(err, &conflict) {
			status = fiber.StatusConflict
			return c.Status(status).JSON(fiber.Map{
				"code":  "WORKSHOP_SCHEDULE_CONFLICT",
				"error": "El horario se superpone con otro taller en el que ya estás inscrito.",
				"conflict": fiber.Map{
					"title": conflict.Title, "dayText": conflict.DayText,
					"scheduleText": conflict.ScheduleText,
				},
			})
		}

		return c.Status(status).JSON(fiber.Map{
			"error": message,
		})
	}

	if created {
		return c.Status(fiber.StatusCreated).JSON(workshop)
	}
	return c.Status(fiber.StatusOK).JSON(workshop)
}
