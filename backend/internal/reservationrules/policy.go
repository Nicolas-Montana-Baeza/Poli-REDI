package reservationrules

import (
	"fmt"
	"time"
)

// ValidateReservableWindow compares institutional calendar dates. A window of
// seven days includes today and the following six dates.
func ValidateReservableWindow(now, start time.Time, windowDays int) error {
	if windowDays <= 0 {
		return fmt.Errorf("la ventana reservable configurada no es v\u00e1lida")
	}

	today := calendarDate(now)
	startDate := calendarDate(start)
	lastDate := today.AddDate(0, 0, windowDays-1)

	if startDate.Before(today) || startDate.After(lastDate) {
		return fmt.Errorf(
			"la fecha debe estar entre %s y %s",
			today.Format("2006-01-02"),
			lastDate.Format("2006-01-02"),
		)
	}

	return nil
}

func NextRequestDate(createdAt time.Time, frequencyDays int, location *time.Location) time.Time {
	createdLocal := createdAt.In(location)
	createdDate := calendarDate(createdLocal)
	return createdDate.AddDate(0, 0, frequencyDays)
}

func calendarDate(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}
