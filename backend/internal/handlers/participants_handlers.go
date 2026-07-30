package handlers

import (
	"errors"
	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func participantCode(c *fiber.Ctx) string { return strings.TrimSpace(c.Params("code")) }
func GetReservationProgress(c *fiber.Ctx) error {
	u, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.SendStatus(401)
	}
	p, e := repositories.GetReservationProgress(participantCode(c), u.ID)
	if errors.Is(e, repositories.ErrInvalidJoinCode) {
		return c.Status(404).JSON(fiber.Map{"error": "El código no existe o ya no está disponible."})
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
		status, message := targetUpdateErrorResponse(err)
		return c.Status(status).JSON(fiber.Map{"error": message})
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
		status, message := participationErrorResponse(e)
		return c.Status(status).JSON(fiber.Map{"error": message})
	}
	return c.JSON(p)
}

func participationHTTPStatus(err error) int {
	status, _ := participationErrorResponse(err)
	return status
}

func participationErrorResponse(err error) (int, string) {
	switch {
	case errors.Is(err, repositories.ErrInvalidJoinCode):
		return fiber.StatusNotFound, repositories.ErrInvalidJoinCode.Error()
	case errors.Is(err, repositories.ErrParticipantIneligible):
		return fiber.StatusForbidden, repositories.ErrParticipantIneligible.Error()
	case errors.Is(err, repositories.ErrGroupCapacity), errors.Is(err, repositories.ErrOwnerCannotWithdraw), errors.Is(err, repositories.ErrParticipantConflict):
		return fiber.StatusConflict, err.Error()
	case errors.Is(err, repositories.ErrParticipationDeadline):
		return fiber.StatusGone, repositories.ErrParticipationDeadline.Error()
	default:
		return fiber.StatusInternalServerError, "no se pudo cambiar la participacion"
	}
}

func targetUpdateErrorResponse(err error) (int, string) {
	switch {
	case errors.Is(err, repositories.ErrTargetForbidden):
		return fiber.StatusForbidden, repositories.ErrTargetForbidden.Error()
	case errors.Is(err, repositories.ErrInvalidJoinCode):
		return fiber.StatusNotFound, repositories.ErrInvalidJoinCode.Error()
	case errors.Is(err, repositories.ErrTargetDeadline):
		return fiber.StatusGone, repositories.ErrTargetDeadline.Error()
	case errors.Is(err, repositories.ErrTargetBelowConfirmed):
		return fiber.StatusConflict, repositories.ErrTargetBelowConfirmed.Error()
	case errors.Is(err, repositories.ErrInvalidTargetParticipants):
		return fiber.StatusBadRequest, repositories.ErrInvalidTargetParticipants.Error()
	default:
		return fiber.StatusInternalServerError, "no se pudo actualizar el objetivo de participantes"
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
