package handlers

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/repositories"
	"strconv"
	"strings"
)

func participantCode(c *fiber.Ctx) string { return strings.TrimSpace(c.Params("code")) }
func GetReservationProgress(c *fiber.Ctx) error {
	u, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.SendStatus(401)
	}
	p, e := repositories.GetReservationProgress(participantCode(c), u.ID)
	if errors.Is(e, repositories.ErrInvalidJoinCode) {
		return c.Status(404).JSON(fiber.Map{"error": e.Error()})
	}
	if e != nil {
		return c.Status(500).JSON(fiber.Map{"error": "no se pudo consultar el progreso"})
	}
	return c.JSON(p)
}
func ConfirmParticipation(c *fiber.Ctx) error  { return changeParticipation(c, true) }
func WithdrawParticipation(c *fiber.Ctx) error { return changeParticipation(c, false) }
func changeParticipation(c *fiber.Ctx, confirm bool) error {
	u, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.SendStatus(401)
	}
	p, e := repositories.ChangeParticipation(participantCode(c), u.ID, confirm)
	if e != nil {
		status := 400
		if errors.Is(e, repositories.ErrInvalidJoinCode) {
			status = 404
		}
		if errors.Is(e, repositories.ErrParticipantIneligible) {
			status = 403
		}
		return c.Status(status).JSON(fiber.Map{"error": e.Error()})
	}
	return c.JSON(p)
}
func GetReservationParticipants(c *fiber.Ctx) error {
	id, e := strconv.Atoi(c.Params("id"))
	if e != nil {
		return c.Status(400).JSON(fiber.Map{"error": "reserva invalida"})
	}
	p, e := repositories.GetReservationParticipants(id)
	if e != nil {
		return c.Status(500).JSON(fiber.Map{"error": "no se pudieron cargar los participantes"})
	}
	return c.JSON(p)
}
