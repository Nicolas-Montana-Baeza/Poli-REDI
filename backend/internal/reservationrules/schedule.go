package reservationrules

import (
	"errors"
	"fmt"
	"time"
)

const (
	OpeningHour = 8
	ClosingHour = 22
	SlotMinutes = 15
)

var allowedDurations = map[int]struct{}{
	30:  {},
	60:  {},
	90:  {},
	120: {},
	150: {},
	180: {},
}

func ValidateSchedule(start time.Time, durationMinutes int) error {
	durations := make([]int, 0, len(allowedDurations))
	for duration := range allowedDurations {
		durations = append(durations, duration)
	}
	return ValidateScheduleWithPolicy(start, durationMinutes, OpeningHour*60, ClosingHour*60, SlotMinutes, durations)
}

func ValidateScheduleWithPolicy(start time.Time, durationMinutes, openingMinute, closingMinute, slotMinutes int, allowedDurations []int) error {
	allowed := false
	for _, duration := range allowedDurations {
		if duration == durationMinutes {
			allowed = true
			break
		}
	}
	if !allowed {
		return errors.New("selecciona una duracion de 30 a 180 minutos en intervalos de 30")
	}

	minuteOfDay := start.Hour()*60 + start.Minute()
	if slotMinutes <= 0 || start.Second() != 0 || start.Nanosecond() != 0 || minuteOfDay%slotMinutes != 0 {
		return fmt.Errorf("la hora de inicio debe usar intervalos de %d minutos", slotMinutes)
	}

	opening := time.Date(
		start.Year(), start.Month(), start.Day(),
		openingMinute/60, openingMinute%60, 0, 0, start.Location(),
	)
	closing := time.Date(
		start.Year(), start.Month(), start.Day(),
		closingMinute/60, closingMinute%60, 0, 0, start.Location(),
	)

	if start.Before(opening) {
		return fmt.Errorf("la jornada de reservas comienza a las %02d:%02d", openingMinute/60, openingMinute%60)
	}

	if !start.Before(closing) {
		return fmt.Errorf("la hora de inicio debe ser anterior a las %02d:%02d", closingMinute/60, closingMinute%60)
	}

	end := start.Add(time.Duration(durationMinutes) * time.Minute)
	if end.After(closing) || !sameDate(start, end) {
		return fmt.Errorf("la reserva debe finalizar a mas tardar a las %02d:%02d", closingMinute/60, closingMinute%60)
	}

	return nil
}

func sameDate(left time.Time, right time.Time) bool {
	return left.Year() == right.Year() && left.YearDay() == right.YearDay()
}
