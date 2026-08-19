package services

import (
	"errors"
	"strings"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
)

// ============================================================================
// CONSULTAS ADMINISTRATIVAS DE CONFLICTOS
// ============================================================================

var ErrSchedulingConflictInvalidFilter = errors.New(
	"filtro de conflictos de programación inválido",
)

// GetSchedulingConflictsForAdmin devuelve conflictos completos para el panel
// administrativo.
//
// Filtros soportados:
//
//	PENDING
//	RESOLVED
//	ALL
//
// Si el cliente no envía filtro utilizamos PENDING porque son los casos que
// requieren intervención administrativa.
func GetSchedulingConflictsForAdmin(
	requestedBy models.LocalAuthUser,
	status string,
) ([]models.SchedulingConflict, error) {
	if requestedBy.ID <= 0 ||
		!requestedBy.IsAdmin ||
		requestedBy.IsBlocked {

		return nil,
			ErrSchedulingConflictUnauthorized
	}

	status =
		strings.ToUpper(
			strings.TrimSpace(
				status,
			),
		)

	if status == "" {
		status =
			models.
				SchedulingConflictStatusPending
	}

	switch status {

	case models.SchedulingConflictStatusPending,
		models.SchedulingConflictStatusResolved:

		return repositories.
			GetSchedulingConflicts(
				status,
			)

	case "ALL":

		return repositories.
			GetSchedulingConflicts("")

	default:

		return nil,
			ErrSchedulingConflictInvalidFilter
	}
}

// GetSchedulingConflictForAdmin carga el detalle N-elementos de un conflicto.
//
// La autorización también se valida en servicio. Así la seguridad no depende
// exclusivamente de que una ruta HTTP concreta recuerde agregar RequireAdmin.
func GetSchedulingConflictForAdmin(
	requestedBy models.LocalAuthUser,
	conflictID int,
) (models.SchedulingConflict, error) {
	if requestedBy.ID <= 0 ||
		!requestedBy.IsAdmin ||
		requestedBy.IsBlocked {

		return models.SchedulingConflict{},
			ErrSchedulingConflictUnauthorized
	}

	if conflictID <= 0 {
		return models.SchedulingConflict{},
			repositories.
				ErrSchedulingConflictNotFound
	}

	return repositories.
		GetSchedulingConflictByID(
			conflictID,
		)
}
