package reservationrules

import (
	"errors"
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
	if _, ok := allowedDurations[durationMinutes]; !ok {
		return errors.New("selecciona una duracion de 30 a 180 minutos en intervalos de 30")
	}

	if start.Second() != 0 || start.Nanosecond() != 0 || start.Minute()%SlotMinutes != 0 {
		return errors.New("la hora de inicio debe usar intervalos de 15 minutos")
	}

	opening := time.Date(
		start.Year(), start.Month(), start.Day(),
		OpeningHour, 0, 0, 0, start.Location(),
	)
	closing := time.Date(
		start.Year(), start.Month(), start.Day(),
		ClosingHour, 0, 0, 0, start.Location(),
	)

	if start.Before(opening) {
		return errors.New("la jornada de reservas comienza a las 08:00")
	}

	if !start.Before(closing) {
		return errors.New("la hora de inicio debe ser anterior a las 22:00")
	}

	end := start.Add(time.Duration(durationMinutes) * time.Minute)
	if end.After(closing) || !sameDate(start, end) {
		return errors.New("la reserva debe finalizar a mas tardar a las 22:00")
	}

	return nil
}

func sameDate(left time.Time, right time.Time) bool {
	return left.Year() == right.Year() && left.YearDay() == right.YearDay()
}
