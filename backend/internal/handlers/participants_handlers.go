package handlers

import (
	"errors"
	"strconv"
	"strings"

	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/services"

	"github.com/gofiber/fiber/v2"
)

// GetGroupReservationProgress obtiene el estado actual de una reserva grupal
// mediante su join code.
//
// El usuario debe estar autenticado, pero no necesita pertenecer previamente
// al grupo. Esto permite consultar el estado antes de decidir incorporarse.
//
// El join code nunca se utiliza directamente para consultar PostgreSQL:
// la capa repository lo normaliza y compara únicamente mediante su hash.
func GetGroupReservationProgress(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"error": "usuario no autenticado",
			},
		)
	}

	code := strings.TrimSpace(c.Params("code"))

	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "código de reserva obligatorio",
			},
		)
	}

	progress, err :=
		services.GetReservationProgressByJoinCode(
			code,
			user.ID,
		)

	if err != nil {
		return participantErrorResponse(c, err)
	}

	return c.JSON(progress)
}

// JoinGroupReservation incorpora al usuario autenticado a una reserva grupal.
//
// El cliente únicamente entrega el join code. El userID nunca proviene del
// body ni de parámetros controlados por el cliente: se obtiene exclusivamente
// desde la identidad local establecida por el middleware de autenticación.
//
// La operación es idempotente. Si el usuario ya pertenece al grupo como
// participante confirmado, la respuesta simplemente refleja el estado actual.
func JoinGroupReservation(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"error": "usuario no autenticado",
			},
		)
	}

	code := strings.TrimSpace(c.Params("code"))

	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "código de reserva obligatorio",
			},
		)
	}

	progress, err :=
		services.JoinGroupReservation(
			code,
			user.ID,
		)

	if err != nil {
		return participantErrorResponse(c, err)
	}

	return c.JSON(progress)
}

// LeaveGroupReservation retira al usuario autenticado del grupo.
//
// El owner no puede abandonar una reserva utilizando este endpoint porque
// sigue siendo responsable de ella. Para dejar de ser responsable debe
// cancelar la reserva completa mediante el flujo de cancelación.
//
// Si una reserva ya estaba CONFIRMED y queda bajo el mínimo después del
// retiro, conserva reservation.status = CONFIRMED y pasa a AT_RISK.
func LeaveGroupReservation(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"error": "usuario no autenticado",
			},
		)
	}

	code := strings.TrimSpace(c.Params("code"))

	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "código de reserva obligatorio",
			},
		)
	}

	progress, err :=
		services.LeaveGroupReservation(
			code,
			user.ID,
		)

	if err != nil {
		return participantErrorResponse(c, err)
	}

	return c.JSON(progress)
}

// GetGroupReservationParticipants entrega la lista de participantes de una
// reserva grupal.
//
// Por privacidad, la lista completa no queda disponible para cualquier
// usuario que conozca el join code.
//
// La capa de servicio restringe el acceso al:
//
//   - owner de la reserva;
//   - administrador.
//
// El ID de reserva se valida antes de delegar la autorización al servicio.
func GetGroupReservationParticipants(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"error": "usuario no autenticado",
			},
		)
	}

	reservationID, err :=
		strconv.Atoi(
			c.Params("id"),
		)

	if err != nil || reservationID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "reserva inválida",
			},
		)
	}

	participants, err :=
		services.GetReservationParticipantsForUser(
			reservationID,
			user,
		)

	if err != nil {
		return participantErrorResponse(c, err)
	}

	return c.JSON(participants)
}

// RotateGroupReservationJoinCode genera un nuevo código de invitación para
// una reserva grupal existente.
//
// Seguridad:
//   - requiere usuario autenticado;
//   - el servicio autoriza únicamente al owner o administrador;
//   - el código anterior queda invalidado al reemplazar su hash;
//   - el nuevo código en texto plano se devuelve una sola vez al cliente.
//
// El reservationID proviene de la URL, pero nunca se utiliza como criterio
// suficiente de autorización.
func RotateGroupReservationJoinCode(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"error": "usuario no autenticado",
			},
		)
	}

	reservationID, err :=
		strconv.Atoi(
			c.Params("id"),
		)

	if err != nil || reservationID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{
				"error": "reserva inválida",
			},
		)
	}

	joinCode, err :=
		services.RotateReservationJoinCodeForUser(
			reservationID,
			user,
		)

	if err != nil {
		return participantErrorResponse(c, err)
	}

	// El código en texto plano solamente aparece en esta respuesta.
	// Una lectura posterior de la reserva no permite recuperarlo.
	return c.JSON(
		fiber.Map{
			"joinCode": joinCode,
		},
	)
}

// participantErrorResponse centraliza la traducción de errores de dominio
// a códigos HTTP.
//
// Mantener este mapeo en un único punto evita que los endpoints de consulta,
// incorporación y retiro respondan de forma inconsistente ante la misma
// regla de negocio.
func participantErrorResponse(
	c *fiber.Ctx,
	err error,
) error {

	switch {
	case errors.Is(
		err,
		services.ErrInvalidJoinCode,
	):
		return c.Status(fiber.StatusNotFound).JSON(
			fiber.Map{
				"error": err.Error(),
			},
		)

	case errors.Is(
		err,
		services.ErrParticipantIneligible,
	):
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{
				"error": err.Error(),
			},
		)

	case errors.Is(
		err,
		services.ErrParticipationWindowClosed,
	):
		return c.Status(fiber.StatusConflict).JSON(
			fiber.Map{
				"error": err.Error(),
			},
		)

	case errors.Is(
		err,
		services.ErrParticipantScheduleOverlap,
	):
		return c.Status(fiber.StatusConflict).JSON(
			fiber.Map{
				"error": err.Error(),
			},
		)

	case errors.Is(
		err,
		services.ErrOwnerCannotWithdraw,
	):
		return c.Status(fiber.StatusConflict).JSON(
			fiber.Map{
				"error": err.Error(),
			},
		)

	case errors.Is(
		err,
		services.ErrGroupCapacity,
	):
		return c.Status(fiber.StatusConflict).JSON(
			fiber.Map{
				"error": err.Error(),
			},
		)

	case errors.Is(
		err,
		services.ErrReservationNotJoinable,
	):
		return c.Status(fiber.StatusConflict).JSON(
			fiber.Map{
				"error": err.Error(),
			},
		)

	case errors.Is(
		err,
		services.ErrReservationNotFound,
	):
		return c.Status(fiber.StatusNotFound).JSON(
			fiber.Map{
				"error": err.Error(),
			},
		)

	case errors.Is(
		err,
		services.ErrReservationForbidden,
	):
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{
				"error": err.Error(),
			},
		)

	case errors.Is(
		err,
		services.ErrInvalidGroupConfig,
	):
		// Una configuración grupal inválida representa un problema interno
		// de política o datos y no un error corregible por el usuario.
		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(
			fiber.Map{
				"error": "configuración grupal inválida",
			},
		)

	default:
		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(
			fiber.Map{
				"error": "no se pudo procesar la participación",
			},
		)
	}
}
