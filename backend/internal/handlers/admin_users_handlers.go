package handlers

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
	"poli-redi-api/internal/services"
)

func adminUserAccessError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repositories.ErrUserAccessUnauthorized), errors.Is(err, repositories.ErrUserAccessSelf):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, repositories.ErrUserAccessNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, services.ErrUserAccessInvalidReason), errors.Is(err, services.ErrUserAccessInvalidRequest):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, repositories.ErrUserAccessUnchanged), errors.Is(err, repositories.ErrUserAccessLastAdmin), errors.Is(err, repositories.ErrUserAccessConflict):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no se pudo gestionar el acceso del usuario"})
	}
}

func parseAdminUserID(c *fiber.Ctx) (int, error) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return 0, repositories.ErrUserAccessNotFound
	}
	return id, nil
}

func PatchAdminUserAccess(c *fiber.Ctx) error {
	actor, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "usuario no autenticado"})
	}
	id, err := parseAdminUserID(c)
	if err != nil {
		return adminUserAccessError(c, err)
	}
	var request models.AdminUserAccessRequest
	if err := decodeStrictJSON(c.Body(), &request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "datos inválidos o campos no permitidos"})
	}
	result, err := services.SetAdminUserAccess(actor, id, request)
	if err != nil {
		return adminUserAccessError(c, err)
	}
	return c.JSON(result)
}

func GetAdminUserAccessHistory(c *fiber.Ctx) error {
	actor, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "usuario no autenticado"})
	}
	id, err := parseAdminUserID(c)
	if err != nil {
		return adminUserAccessError(c, err)
	}
	history, err := services.GetAdminUserAccessHistory(actor, id)
	if err != nil {
		return adminUserAccessError(c, err)
	}
	return c.JSON(history)
}
