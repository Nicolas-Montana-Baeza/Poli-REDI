package handlers

import (
	"errors"
	"strconv"

	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/repositories"
	"poli-redi-api/internal/services"

	"github.com/gofiber/fiber/v2"
)

// ============================================================================
// TALLERES INSTITUCIONALES
// ============================================================================
//
// WORKSHOP no posee un calendario independiente.
//
// Es una institutional_activity con:
//
//   activity_type       = WORKSHOP
//   requires_enrollment = true
//   capacity            = N
//
// De esta forma reutiliza programación, conflictos y disponibilidad del
// módulo institucional.

// GET /api/workshops
func GetInstitutionalWorkshops(
	c *fiber.Ctx,
) error {
	user, ok :=
		middleware.GetLocalUser(c)

	if !ok {
		return c.
			Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{
				"error": "usuario no autenticado",
			})
	}

	workshops, err :=
		services.GetInstitutionalWorkshops(
			user,
		)

	if err != nil {
		status, message :=
			institutionalWorkshopErrorResponse(
				err,
			)

		return c.
			Status(status).
			JSON(fiber.Map{
				"error": message,
			})
	}

	return c.JSON(workshops)
}

// GET /api/workshops/:id
func GetInstitutionalWorkshop(
	c *fiber.Ctx,
) error {
	user, ok :=
		middleware.GetLocalUser(c)

	if !ok {
		return c.
			Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{
				"error": "usuario no autenticado",
			})
	}

	workshopID, err :=
		strconv.Atoi(
			c.Params("id"),
		)

	if err != nil ||
		workshopID <= 0 {

		return c.
			Status(fiber.StatusBadRequest).
			JSON(fiber.Map{
				"error": "id de taller inválido",
			})
	}

	workshop, err :=
		services.GetInstitutionalWorkshop(
			workshopID,
			user,
		)

	if err != nil {
		status, message :=
			institutionalWorkshopErrorResponse(
				err,
			)

		return c.
			Status(status).
			JSON(fiber.Map{
				"error": message,
			})
	}

	return c.JSON(workshop)
}

// POST /api/workshops/:id/enroll
func EnrollInInstitutionalWorkshop(
	c *fiber.Ctx,
) error {
	user, ok :=
		middleware.GetLocalUser(c)

	if !ok {
		return c.
			Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{
				"error": "usuario no autenticado",
			})
	}

	workshopID, err :=
		strconv.Atoi(
			c.Params("id"),
		)

	if err != nil ||
		workshopID <= 0 {

		return c.
			Status(fiber.StatusBadRequest).
			JSON(fiber.Map{
				"error": "id de taller inválido",
			})
	}

	workshop, err :=
		services.EnrollInInstitutionalWorkshop(
			workshopID,
			user,
		)

	if err != nil {
		status, message :=
			institutionalWorkshopErrorResponse(
				err,
			)

		return c.
			Status(status).
			JSON(fiber.Map{
				"error": message,
			})
	}

	return c.
		Status(fiber.StatusOK).
		JSON(workshop)
}

// DELETE /api/workshops/:id/enroll
func LeaveInstitutionalWorkshop(
	c *fiber.Ctx,
) error {
	user, ok :=
		middleware.GetLocalUser(c)

	if !ok {
		return c.
			Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{
				"error": "usuario no autenticado",
			})
	}

	workshopID, err :=
		strconv.Atoi(
			c.Params("id"),
		)

	if err != nil ||
		workshopID <= 0 {

		return c.
			Status(fiber.StatusBadRequest).
			JSON(fiber.Map{
				"error": "id de taller inválido",
			})
	}

	workshop, err :=
		services.LeaveInstitutionalWorkshop(
			workshopID,
			user,
		)

	if err != nil {
		status, message :=
			institutionalWorkshopErrorResponse(
				err,
			)

		return c.
			Status(status).
			JSON(fiber.Map{
				"error": message,
			})
	}

	return c.JSON(workshop)
}

// ============================================================================
// ERRORES HTTP
// ============================================================================

func institutionalWorkshopErrorResponse(
	err error,
) (int, string) {
	publicMessage := sec005KnownSentinelMessage(err,
		services.ErrInstitutionalWorkshopUnauthorized,
		services.ErrInstitutionalWorkshopInvalidID,
		repositories.ErrInstitutionalWorkshopNotFound,
		repositories.ErrInstitutionalWorkshopUnavailable,
		repositories.ErrInstitutionalWorkshopAlreadyEnrolled,
		repositories.ErrInstitutionalWorkshopNotEnrolled,
		repositories.ErrInstitutionalWorkshopFull,
	)

	switch {

	case errors.Is(
		err,
		services.
			ErrInstitutionalWorkshopUnauthorized,
	):

		return fiber.StatusUnauthorized,
			publicMessage

	case errors.Is(
		err,
		services.
			ErrInstitutionalWorkshopInvalidID,
	):

		return fiber.StatusBadRequest,
			publicMessage

	case errors.Is(
		err,
		repositories.
			ErrInstitutionalWorkshopNotFound,
	):

		return fiber.StatusNotFound,
			publicMessage

	case errors.Is(
		err,
		repositories.
			ErrInstitutionalWorkshopUnavailable,
	),
		errors.Is(
			err,
			repositories.
				ErrInstitutionalWorkshopAlreadyEnrolled,
		),
		errors.Is(
			err,
			repositories.
				ErrInstitutionalWorkshopNotEnrolled,
		),
		errors.Is(
			err,
			repositories.
				ErrInstitutionalWorkshopFull,
		):

		return fiber.StatusConflict,
			publicMessage

	default:

		return fiber.
				StatusInternalServerError,
			"no se pudo procesar el taller"
	}
}
