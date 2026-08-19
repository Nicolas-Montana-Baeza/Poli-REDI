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
// UNIDADES INSTITUCIONALES
// ============================================================================
//
// Los handlers mantienen una responsabilidad deliberadamente pequeña:
//
//   - obtener la identidad autenticada;
//   - validar parámetros HTTP;
//   - decodificar JSON estricto;
//   - delegar autorización y reglas de negocio al service;
//   - traducir errores de dominio a códigos HTTP.
//
// Las reglas de quién puede administrar una unidad NO se duplican aquí.
// services/institutional_units_service.go sigue siendo la fuente autoritativa.

// GetInstitutionalUnits devuelve las unidades que el actor autenticado puede
// gestionar.
//
// Administradores globales ven todas las unidades activas.
// Managers ven únicamente las unidades que administran.
func GetInstitutionalUnits(
	c *fiber.Ctx,
) error {
	user, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"error": "usuario no autenticado",
			},
		)
	}

	units, err :=
		services.GetInstitutionalUnitsForUser(user)

	if err != nil {
		return institutionalUnitHTTPError(c, err)
	}

	return c.JSON(units)
}

// CreateInstitutionalUnit crea una estructura institucional.
//
// La autorización administrativa se valida también en el service para evitar
// depender exclusivamente del middleware de rutas.
func CreateInstitutionalUnit(
	c *fiber.Ctx,
) error {
	user, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{
				"error": "usuario no autenticado",
			},
		)
	}

	var request models.CreateInstitutionalUnitRequest

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

	unit, err :=
		services.CreateInstitutionalUnit(
			user,
			request,
		)

	if err != nil {
		return institutionalUnitHTTPError(c, err)
	}

	return c.Status(
		fiber.StatusCreated,
	).JSON(unit)
}

// GetInstitutionalUnitMemberships devuelve los miembros de una unidad.
//
// Un administrador global o un MANAGER activo de la unidad puede consultar
// esta información.
func GetInstitutionalUnitMemberships(
	c *fiber.Ctx,
) error {
	user, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(
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

	memberships, err :=
		services.GetInstitutionalUnitMemberships(
			user,
			unitID,
		)

	if err != nil {
		return institutionalUnitHTTPError(c, err)
	}

	return c.JSON(memberships)
}

// AddInstitutionalUnitMembership agrega o reactiva una relación institucional.
//
// En MVP2 únicamente el administrador global puede asignar MEMBER/MANAGER.
// El service vuelve a verificar esta regla como defensa en profundidad.
func AddInstitutionalUnitMembership(
	c *fiber.Ctx,
) error {
	user, ok := middleware.GetLocalUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(
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

	var request models.AddInstitutionalUnitMembershipRequest

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

	membership, err :=
		services.AddInstitutionalUnitMembership(
			user,
			unitID,
			request,
		)

	if err != nil {
		return institutionalUnitHTTPError(c, err)
	}

	return c.Status(
		fiber.StatusCreated,
	).JSON(membership)
}

// institutionalUnitHTTPError mantiene centralizado el contrato HTTP del módulo.
//
// De esta manera los handlers no dependen de mensajes de PostgreSQL y el
// frontend recibe estados consistentes para autorización, validación,
// inexistencia y conflictos de estado.
func institutionalUnitHTTPError(
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
			services.ErrInstitutionalMembershipTargetNotFound,
		):
		return c.Status(
			fiber.StatusNotFound,
		).JSON(
			fiber.Map{"error": err.Error()},
		)

	case errors.Is(
		err,
		services.ErrInstitutionalInvalidUnit,
	),
		errors.Is(
			err,
			services.ErrInstitutionalInvalidMembership,
		):
		return c.Status(
			fiber.StatusBadRequest,
		).JSON(
			fiber.Map{"error": err.Error()},
		)

	case errors.Is(
		err,
		services.ErrInstitutionalUnitDuplicate,
	),
		errors.Is(
			err,
			services.ErrInstitutionalUnitInactive,
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
