package services

import (
	"database/sql"
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
	"poli-redi-api/internal/reservationrules"

	mssql "github.com/microsoft/go-mssqldb"
)

var workshopTimePattern = regexp.MustCompile(`(?:(lunes|martes|miercoles|jueves|viernes|sabado|domingo)\s+)?(\d{1,2}:\d{2})\s*a\s*(\d{1,2}:\d{2})`)

var dayNames = []string{
	"domingo",
	"lunes",
	"martes",
	"miercoles",
	"jueves",
	"viernes",
	"sabado",
}

func GetReservations() ([]models.Reservation, error) {
	return repositories.GetAllReservations()
}

func GetAvailabilityItems() ([]models.AvailabilityItem, error) {
	reservations, err := repositories.GetAllReservations()

	if err != nil {
		return nil, err
	}

	scheduledActivities, err := repositories.GetActiveScheduledActivities()

	if err != nil {
		return nil, err
	}

	items := make(
		[]models.AvailabilityItem,
		0,
		len(reservations)+len(scheduledActivities),
	)

	for _, reservation := range reservations {
		items = append(items, models.AvailabilityItem{
			ID:              reservation.ID,
			AvailabilityKey: "reservation-" + strconv.Itoa(reservation.ID),
			UserID:          reservation.UserID,
			ResourceID:      reservation.ResourceID,
			StartTime:       reservation.StartTime,
			DurationMinutes: reservation.DurationMinutes,
			Status:          reservation.Status,
			Hour:            reservation.Hour,
			Title:           reservation.Title,
			Type:            reservation.Type,
			ResourceName:    reservation.ResourceName,
			UserFullName:    reservation.UserFullName,
			UserEmail:       reservation.UserEmail,
			UserRUT:         reservation.UserRUT,
		})
	}

	for _, activity := range scheduledActivities {
		duration := int(activity.EndTime.Sub(activity.StartTime).Minutes())

		items = append(items, models.AvailabilityItem{
			ID:                  activity.ID,
			AvailabilityKey:     "scheduled-" + strconv.Itoa(activity.ID),
			UserID:              activity.CreatedByUserID,
			ResourceID:          activity.ResourceID,
			StartTime:           activity.StartTime,
			DurationMinutes:     duration,
			Status:              "CONFIRMED",
			Hour:                activity.StartTime.Format("15:04"),
			Title:               activity.Title,
			Type:                "scheduled",
			ResourceName:        activity.ResourceName,
			IsScheduledActivity: true,
			ActivityType:        activity.ActivityType,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].StartTime.Before(items[j].StartTime)
	})

	return items, nil
}

func GetMyReservations(userID int) ([]models.Reservation, error) {
	if userID <= 0 {
		return nil, errors.New("usuario autenticado es obligatorio")
	}

	return repositories.GetReservationsByUserID(userID)
}

func CreateReservation(reservation models.Reservation) (models.Reservation, error) {
	return createReservationAt(reservation, businessclock.Now())
}

func createReservationAt(
	reservation models.Reservation,
	now time.Time,
) (models.Reservation, error) {
	reservation = enforceInitialReservationStatus(reservation)

	if reservation.UserID == 0 {
		return models.Reservation{}, errors.New("no se pudo identificar al usuario autenticado")
	}

	if reservation.ResourceID == 0 {
		return models.Reservation{}, errors.New("selecciona una instalaci\u00f3n")
	}

	if reservation.StartTime.IsZero() {
		return models.Reservation{}, errors.New("selecciona una fecha y hora de inicio")
	}

	if reservation.StartTime.Before(now) {
		return models.Reservation{}, errors.New("no puedes crear reservas en el pasado")
	}

	// La validacion de horario entrega errores de dominio antes del INSERT. Los
	// triggers de SQL Server siguen protegiendo conflictos entre filas porque dos
	// requests concurrentes pueden saltarse cualquier chequeo solo en memoria.
	if err := reservationrules.ValidateSchedule(
		reservation.StartTime,
		reservation.DurationMinutes,
	); err != nil {
		return models.Reservation{}, err
	}

	resource, err := repositories.GetResourceByID(reservation.ResourceID)

	if err != nil {
		return models.Reservation{}, err
	}

	if resource.ReservationMode == "OPEN_USE" {
		reservation.ActivityID = nil
	}

	if err := validateWorkshopAvailability(reservation); err != nil {
		return models.Reservation{}, err
	}

	createdReservation, err := repositories.AddReservation(reservation)

	if err != nil {
		return models.Reservation{}, mapDatabaseReservationError(err)
	}

	return createdReservation, nil
}

