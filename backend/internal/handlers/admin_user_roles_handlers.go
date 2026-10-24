package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
	"poli-redi-api/internal/services"
)

func adminUserRoleError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repositories.ErrUserRoleUnauthorized), errors.Is(err, repositories.ErrUserRoleSelf):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, repositories.ErrUserRoleNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, services.ErrUserRoleInvalidReason), errors.Is(err, services.ErrUserRoleInvalidRequest):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, repositories.ErrUserRoleUnchanged), errors.Is(err, repositories.ErrUserRoleLastAdmin),
		errors.Is(err, repositories.ErrUserRoleBlockedTarget), errors.Is(err, repositories.ErrUserRoleConflict):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no se pudo actualizar el rol"})
	}
}

func PatchAdminUserRole(c *fiber.Ctx) error {
	actor, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "usuario no autenticado"})
	}
	id, err := parseAdminUserID(c)
	if err != nil {
		return adminUserRoleError(c, repositories.ErrUserRoleNotFound)
	}
	var req models.AdminUserRoleRequest
	if err := decodeStrictJSON(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "datos invalidos o campos no permitidos"})
	}
	result, err := services.SetAdminUserRole(actor, id, req)
	if err != nil {
		return adminUserRoleError(c, err)
	}
	return c.JSON(result)
}

func GetAdminUserRoleHistory(c *fiber.Ctx) error {
	actor, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "usuario no autenticado"})
	}
	id, err := parseAdminUserID(c)
	if err != nil {
		return adminUserRoleError(c, repositories.ErrUserRoleNotFound)
	}
	rows, err := services.GetAdminUserRoleHistory(actor, id)
	if err != nil {
		return adminUserRoleError(c, err)
	}
	return c.JSON(rows)
}
