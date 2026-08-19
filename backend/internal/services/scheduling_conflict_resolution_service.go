package services

import (
	"errors"
	"strings"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
)

// ============================================================================
// ERRORES DE DOMINIO
// ============================================================================

var (
	ErrSchedulingConflictUnauthorized = errors.New(
		"solo un administrador puede resolver conflictos de programación",
	)

	ErrSchedulingConflictInvalidResolution = errors.New(
		"resolución de conflicto inválida",
	)

	ErrSchedulingConflictResolutionNoteRequired = errors.New(
		"la resolución administrativa debe incluir una justificación",
	)

	ErrSchedulingConflictInvalidReschedule = errors.New(
		"la nueva fecha u horario de reprogramación es inválido",
	)
)

// ============================================================================
// RESOLVER UN ELEMENTO
// ============================================================================

func ResolveSchedulingConflictItem(
	conflictID int,
	itemID int,
	requestedBy models.LocalAuthUser,
	request models.ResolveSchedulingConflictItemRequest,
) (models.SchedulingConflict, error) {
	if requestedBy.ID <= 0 ||
		!requestedBy.IsAdmin ||
		requestedBy.IsBlocked {

		return models.SchedulingConflict{},
			ErrSchedulingConflictUnauthorized
	}

	if conflictID <= 0 ||
		itemID <= 0 {

		return models.SchedulingConflict{},
			ErrSchedulingConflictInvalidResolution
	}

	request.Resolution =
		strings.ToUpper(
			strings.TrimSpace(
				request.Resolution,
			),
		)

	request.ResolutionNote =
		strings.TrimSpace(
			request.ResolutionNote,
		)

	if request.ResolutionNote == "" {
		return models.SchedulingConflict{},
			ErrSchedulingConflictResolutionNoteRequired
	}

	switch request.Resolution {

	case models.SchedulingItemResolutionKeep,
		models.SchedulingItemResolutionAllow,
		models.SchedulingItemResolutionCancel:

		// Estos tipos de resolución no deben transportar accidentalmente
		// parámetros de reprogramación.
		if request.NewDate != nil ||
			request.NewStartTime != nil ||
			request.NewEndTime != nil {

			return models.SchedulingConflict{},
				ErrSchedulingConflictInvalidResolution
		}

	case models.SchedulingItemResolutionReschedule:

		if err := validateSchedulingRescheduleRequest(
			&request,
		); err != nil {

			return models.SchedulingConflict{},
				err
		}

	default:

		return models.SchedulingConflict{},
			ErrSchedulingConflictInvalidResolution
	}

	return repositories.ResolveSchedulingConflictItem(
		conflictID,
		itemID,
		requestedBy.ID,
		request,
	)
}

// ============================================================================
// VALIDACIÓN DE RESCHEDULE
// ============================================================================

func validateSchedulingRescheduleRequest(
	request *models.ResolveSchedulingConflictItemRequest,
) error {
	if request.NewDate == nil ||
		request.NewStartTime == nil ||
		request.NewEndTime == nil {

		return ErrSchedulingConflictInvalidReschedule
	}

	dateValue :=
		strings.TrimSpace(
			*request.NewDate,
		)

	startValue :=
		strings.TrimSpace(
			*request.NewStartTime,
		)

	endValue :=
		strings.TrimSpace(
			*request.NewEndTime,
		)

	date, err :=
		time.Parse(
			"2006-01-02",
			dateValue,
		)

	if err != nil {
		return ErrSchedulingConflictInvalidReschedule
	}

	startClock, err :=
		time.Parse(
			"15:04",
			startValue,
		)

	if err != nil {
		return ErrSchedulingConflictInvalidReschedule
	}

	endClock, err :=
		time.Parse(
			"15:04",
			endValue,
		)

	if err != nil {
		return ErrSchedulingConflictInvalidReschedule
	}

	start := time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		startClock.Hour(),
		startClock.Minute(),
		0,
		0,
		businessclock.Location(),
	)

	end := time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		endClock.Hour(),
		endClock.Minute(),
		0,
		0,
		businessclock.Location(),
	)

	// MVP2 no permite programación institucional que cruce medianoche.
	if !end.After(start) {
		return ErrSchedulingConflictInvalidReschedule
	}

	// No reprogramamos hacia una ocurrencia que ya terminó.
	if !end.After(
		businessclock.Now(),
	) {
		return ErrSchedulingConflictInvalidReschedule
	}

	*request.NewDate = dateValue
	*request.NewStartTime = startValue
	*request.NewEndTime = endValue

	return nil
}