func enforceInitialReservationStatus(
	reservation models.Reservation,
) models.Reservation {
	reservation.Status = models.ReservationStatusConfirmed
	return reservation
}

func validateWorkshopAvailability(reservation models.Reservation) error {
	resource, err := repositories.GetResourceByID(reservation.ResourceID)

	if err != nil {
		return err
	}

	if resource.ReservationMode == "OPEN_USE" {
		return nil
	}

	workshops, err := repositories.GetActiveWorkshops()

	if err != nil {
		return err
	}

	reservationStart := reservation.StartTime
	reservationEnd := reservationStart.Add(
		time.Duration(reservation.DurationMinutes) * time.Minute,
	)

	for _, workshop := range workshops {
		if workshop.ResourceID != reservation.ResourceID {
			continue
		}

		if !workshopOccursOnDate(workshop, reservationStart) {
			continue
		}

		for _, timeRange := range workshopTimeRangesForDate(workshop, reservationStart) {
			workshopStart := time.Date(
				reservationStart.Year(),
				reservationStart.Month(),
				reservationStart.Day(),
				timeRange.startHour,
				timeRange.startMinute,
				0,
				0,
				reservationStart.Location(),
			)

			workshopEnd := time.Date(
				reservationStart.Year(),
				reservationStart.Month(),
				reservationStart.Day(),
				timeRange.endHour,
				timeRange.endMinute,
				0,
				0,
				reservationStart.Location(),
			)

			if reservationStart.Before(workshopEnd) && reservationEnd.After(workshopStart) {
				return errors.New("el recurso tiene un taller programado en ese horario")
			}
		}
	}

	return nil
}

type workshopTimeRange struct {
	startHour   int
	startMinute int
	endHour     int
	endMinute   int
}

func workshopOccursOnDate(workshop models.Workshop, date time.Time) bool {
	selectedDay := dayNames[int(date.Weekday())]
	dayText := normalizeWorkshopText(workshop.DayText)

	return strings.Contains(dayText, selectedDay) ||
		dayMatchesWorkshopRange(dayText, selectedDay)
}

func dayMatchesWorkshopRange(text string, selectedDay string) bool {
	for startIndex, startDay := range dayNames {
		for endIndex, endDay := range dayNames {
			if !strings.Contains(text, startDay+" a "+endDay) {
				continue
			}

			selectedIndex := dayIndex(selectedDay)

			if selectedIndex < 0 {
				return false
			}

			if startIndex <= endIndex {
				return selectedIndex >= startIndex && selectedIndex <= endIndex
			}

			return selectedIndex >= startIndex || selectedIndex <= endIndex
		}
	}

	return false
}

func workshopTimeRangesForDate(
	workshop models.Workshop,
	date time.Time,
) []workshopTimeRange {
	selectedDay := dayNames[int(date.Weekday())]
	scheduleText := normalizeWorkshopText(workshop.ScheduleText)
	matches := workshopTimePattern.FindAllStringSubmatch(scheduleText, -1)
	ranges := []workshopTimeRange{}

	for _, match := range matches {
		if len(match) != 4 {
			continue
		}

		explicitDay := match[1]

		if explicitDay != "" && explicitDay != selectedDay {
			continue
		}

		startHour, startMinute, ok := parseWorkshopHour(match[2])

		if !ok {
			continue
		}

		endHour, endMinute, ok := parseWorkshopHour(match[3])

		if !ok {
			continue
		}

		if endHour*60+endMinute <= startHour*60+startMinute {
			continue
		}

		ranges = append(ranges, workshopTimeRange{
			startHour:   startHour,
			startMinute: startMinute,
			endHour:     endHour,
			endMinute:   endMinute,
		})
	}

	return ranges
}

