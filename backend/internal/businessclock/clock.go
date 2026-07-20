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

// ParseDateTime normaliza fechas de agenda enviadas por el cliente hacia la zona
// institucional. Los valores con offset explicito se convierten a APP_TIMEZONE;
// los valores sin offset se interpretan como hora de muro en esa misma zona.
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

// FromDatabaseWallTime asigna la zona institucional sin mover la fecha ni la
// hora almacenadas en SQL Server DATETIME2. Los DATETIME2 de reservas representan
// hora de muro chilena, no instantes UTC.
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

// ToDatabaseWallTime quita la zona manteniendo los campos de hora institucional
// esperados por SQL Server DATETIME2. La zona UTC retornada solo transporta el
// valor para database/sql; no debe interpretarse como instante UTC de agenda.
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
