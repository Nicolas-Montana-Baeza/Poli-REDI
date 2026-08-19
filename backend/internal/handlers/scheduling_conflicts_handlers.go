package handlers

import (
	"errors"
	"strconv"

	"poli-redi-api/internal/middleware"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
	"poli-redi-api/internal/services"

	"github.com/gofiber/fiber/v2"
)

// ============================================================================
// LISTAR CONFLICTOS
// ============================================================================
//
// GET /api/admin/scheduling-conflicts
//
// Query:
//
//   ?status=PENDING
//   ?status=RESOLVED
//   ?status=ALL
//
// Por defecto se muestran los conflictos pendientes.

func GetSchedulingConflicts(
	c *fiber.Ctx,
) error {
	user, ok :=
		middleware.GetLocalUser(c)

	if !ok {
		return c.
			Status(
				fiber.StatusUnauthorized,
			).
			JSON(
				fiber.Map{
					"error": "usuario no autenticado",
				},
			)
	}

	conflicts, err :=
		services.
			GetSchedulingConflictsForAdmin(
				user,
				c.Query(
					"status",
					models.
						SchedulingConflictStatusPending,
				),
			)

	if err != nil {
		status, message :=
			schedulingConflictErrorResponse(
				err,
			)

		return c.
			Status(status).
			JSON(
				fiber.Map{
					"error": message,
				},
			)
	}

	return c.JSON(conflicts)
}

// ============================================================================
// DETALLE
// ============================================================================
//
// GET /api/admin/scheduling-conflicts/:id

func GetSchedulingConflict(
	c *fiber.Ctx,
) error {
	user, ok :=
		middleware.GetLocalUser(c)

	if !ok {
		return c.
			Status(
				fiber.StatusUnauthorized,
			).
			JSON(
				fiber.Map{
					"error": "usuario no autenticado",
				},
			)
	}

	conflictID, err :=
		strconv.Atoi(
			c.Params("id"),
		)

	if err != nil ||
		conflictID <= 0 {

		return c.
			Status(
				fiber.StatusBadRequest,
			).
			JSON(
				fiber.Map{
					"error": "id de conflicto inválido",
				},
			)
	}

	conflict, err :=
		services.
			GetSchedulingConflictForAdmin(
				user,
				conflictID,
			)

	if err != nil {
		status, message :=
			schedulingConflictErrorResponse(
				err,
			)

		return c.
			Status(status).
			JSON(
				fiber.Map{
					"error": message,
				},
			)
	}

	return c.JSON(conflict)
}

// ============================================================================
// RESOLVER ELEMENTO
// ============================================================================
//
// PATCH /api/admin/scheduling-conflicts/:id/items/:itemId
//
// El body representa UNA decisión administrativa:
//
//   KEEP
//   ALLOW
//   CANCEL
//   RESCHEDULE
//
// Las resoluciones parciales conservan el conflicto PENDING.
// El servicio cierra automáticamente el grupo cuando no quedan elementos
// pendientes.

func PatchSchedulingConflictItem(
	c *fiber.Ctx,
) error {
	user, ok :=
		middleware.GetLocalUser(c)

	if !ok {
		return c.
			Status(
				fiber.StatusUnauthorized,
			).
			JSON(
				fiber.Map{
					"error": "usuario no autenticado",
				},
			)
	}

	conflictID, err :=
		strconv.Atoi(
			c.Params("id"),
		)

	if err != nil ||
		conflictID <= 0 {

		return c.
			Status(
				fiber.StatusBadRequest,
			).
			JSON(
				fiber.Map{
					"error": "id de conflicto inválido",
				},
			)
	}

	itemID, err :=
		strconv.Atoi(
			c.Params("itemId"),
		)

	if err != nil ||
		itemID <= 0 {

		return c.
			Status(
				fiber.StatusBadRequest,
			).
			JSON(
				fiber.Map{
					"error": "id de elemento inválido",
				},
			)
	}

	var request models.
		ResolveSchedulingConflictItemRequest

	if err := decodeStrictJSON(
		c.Body(),
		&request,
	); err != nil {

		return c.
			Status(
				fiber.StatusBadRequest,
			).
			JSON(
				fiber.Map{
					"error": "datos inválidos o campos no permitidos",
				},
			)
	}

	conflict, err :=
		services.
			ResolveSchedulingConflictItem(
				conflictID,
				itemID,
				user,
				request,
			)

	if err != nil {
		status, message :=
			schedulingConflictErrorResponse(
				err,
			)

		return c.
			Status(status).
			JSON(
				fiber.Map{
					"error": message,
				},
			)
	}

	return c.JSON(conflict)
}

// ============================================================================
// MAPEO DE ERRORES HTTP
// ============================================================================

func schedulingConflictErrorResponse(
	err error,
) (int, string) {
	switch {

	case errors.Is(
		err,
		services.
			ErrSchedulingConflictUnauthorized,
	):

		return fiber.StatusForbidden,
			err.Error()

	case errors.Is(
		err,
		services.
			ErrSchedulingConflictInvalidFilter,
	),
		errors.Is(
			err,
			services.
				ErrSchedulingConflictInvalidResolution,
		),
		errors.Is(
			err,
			services.
				ErrSchedulingConflictResolutionNoteRequired,
		),
		errors.Is(
			err,
			services.
				ErrSchedulingConflictInvalidReschedule,
		),
		errors.Is(
			err,
			repositories.
				ErrSchedulingRescheduleReservation,
		):

		return fiber.StatusBadRequest,
			err.Error()

	case errors.Is(
		err,
		repositories.
			ErrSchedulingConflictNotFound,
	),
		errors.Is(
			err,
			repositories.
				ErrSchedulingConflictItemNotFound,
		):

		return fiber.StatusNotFound,
			err.Error()

	case errors.Is(
		err,
		repositories.
			ErrSchedulingConflictResolved,
	),
		errors.Is(
			err,
			repositories.
				ErrSchedulingConflictItemResolved,
		),
		errors.Is(
			err,
			repositories.
				ErrSchedulingResolutionInvalidPlan,
		),
		errors.Is(
			err,
			repositories.
				ErrSchedulingOccurrenceAlreadyAdjusted,
		),
		errors.Is(
			err,
			repositories.
				ErrSchedulingRescheduleBlocked,
		):

		return fiber.StatusConflict,
			err.Error()

	default:

		return fiber.
				StatusInternalServerError,
			"no se pudo procesar el conflicto de programación"
	}
}