func parseWorkshopHour(value string) (int, int, bool) {
	parts := strings.Split(value, ":")

	if len(parts) != 2 {
		return 0, 0, false
	}

	hour, err := strconv.Atoi(parts[0])

	if err != nil {
		return 0, 0, false
	}

	minute, err := strconv.Atoi(parts[1])

	if err != nil {
		return 0, 0, false
	}

	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return 0, 0, false
	}

	return hour, minute, true
}

func dayIndex(day string) int {
	for index, name := range dayNames {
		if name == day {
			return index
		}
	}

	return -1
}

func normalizeWorkshopText(value string) string {
	replacer := strings.NewReplacer(
		"á", "a",
		"é", "e",
		"í", "i",
		"ó", "o",
		"ú", "u",
		"ñ", "n",
		"Ã¡", "a",
		"Ã©", "e",
		"Ã­", "i",
		"Ã³", "o",
		"Ãº", "u",
		"Ã±", "n",
	)

	return replacer.Replace(strings.ToLower(value))
}

func mapDatabaseReservationError(err error) error {
	var sqlErr mssql.Error

	if errors.As(err, &sqlErr) {
		switch sqlErr.Number {
		// 51000-51007 son reglas de negocio emitidas por
		// trg_reservations_validate_conflicts en database/schema.sql.
		case 51000:
			return errors.New("el usuario se encuentra bloqueado y no puede crear reservas")
		case 51001:
			return errors.New("el recurso no est\u00e1 activo")
		case 51002:
			return errors.New("el recurso es solo informativo y no permite reservas")
		case 51003:
			return errors.New("el recurso solo puede ser reservado por administradores")
		case 51004:
			return errors.New("el recurso ya est\u00e1 reservado en ese horario")
		case 51005:
			return errors.New("el usuario ya tiene una reserva en ese horario")
		case 51006:
			return errors.New("el recurso est\u00e1 bloqueado en ese horario")
		case 51007:
			return errors.New("el recurso tiene una actividad programada en ese horario")
		case 547:
			return errors.New("usuario, recurso o actividad no existe, o los datos no cumplen restricciones")
		case 2601, 2627:
			return errors.New("ya existe un registro con esos datos")
		default:
			return errors.New(sqlErr.Message)
		}
	}

	return err
}

func CancelReservation(
	reservationID int,
	requestedByUser models.LocalAuthUser,
) (models.Reservation, error) {
	return cancelReservationAt(
		reservationID,
		requestedByUser,
		businessclock.Now(),
	)
}

func cancelReservationAt(
	reservationID int,
	requestedByUser models.LocalAuthUser,
	now time.Time,
) (models.Reservation, error) {
	if reservationID <= 0 {
		return models.Reservation{}, errors.New("no se pudo identificar la reserva")
	}

	if requestedByUser.ID <= 0 {
		return models.Reservation{}, errors.New("usuario autenticado es obligatorio")
	}

	ownerID, status, startTime, durationMinutes, err :=
		repositories.GetReservationCancellationSnapshot(reservationID)

	if errors.Is(err, sql.ErrNoRows) {
		return models.Reservation{}, errors.New("reserva no encontrada")
	}

	if err != nil {
		return models.Reservation{}, err
	}

	if !requestedByUser.IsAdmin && ownerID != requestedByUser.ID {
		return models.Reservation{}, errors.New("no tienes permisos para cancelar esta reserva")
	}

	if err := validateCancellationStatus(status); err != nil {
		return models.Reservation{}, err
	}

	reservationEnd := startTime.Add(
		time.Duration(durationMinutes) * time.Minute,
	)

	if !reservationEnd.After(now) {
		return models.Reservation{}, errors.New("no puedes cancelar una reserva finalizada")
	}

	cancelledReservation, err := repositories.CancelReservation(reservationID)

	if errors.Is(err, sql.ErrNoRows) {
		return models.Reservation{}, errors.New("la reserva ya no se puede cancelar")
	}

	if err != nil {
		return models.Reservation{}, err
	}

	return cancelledReservation, nil
}

func validateCancellationStatus(status string) error {
	switch status {
	case models.ReservationStatusConfirmed, models.ReservationStatusPending:
		return nil
	case models.ReservationStatusCancelled:
		return errors.New("la reserva ya est\u00e1 cancelada")
	default:
		return errors.New("la reserva en este estado no se puede cancelar")
	}
}
