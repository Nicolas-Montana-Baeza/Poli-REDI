package businessclock

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	_ "time/tzdata"
)

const DefaultLocationName = "America/Santiago"

var (
	locationMu       sync.RWMutex
	businessLocation = loadDefaultLocation()
)

func loadDefaultLocation() *time.Location {
	location, err := time.LoadLocation(DefaultLocationName)
	if err != nil {
		panic(err)
	}

	return location
}

func Configure(locationName string) error {
	locationName = strings.TrimSpace(locationName)
	if locationName == "" {
		locationName = DefaultLocationName
	}

	location, err := time.LoadLocation(locationName)
	if err != nil {
		return fmt.Errorf("APP_TIMEZONE %q no es valida: %w", locationName, err)
	}

	locationMu.Lock()
	businessLocation = location
	locationMu.Unlock()

	return nil
}

func Location() *time.Location {
	locationMu.RLock()
	defer locationMu.RUnlock()

	return businessLocation
}

func Now() time.Time {
	return time.Now().In(Location())
}

func ParseDateTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errors.New("startTime es obligatorio")
	}

	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return parsed.In(Location()), nil
	}

	for _, layout := range []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
	} {
		if parsed, err := time.ParseInLocation(layout, value, Location()); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, errors.New("fecha de inicio invalida")
}

// FromDatabaseWallTime assigns the institutional zone without moving the
// date or clock fields stored in SQL Server DATETIME2.
func FromDatabaseWallTime(value time.Time) time.Time {
	return time.Date(
		value.Year(),
		value.Month(),
		value.Day(),
		value.Hour(),
		value.Minute(),
		value.Second(),
		value.Nanosecond(),
		Location(),
	)
}

// ToDatabaseWallTime strips the zone while preserving the institutional
// wall-clock fields expected by SQL Server DATETIME2.
func ToDatabaseWallTime(value time.Time) time.Time {
	localValue := value.In(Location())

	return time.Date(
		localValue.Year(),
		localValue.Month(),
		localValue.Day(),
		localValue.Hour(),
		localValue.Minute(),
		localValue.Second(),
		localValue.Nanosecond(),
		time.UTC,
	)
}
