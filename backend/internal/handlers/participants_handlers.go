package handlers

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/models"
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

func UpdateReservationTarget(c *fiber.Ctx) error {
	u, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.SendStatus(401)
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "reserva invalida"})
	}
	var request models.UpdateTargetParticipantsRequest
	if err := decodeStrictJSON(c.Body(), &request); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Datos invalidos o campos no permitidos"})
	}
	progress, err := repositories.UpdateReservationTarget(id, u.ID, request.TargetParticipants, businessclock.Now())
	if err != nil {
		status := 400
		if errors.Is(err, repositories.ErrTargetForbidden) {
			status = 403
		}
		if errors.Is(err, repositories.ErrInvalidJoinCode) {
			status = 404
		}
		if errors.Is(err, repositories.ErrTargetDeadline) {
			status = 410
		}
		if errors.Is(err, repositories.ErrTargetBelowConfirmed) {
			status = 409
		}
		return c.Status(status).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(progress)
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
		return c.Status(participationHTTPStatus(e)).JSON(fiber.Map{"error": e.Error()})
	}
	return c.JSON(p)
}

func participationHTTPStatus(err error) int {
	switch {
	case errors.Is(err, repositories.ErrInvalidJoinCode):
		return 404
	case errors.Is(err, repositories.ErrParticipantIneligible):
		return 403
	case errors.Is(err, repositories.ErrGroupCapacity), errors.Is(err, repositories.ErrOwnerCannotWithdraw):
		return 409
	case errors.Is(err, repositories.ErrParticipationDeadline):
		return 410
	default:
		return 400
	}
}

func GetOwnerJoinCode(c *fiber.Ctx) error {
	return ownerJoinCode(c, false)
}

func RotateOwnerJoinCode(c *fiber.Ctx) error {
	return ownerJoinCode(c, true)
}

func ownerJoinCode(c *fiber.Ctx, rotate bool) error {
	u, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.SendStatus(401)
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": repositories.ErrInvalidJoinCode.Error()})
	}
	var code string
	if rotate {
		code, err = repositories.RotateOwnerJoinCode(id, u.ID)
	} else {
		code, err = repositories.GetOwnerJoinCode(id, u.ID)
	}
	if errors.Is(err, repositories.ErrInvalidJoinCode) {
		return c.Status(404).JSON(fiber.Map{"error": repositories.ErrInvalidJoinCode.Error()})
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "no se pudo recuperar el codigo"})
	}
	return c.JSON(models.JoinCodeResponse{JoinCode: code})
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
