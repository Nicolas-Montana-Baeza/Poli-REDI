package services

import (
	"database/sql"
	"errors"
	"sort"
	"strconv"
	"time"

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

func GetReservations() ([]models.Reservation, error) {
	if err := runReservationHousekeeping(); err != nil {
		return nil, err
	}

	return repositories.GetAllReservations()
}

func GetAvailabilityItems(from, to time.Time, userID int, isAdmin bool) ([]models.AvailabilityItem, error) {
	if err := runReservationHousekeeping(); err != nil {
		return nil, err
	}

	reservations, err := repositories.GetActiveReservationsForAvailability(from, to, userID, isAdmin)

	if err != nil {
		return nil, err
	}

	blocks, err := repositories.GetAvailabilityBlocks(from, to)

	if err != nil {
		return nil, err
	}

	scheduledActivities, err :=
		repositories.GetScheduledInstitutionalActivitiesForAvailability(
			from,
			to,
		)

	if err != nil {
		return nil, err
	}

	items := make(
		[]models.AvailabilityItem,
		0,
		len(reservations)+len(blocks)+len(scheduledActivities),
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
	// ========================================================================
	// PROGRAMACIÓN INSTITUCIONAL
	// ========================================================================
	//
	// Una actividad institucional aparece como ocupación del recurso, pero no
	// como reserva perteneciente al usuario que la creó.
	//
	// Por eso UserID queda en cero: created_by representa trazabilidad
	// administrativa, no propiedad personal de la ocupación.

	for _, activity := range scheduledActivities {
		duration := int(
			activity.EndTime.
				Sub(activity.StartTime).
				Minutes(),
		)

		items = append(
			items,
			models.AvailabilityItem{
				ID: activity.ID,

				// Una actividad WEEKLY puede generar múltiples ocurrencias.
				// El ID de actividad por sí solo no sería una clave única para
				// el calendario.
				AvailabilityKey: "scheduled-" +
					strconv.Itoa(activity.ID) +
					"-" +
					strconv.FormatInt(
						activity.StartTime.Unix(),
						10,
					),

				ResourceID: activity.ResourceID,

				StartTime: activity.StartTime,

				DurationMinutes: duration,

				Status: "CONFIRMED",

				Hour: activity.StartTime.
					Format("15:04"),

				Title: activity.Title,

				Type: "scheduled",

				ResourceName: activity.ResourceName,

				IsScheduledActivity: true,

				ActivityType: activity.ActivityType,
			},
		)
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

	if err := runReservationHousekeeping(); err != nil {
		return nil, err
	}

	return repositories.GetReservationsByUserID(userID)
}

func GetReservationDetail(id int, requestedBy models.LocalAuthUser) (models.Reservation, error) {
	if err := runReservationHousekeeping(); err != nil {
		return models.Reservation{}, err
	}

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

	if err := runReservationHousekeepingAt(now); err != nil {
		return models.Reservation{}, err
	}

	resource, err := repositories.GetResourceByID(reservation.ResourceID)

	if err != nil {
		return models.Reservation{}, err
	}

	if resource.ReservationMode == "OPEN_USE" {
		reservation.ActivityID = nil
	}

	// La disponibilidad institucional ya se protege en PostgreSQL.
	// PG16_0006 impide que una reserva nueva se solape con una actividad
	// institucional SCHEDULED, incluidos talleres institucionales.
	//
	// No se mantiene una segunda validacion basada en las tablas legacy
	// workshops/workshop_enrollments de SQL Server.

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
		case "P1011":
			return errors.New("el recurso tiene programación institucional en ese horario")
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

	if err := runReservationHousekeepingAt(now); err != nil {
		return models.Reservation{}, err
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

func runReservationHousekeeping() error {
	return runReservationHousekeepingAt(
		businessclock.Now(),
	)
}

func runReservationHousekeepingAt(now time.Time) error {
	_, err := repositories.ExpirePendingGroupReservations(now)
	return err
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
