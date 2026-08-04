package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
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
			ID:                   reservation.ID,
			AvailabilityKey:      "reservation-" + strconv.Itoa(reservation.ID),
			AvailabilityKind:     reservationAvailabilityKind(reservation),
			UserID:               reservation.UserID,
			ResourceID:           reservation.ResourceID,
			StartTime:            reservation.StartTime,
			DurationMinutes:      reservation.DurationMinutes,
			Status:               reservation.Status,
			Hour:                 reservation.Hour,
			Title:                reservation.Title,
			Type:                 reservation.Type,
			ResourceName:         reservation.ResourceName,
			UserFullName:         reservation.UserFullName,
			UserEmail:            reservation.UserEmail,
			UserRUT:              reservation.UserRUT,
			ParticipantCount:     reservation.ParticipantCount,
			MinimumParticipants:  reservation.MinimumParticipants,
			TargetParticipants:   reservation.TargetParticipants,
			Capacity:             reservation.Capacity,
			ConfirmationDeadline: reservation.ConfirmationDeadline,
		})
	}

	for _, activity := range scheduledActivities {
		duration := int(activity.EndTime.Sub(activity.StartTime).Minutes())

		items = append(items, models.AvailabilityItem{
			ID:                  activity.ID,
			AvailabilityKey:     "scheduled-" + strconv.Itoa(activity.ID),
			AvailabilityKind:    models.AvailabilityKindScheduled,
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

func GetAvailabilityItemsForRange(
	from time.Time,
	toExclusive time.Time,
	userID int,
) ([]models.AvailabilityItem, error) {
	reservations, err := repositories.GetAvailabilityReservationsForRange(from, toExclusive, userID)
	if err != nil {
		return nil, err
	}
	scheduledActivities, err := repositories.GetScheduledActivitiesForRange(from, toExclusive)
	if err != nil {
		return nil, err
	}
	workshops, err := repositories.GetWorkshopOccurrencesForRange(from, toExclusive)
	if err != nil {
		return nil, err
	}
	blocks, err := repositories.GetAvailabilityBlocksForRange(from, toExclusive)
	if err != nil {
		return nil, err
	}
	return buildAvailabilityItems(reservations, scheduledActivities, workshops, blocks), nil
}

func buildAvailabilityItems(
	reservations []models.AvailabilityReservation,
	scheduledActivities []models.ScheduledActivity,
	workshops []models.WorkshopAvailabilityOccurrence,
	blocks []models.AvailabilityBlock,
) []models.AvailabilityItem {
	items := make([]models.AvailabilityItem, 0,
		len(reservations)+len(scheduledActivities)+len(workshops)+len(blocks))
	for _, source := range reservations {
		reservation := source.Reservation
		items = append(items, models.AvailabilityItem{
			ID:                    reservation.ID,
			AvailabilityKey:       "reservation-" + strconv.Itoa(reservation.ID),
			AvailabilityKind:      models.AvailabilityKindReservation,
			UserID:                reservation.UserID,
			ResourceID:            reservation.ResourceID,
			StartTime:             reservation.StartTime,
			DurationMinutes:       reservation.DurationMinutes,
			Status:                reservation.Status,
			Hour:                  reservation.Hour,
			Title:                 reservation.Title,
			Type:                  reservation.Type,
			ItemType:              reservation.Type,
			ResourceName:          reservation.ResourceName,
			UserFullName:          reservation.UserFullName,
			UserEmail:             reservation.UserEmail,
			UserRUT:               reservation.UserRUT,
			ParticipantCount:      reservation.ParticipantCount,
			MinimumParticipants:   reservation.MinimumParticipants,
			TargetParticipants:    reservation.TargetParticipants,
			Capacity:              reservation.Capacity,
			ConfirmationDeadline:  reservation.ConfirmationDeadline,
			BlocksResource:        source.ReservationMode != "OPEN_USE",
			IsCurrentUserConflict: source.IsCurrentUserConflict,
		})
	}
	for _, source := range scheduledActivities {
		items = append(items, availabilityItemFromInterval(
			source.ID,
			"scheduled-"+strconv.Itoa(source.ID),
			models.AvailabilityKindScheduled,
			source.ResourceID,
			source.StartTime,
			source.EndTime,
			source.Title,
			"scheduled",
			source.ResourceName,
			source.ActivityType,
			source.CreatedByUserID,
			source.ReservationMode != "OPEN_USE",
		))
	}
	for _, source := range workshops {
		items = append(items, availabilityItemFromInterval(
			source.ID,
			"workshop-occurrence-"+strconv.Itoa(source.ID)+"-"+source.StartTime.Format("20060102"),
			models.AvailabilityKindWorkshop,
			source.ResourceID,
			source.StartTime,
			source.EndTime,
			source.Title,
			"workshop",
			source.ResourceName,
			"WORKSHOP",
			0,
			source.ReservationMode != "OPEN_USE",
		))
	}
	for _, source := range blocks {
		title := source.Reason
		if strings.TrimSpace(title) == "" {
			title = source.BlockType
		}
		item := availabilityItemFromInterval(
			source.ID,
			"availability-block-"+strconv.Itoa(source.ID),
			models.AvailabilityKindBlock,
			source.ResourceID,
			source.StartTime,
			source.EndTime,
			title,
			"block",
			source.ResourceName,
			source.BlockType,
			source.CreatedByUserID,
			true,
		)
		items = append(items, item)
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].StartTime.Equal(items[j].StartTime) {
			return items[i].AvailabilityKey < items[j].AvailabilityKey
		}
		return items[i].StartTime.Before(items[j].StartTime)
	})
	return items
}

func availabilityItemFromInterval(
	id int,
	key string,
	kind string,
	resourceID int,
	startTime time.Time,
	endTime time.Time,
	title string,
	itemType string,
	resourceName string,
	activityType string,
	userID int,
	blocksResource bool,
) models.AvailabilityItem {
	return models.AvailabilityItem{
		ID:                    id,
		AvailabilityKey:       key,
		AvailabilityKind:      kind,
		UserID:                userID,
		ResourceID:            resourceID,
		StartTime:             startTime,
		DurationMinutes:       int(endTime.Sub(startTime).Minutes()),
		Status:                models.ReservationStatusConfirmed,
		Hour:                  startTime.Format("15:04"),
		Title:                 title,
		Type:                  itemType,
		ItemType:              itemType,
		ResourceName:          resourceName,
		IsScheduledActivity:   kind == models.AvailabilityKindScheduled || kind == models.AvailabilityKindWorkshop,
		ActivityType:          activityType,
		BlocksResource:        blocksResource,
		IsCurrentUserConflict: false,
	}
}

func reservationAvailabilityKind(reservation models.Reservation) string {
	if reservation.TargetParticipants != nil {
		return models.AvailabilityKindGroupReservation
	}

	return models.AvailabilityKindReservation
}

// SanitizeAvailabilityItemsForAudience applies the shared-calendar audience
// contract. Administrators keep operational detail. A regular user can identify
// their own reservation, while reservations owned by somebody else expose only
// occupancy and a non-personal kind. Institutional programming keeps its public
// category but not the creator's identity or its internal title.
func SanitizeAvailabilityItemsForAudience(
	items []models.AvailabilityItem,
	audience models.LocalAuthUser,
) []models.AvailabilityItem {
	sanitized := append([]models.AvailabilityItem(nil), items...)

	if audience.IsAdmin {
		return sanitized
	}

	for index := range sanitized {
		item := &sanitized[index]
		isReservation := item.AvailabilityKind == models.AvailabilityKindReservation ||
			item.AvailabilityKind == models.AvailabilityKindGroupReservation ||
			(item.AvailabilityKind == "" && !item.IsScheduledActivity)
		isOwnReservation := isReservation && item.UserID == audience.ID

		item.UserFullName = ""
		item.UserEmail = ""
		item.UserRUT = ""

		switch item.AvailabilityKind {
		case models.AvailabilityKindWorkshop:
			item.UserID = 0
			item.CanEditTarget = false
			continue
		case models.AvailabilityKindBlock:
			item.UserID = 0
			item.Title = "No disponible"
			item.ActivityType = ""
			item.CanEditTarget = false
			continue
		case models.AvailabilityKindScheduled:
			item.UserID = 0
			item.Title = "Actividad institucional"
			item.CanEditTarget = false
			continue
		}

		if item.IsScheduledActivity {
			item.UserID = 0
			item.Title = "Actividad institucional"
			item.CanEditTarget = false
			continue
		}

		if isOwnReservation {
			continue
		}

		item.UserID = 0
		item.Title = "Reserva"
		item.ParticipantCount = 0
		item.MinimumParticipants = 0
		item.TargetParticipants = nil
		item.Capacity = nil
		item.ConfirmationDeadline = nil
		item.CanEditTarget = false
	}

	return sanitized
}

func GetMyReservations(userID int) ([]models.Reservation, error) {
	if userID <= 0 {
		return nil, errors.New("usuario autenticado es obligatorio")
	}

	reservations, err := repositories.GetReservationsByUserID(userID)
	if err != nil {
		return nil, err
	}
	return sanitizeMyReservations(reservations, userID), nil
}

// sanitizeMyReservations applies the audience contract of GET /reservations/mine.
// A user may receive reservations owned by them and reservations where they are
// a confirmed participant. Only owners need their own identity fields in this
// response; participants must not learn the owner's identity or identifiers.
func sanitizeMyReservations(reservations []models.Reservation, audienceUserID int) []models.Reservation {
	for index := range reservations {
		if reservations[index].UserID == audienceUserID {
			continue
		}
		reservations[index].UserID = 0
		reservations[index].UserFullName = ""
		reservations[index].UserEmail = ""
		reservations[index].UserRUT = ""
		reservations[index].CanEditTarget = false
	}
	return reservations
}

func CreateReservation(reservation models.Reservation) (models.Reservation, error) {
	return createReservationAt(reservation, businessclock.Now())
}

func createReservationAt(
	reservation models.Reservation,
	now time.Time,
) (models.Reservation, error) {
	codeBytes := make([]byte, 24)
	if _, err := rand.Read(codeBytes); err != nil {
		return models.Reservation{}, err
	}
	reservation.JoinCode = base64.RawURLEncoding.EncodeToString(codeBytes)

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

	if !modeConsumesRequestFrequency(resource.ReservationMode) {
		reservation.ActivityID = nil
	} else {
		previousCreatedAt, frequencyDays, err := repositories.GetLatestConsumingReservation(reservation.UserID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return models.Reservation{}, err
		}
		if err == nil {
			if err := validateRequestFrequency(previousCreatedAt, frequencyDays, now); err != nil {
				return models.Reservation{}, err
			}
		}
	}

	if err := validateWorkshopAvailability(reservation); err != nil {
		return models.Reservation{}, err
	}

	createdReservation, err := repositories.AddReservationWithPolicy(reservation, func(policy models.ReservationPolicy) error {
		return validateReservationPolicySnapshot(reservation, now, policy)
	})

	if err != nil {
		return models.Reservation{}, mapDatabaseReservationError(err)
	}

	return createdReservation, nil
}

func validateRequestFrequency(previousCreatedAt time.Time, frequencyDays int, now time.Time) error {
	nextDate := reservationrules.NextRequestDate(
		previousCreatedAt,
		frequencyDays,
		businessclock.Location(),
	)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if today.Before(nextDate) {
		return fmt.Errorf(
			"A\u00fan no puedes enviar otra solicitud. Podr\u00e1s hacerlo desde el %s; la espera depende de cu\u00e1ndo env\u00edas la solicitud, no de la fecha que quieres reservar.",
			nextDate.Format("02/01/2006"),
		)
	}
	return nil
}

func modeConsumesRequestFrequency(reservationMode string) bool {
	return reservationMode != "OPEN_USE"
}

func validateReservationPolicySnapshot(reservation models.Reservation, now time.Time, policy models.ReservationPolicy) error {
	if err := reservationrules.ValidateScheduleWithPolicy(reservation.StartTime, reservation.DurationMinutes,
		policy.OpeningMinute, policy.ClosingMinute, policy.SlotIntervalMinutes, policy.AllowedDurations); err != nil {
		return err
	}
	return reservationrules.ValidateReservableWindow(now, reservation.StartTime, policy.ReservableWindowDays)
}

func validateWorkshopAvailability(reservation models.Reservation) error {
	resource, err := repositories.GetResourceByID(reservation.ResourceID)

	if err != nil {
		return err
	}

	if resource.ReservationMode == "OPEN_USE" {
		return nil
	}

	reservationStart := reservation.StartTime
	reservationEnd := reservationStart.Add(
		time.Duration(reservation.DurationMinutes) * time.Minute,
	)
	conflict, err := repositories.HasWorkshopAvailabilityConflict(
		reservation.ResourceID,
		reservationStart,
		reservationEnd,
	)
	if err != nil {
		return err
	}
	if conflict {
		return errors.New("el recurso tiene un taller programado en ese horario")
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
		case 51008:
			return errors.New("no existe una pol\u00edtica de reservas vigente")
		case 51009, 51010:
			return errors.New(sqlErr.Message)
		case 51015:
			return errors.New("el horario o la duracion no estan permitidos por la politica vigente")
		case 51016:
			return errors.New("el recurso no esta permitido por la politica vigente")
		case 51017:
			return errors.New("debes registrar tu RUT antes de crear reservas")
		case 51023:
			return repositories.ErrParticipantConflict
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
