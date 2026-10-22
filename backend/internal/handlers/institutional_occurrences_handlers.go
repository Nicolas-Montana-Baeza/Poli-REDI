package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/repositories"
	"poli-redi-api/internal/services"
)

func occurrenceActivityID(c *fiber.Ctx) (int, error) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return 0, repositories.ErrInstitutionalOccurrenceInvalid
	}
	return id, nil
}

func institutionalOccurrenceHTTPError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, services.ErrInstitutionalUnauthorized):
		return c.Status(403).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, services.ErrInstitutionalUnitInactive):
		return c.Status(409).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, services.ErrInstitutionalActivityNotFound):
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, repositories.ErrInstitutionalOccurrenceInvalid), errors.Is(err, services.ErrInstitutionalScheduleInvalid):
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, repositories.ErrInstitutionalOccurrenceAlreadyHandled),
		errors.Is(err, repositories.ErrInstitutionalOccurrenceEnrolled),
		errors.Is(err, repositories.ErrInstitutionalOccurrenceBlocked),
		errors.Is(err, repositories.ErrInstitutionalOccurrenceSelfOverlap),
		errors.Is(err, repositories.ErrInstitutionalActivityAlreadyCancelled):
		return c.Status(409).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(500).JSON(fiber.Map{"error": "no se pudo gestionar esta sesión"})
	}
}

func GetInstitutionalOccurrenceExceptions(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "usuario no autenticado"})
	}
	id, err := occurrenceActivityID(c)
	if err != nil {
		return institutionalOccurrenceHTTPError(c, err)
	}
	items, err := services.GetInstitutionalOccurrenceExceptions(user, id)
	if err != nil {
		return institutionalOccurrenceHTTPError(c, err)
	}
	return c.JSON(items)
}

func PatchInstitutionalOccurrence(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "usuario no autenticado"})
	}
	id, err := occurrenceActivityID(c)
	if err != nil {
		return institutionalOccurrenceHTTPError(c, err)
	}
	var req services.InstitutionalOccurrenceRequest
	if err := decodeStrictJSON(c.Body(), &req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "datos inválidos o campos no permitidos"})
	}
	result, err := services.PatchInstitutionalOccurrence(user, id, req)
	if err != nil {
		return institutionalOccurrenceHTTPError(c, err)
	}
	return c.JSON(result)
}
