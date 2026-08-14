package services

import (
	"database/sql"
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"poli-redi-api/internal/appscope"
	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
	"poli-redi-api/internal/reservationrules"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrReservationNotFound  = errors.New("reserva no encontrada")
	ErrReservationForbidden = errors.New("no tienes permisos para consultar esta reserva")
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

func GetAvailabilityItems(from, to time.Time, userID int, isAdmin bool) ([]models.AvailabilityItem, error) {
	reservations, err := repositories.GetActiveReservationsForAvailability(from, to, userID, isAdmin)

	if err != nil {
		return nil, err
	}

	blocks, err := repositories.GetAvailabilityBlocks(from, to)

	if err != nil {
		return nil, err
	}

	items := make(
		[]models.AvailabilityItem,
		0,
		len(reservations)+len(blocks),
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

	for _, block := range blocks {
		duration := int(block.EndTime.Sub(block.StartTime).Minutes())
		items = append(items, models.AvailabilityItem{
			ID:                  block.ID,
			AvailabilityKey:     "block-" + strconv.Itoa(block.ID),
			ResourceID:          block.ResourceID,
			StartTime:           block.StartTime,
			DurationMinutes:     duration,
			Status:              "CONFIRMED",
			Hour:                block.StartTime.Format("15:04"),
			Title:               "No disponible",
			Type:                "blocked",
			ResourceName:        block.ResourceName,
			IsAvailabilityBlock: true,
			ActivityType:        block.BlockType,
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

func GetReservationDetail(id int, requestedBy models.LocalAuthUser) (models.Reservation, error) {
	reservation, err := repositories.GetReservationByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Reservation{}, ErrReservationNotFound
	}
	if err != nil {
		return models.Reservation{}, err
	}
	if !requestedBy.IsAdmin && reservation.UserID != requestedBy.ID {
		return models.Reservation{}, ErrReservationForbidden
	}
	return reservation, nil
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

	resource, err := repositories.GetResourceByID(reservation.ResourceID)

	if err != nil {
		return models.Reservation{}, err
	}

	if resource.ReservationMode == "OPEN_USE" {
		reservation.ActivityID = nil
	}

	if !appscope.IsMVP1() {
		if err := validateWorkshopAvailability(reservation); err != nil {
			return models.Reservation{}, err
		}
	}

	createdReservation, err := repositories.AddReservationWithPolicy(reservation, func(policy models.ReservationPolicy) error {
		return validateReservationPolicySnapshot(reservation, now, policy)
	})

	if err != nil {
		return models.Reservation{}, mapDatabaseReservationError(err)
	}

	return createdReservation, nil
}

func validateReservationPolicySnapshot(reservation models.Reservation, now time.Time, policy models.ReservationPolicy) error {
	if err := reservationrules.ValidateScheduleWithPolicy(reservation.StartTime, reservation.DurationMinutes,
		policy.OpeningMinute, policy.ClosingMinute, policy.SlotIntervalMinutes, policy.AllowedDurations); err != nil {
		return err
	}
	return reservationrules.ValidateReservableWindow(now, reservation.StartTime, policy.ReservableWindowDays)
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
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("no existe una pol\u00edtica de reservas vigente")
	}
	if errors.Is(err, repositories.ErrResourceNotAllowedByPolicy) {
		return err
	}
	var sqlErr *pgconn.PgError
	if errors.As(err, &sqlErr) {
		switch sqlErr.Code {
		case "P1001":
			return errors.New("el usuario se encuentra bloqueado o no tiene RUT habilitante")
		case "P1002":
			return errors.New("el recurso no esta disponible para reservas")
		case "P1003":
			return repositories.ErrResourceNotAllowedByPolicy
		case "P1004", "P1005":
			return errors.New("el horario o la duracion no estan permitidos por la politica vigente")
		case "P1006":
			return errors.New("la fecha esta fuera de la ventana reservable")
		case "P1007":
			return errors.New("el recurso esta bloqueado en ese horario")
		case "P1008":
			return errors.New("no existe una politica de reservas vigente")
		case "23P01":
			if sqlErr.ConstraintName == "ex_reservations_user_overlap" {
				return errors.New("el usuario ya tiene una reserva en ese horario")
			}
			return errors.New("el recurso ya esta reservado en ese horario")
		case "P1009", "23503", "23514", "23502":
			return errors.New("usuario, recurso o actividad no existe, o los datos no cumplen restricciones")
		case "23505":
			return errors.New("ya existe un registro con esos datos")
		case "40001", "40P01":
			return errors.New("la reserva compitio con otra operacion; intenta nuevamente")
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

	cancelledReservation, err := repositories.CancelReservationAuthorized(reservationID, requestedByUser, now)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Reservation{}, ErrReservationNotFound
	}
	if errors.Is(err, repositories.ErrReservationForbidden) {
		return models.Reservation{}, errors.New("no tienes permisos para cancelar esta reserva")
	}
	if errors.Is(err, repositories.ErrReservationFinalized) {
		return models.Reservation{}, repositories.ErrReservationFinalized
	}
	if errors.Is(err, repositories.ErrReservationNotCancellable) {
		return models.Reservation{}, repositories.ErrReservationNotCancellable
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
