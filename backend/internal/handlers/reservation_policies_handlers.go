package handlers

import (
	"database/sql"
	"errors"

	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/services"

	"github.com/gofiber/fiber/v2"
)

func GetCurrentReservationPolicy(c *fiber.Ctx) error {
	policy, err := services.GetCurrentReservationPolicy()
	if errors.Is(err, sql.ErrNoRows) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no existe una politica de reservas vigente"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no se pudo cargar la politica de reservas"})
	}
	return c.JSON(policy.Public())
}

func GetReservationPolicyHistory(c *fiber.Ctx) error {
	policies, err := services.GetReservationPolicyHistory()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no se pudo cargar el historial de politicas"})
	}
	return c.JSON(policies)
}

func PublishReservationPolicy(c *fiber.Ctx) error {
	user, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "usuario no autenticado"})
	}
	var request models.PublishReservationPolicyRequest
	if err := decodeStrictJSON(c.Body(), &request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "datos invalidos o campos no permitidos"})
	}
	policy, replayed, err := services.PublishReservationPolicy(request, user.ID, c.Get("Idempotency-Key"))
	if err != nil {
		status, message := reservationPolicyErrorResponse(err)
		return c.Status(status).JSON(fiber.Map{"error": message})
	}
	if replayed {
		return c.Status(fiber.StatusOK).JSON(policy)
	}
	return c.Status(fiber.StatusCreated).JSON(policy)
}

func reservationPolicyErrorResponse(err error) (int, string) {
	var validation services.ReservationPolicyValidationError
	if errors.As(err, &validation) {
		return fiber.StatusBadRequest, err.Error()
	}
	if errors.Is(err, services.ErrReservationPolicyConflict) {
		return fiber.StatusConflict, err.Error()
	}
	return fiber.StatusInternalServerError, "no se pudo publicar la politica de reservas"
}
