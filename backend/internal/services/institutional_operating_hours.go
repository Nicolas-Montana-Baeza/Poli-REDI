package services

import (
	"fmt"
	"time"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
)

// La ventana del calendario se basa actualmente en la jornada configurada
// en la política de reservas vigente. Hasta modelar horarios de apertura
// independientes por instalación, la programación institucional debe respetarla
// también, para no crear ocupaciones fuera del calendario visible.
func ensureInstitutionalOperatingHours(
	schedules []models.CreateInstitutionalScheduleRequest,
) error {
	policy, err := repositories.GetCurrentReservationPolicyComplete()
	if err != nil {
		return fmt.Errorf("no se pudo consultar la jornada vigente: %w", err)
	}
	return validateInstitutionalOperatingHours(
		schedules, policy.OpeningMinute, policy.ClosingMinute,
	)
}

func validateInstitutionalOperatingHours(
	schedules []models.CreateInstitutionalScheduleRequest,
	openingMinute, closingMinute int,
) error {
	if openingMinute < 0 || closingMinute > 1440 ||
		openingMinute >= closingMinute {
		return fmt.Errorf("%w: jornada vigente inválida", ErrInstitutionalScheduleInvalid)
	}

	for _, schedule := range schedules {
		start, err := time.Parse("15:04", schedule.StartTime)
		if err != nil {
			return ErrInstitutionalScheduleInvalid
		}
		end, err := time.Parse("15:04", schedule.EndTime)
		if err != nil || !end.After(start) {
			return ErrInstitutionalScheduleInvalid
		}
		startMinute := start.Hour()*60 + start.Minute()
		endMinute := end.Hour()*60 + end.Minute()
		if startMinute < openingMinute || endMinute > closingMinute {
			return fmt.Errorf(
				"%w: el horario %s–%s debe estar dentro de la jornada %02d:%02d–%02d:%02d",
				ErrInstitutionalScheduleInvalid,
				schedule.StartTime, schedule.EndTime,
				openingMinute/60, openingMinute%60,
				closingMinute/60, closingMinute%60,
			)
		}
	}
	return nil
}
