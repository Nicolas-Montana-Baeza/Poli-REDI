package services

import (
	"database/sql"
	"errors"
	"time"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"

	mssql "github.com/microsoft/go-mssqldb"
)

func GetReservations() ([]models.Reservation, error) {
	return repositories.GetAllReservations()
}

func GetMyReservations(userID int) ([]models.Reservation, error) {
	if userID <= 0 {
		return nil, errors.New("usuario autenticado es obligatorio")
	}

	return repositories.GetReservationsByUserID(userID)
}

func CreateReservation(reservation models.Reservation) (models.Reservation, error) {
	if reservation.UserID == 0 {
		return models.Reservation{}, errors.New("no se pudo identificar al usuario autenticado")
	}

	if reservation.ResourceID == 0 {
		return models.Reservation{}, errors.New("selecciona una instalaci\u00f3n")
	}

	if reservation.StartTime.IsZero() {
		return models.Reservation{}, errors.New("selecciona una fecha y hora de inicio")
	}

	if reservation.DurationMinutes <= 0 {
		return models.Reservation{}, errors.New("la duraci\u00f3n debe ser mayor a 0")
	}

	if reservation.StartTime.Before(time.Now()) {
		return models.Reservation{}, errors.New("no puedes crear reservas en el pasado")
	}

	if reservation.Status == "" {
		reservation.Status = "CONFIRMED"
	}

	createdReservation, err := repositories.AddReservation(reservation)

	if err != nil {
		return models.Reservation{}, mapDatabaseReservationError(err)
	}

	return createdReservation, nil
}

func mapDatabaseReservationError(err error) error {
	var sqlErr mssql.Error

	if errors.As(err, &sqlErr) {
		switch sqlErr.Number {
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
	if reservationID <= 0 {
		return models.Reservation{}, errors.New("no se pudo identificar la reserva")
	}

	if requestedByUser.ID <= 0 {
		return models.Reservation{}, errors.New("usuario autenticado es obligatorio")
	}

	ownerID, status, err := repositories.GetReservationOwnerAndStatus(reservationID)

	if errors.Is(err, sql.ErrNoRows) {
		return models.Reservation{}, errors.New("reserva no encontrada")
	}

	if err != nil {
		return models.Reservation{}, err
	}

	if !requestedByUser.IsAdmin && ownerID != requestedByUser.ID {
		return models.Reservation{}, errors.New("no tienes permisos para cancelar esta reserva")
	}

	if status == "CANCELLED" {
		return models.Reservation{}, errors.New("la reserva ya est\u00e1 cancelada")
	}

	cancelledReservation, err := repositories.CancelReservation(reservationID)

	if err != nil {
		return models.Reservation{}, err
	}

	return cancelledReservation, nil
}
