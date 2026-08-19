package handlers

import (
	"errors"
	"strconv"

	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/services"

	"github.com/gofiber/fiber/v2"
)

// ============================================================================
// ACTIVIDADES INSTITUCIONALES
// ============================================================================
//
// Esta capa traduce HTTP al dominio institucional.
//
// Las reglas de autorización, horarios, recursos, recurrencia y conflictos
// permanecen en services. El handler no replica reglas de negocio.

// GetInstitutionalActivitiesForUnit devuelve la programación institucional
// visible para el administrador global o MANAGER de la unidad.
//
// El ID de la unidad por sí solo no entrega acceso: el service vuelve a
// validar la relación institucional del usuario autenticado.
func GetInstitutionalActivitiesForUnit(
	c *fiber.Ctx,
) error {
	user, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.Status(
			fiber.StatusUnauthorized,
		).JSON(
			fiber.Map{
				"error": "usuario no autenticado",
			},
		)
	}

	unitID, err := strconv.Atoi(
		c.Params("id"),
	)

	if err != nil || unitID <= 0 {
		return c.Status(
			fiber.StatusBadRequest,
		).JSON(
			fiber.Map{
				"error": "unidad institucional inválida",
			},
		)
	}

	activities, err :=
		services.GetInstitutionalActivitiesForUnit(
			user,
			unitID,
		)

	if err != nil {
		return institutionalActivityHTTPError(
			c,
			err,
		)
	}

	return c.JSON(activities)
}

// CreateInstitutionalActivity crea una actividad junto con sus horarios.
//
// Tanto el administrador global como un MANAGER de la unidad pueden utilizar
// esta operación. No usamos RequireAdmin en la ruta porque la autorización
// fina depende de la unidad concreta y vive en el service.
func CreateInstitutionalActivity(
	c *fiber.Ctx,
) error {
	user, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.Status(
			fiber.StatusUnauthorized,
		).JSON(
			fiber.Map{
				"error": "usuario no autenticado",
			},
		)
	}

	var request models.CreateInstitutionalActivityRequest

	if err := decodeStrictJSON(
		c.Body(),
		&request,
	); err != nil {
		return c.Status(
			fiber.StatusBadRequest,
		).JSON(
			fiber.Map{
				"error": "datos inválidos o campos no permitidos",
			},
		)
	}

	activity, err :=
		services.CreateInstitutionalActivity(
			user,
			request,
		)

	if err != nil {
		return institutionalActivityHTTPError(
			c,
			err,
		)
	}

	return c.Status(
		fiber.StatusCreated,
	).JSON(activity)
}

// institutionalActivityHTTPError centraliza el contrato HTTP del módulo.
//
// Evitamos filtrar errores técnicos de PostgreSQL al cliente y mantenemos una
// semántica estable para frontend:
//
//	400 -> request/horario inválido
//	403 -> actor sin permisos sobre la unidad
//	404 -> unidad/recurso inexistente
//	409 -> recurso/unidad inactiva o bloqueo duro de disponibilidad
func institutionalActivityHTTPError(
	c *fiber.Ctx,
	err error,
) error {
	switch {

	case errors.Is(
		err,
		services.ErrInstitutionalUnauthorized,
	):
		return c.Status(
			fiber.StatusForbidden,
		).JSON(
			fiber.Map{"error": err.Error()},
		)

	case errors.Is(
		err,
		services.ErrInstitutionalUnitNotFound,
	),
		errors.Is(
			err,
			services.ErrInstitutionalResourceNotFound,
		):
		return c.Status(
			fiber.StatusNotFound,
		).JSON(
			fiber.Map{"error": err.Error()},
		)

	case errors.Is(
		err,
		services.ErrInstitutionalActivityInvalid,
	),
		errors.Is(
			err,
			services.ErrInstitutionalScheduleInvalid,
		):
		return c.Status(
			fiber.StatusBadRequest,
		).JSON(
			fiber.Map{"error": err.Error()},
		)

	case errors.Is(
		err,
		services.ErrInstitutionalUnitInactive,
	),
		errors.Is(
			err,
			services.ErrInstitutionalResourceInactive,
		),
		errors.Is(
			err,
			services.ErrInstitutionalActivityBlocked,
		):
		return c.Status(
			fiber.StatusConflict,
		).JSON(
			fiber.Map{"error": err.Error()},
		)

	default:
		return c.Status(
			fiber.StatusInternalServerError,
		).JSON(
			fiber.Map{
				"error": "no se pudo completar la operación institucional",
			},
		)
	}
}
