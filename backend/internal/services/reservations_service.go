package services

import (
	"errors"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"

	"github.com/jackc/pgx/v5/pgconn"
)

func GetReservations() ([]models.Reservation, error) {
	return repositories.GetAllReservations()
}

func CreateReservation(reservation models.Reservation) (models.Reservation, error) {
	if reservation.UserID == 0 {
		return models.Reservation{}, errors.New("userId es obligatorio")
	}

	if reservation.ResourceID == 0 {
		return models.Reservation{}, errors.New("resourceId es obligatorio")
	}

	if reservation.StartTime.IsZero() {
		return models.Reservation{}, errors.New("startTime es obligatorio")
	}

	if reservation.DurationMinutes <= 0 {
		return models.Reservation{}, errors.New("durationMinutes debe ser mayor a 0")
	}

	if reservation.Status == "" {
		reservation.Status = "CONFIRMED"
	}

	createdReservation, err :=
		repositories.AddReservation(reservation)

	if err != nil {
		return models.Reservation{}, mapDatabaseReservationError(err)
	}

	return createdReservation, nil
}

func mapDatabaseReservationError(err error) error {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {

		case "23P01":
			return errors.New("el recurso ya está reservado en ese horario")

		case "23503":
			return errors.New("usuario, recurso o actividad no existe")

		case "23514":
			return errors.New("los datos de la reserva no cumplen las restricciones")

		default:
			return errors.New(pgErr.Message)
		}
	}

	return err
}
