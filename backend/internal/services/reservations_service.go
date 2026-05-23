package services

import (
	"errors"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
)

func GetReservations() []models.Reservation {
	return repositories.GetAllReservations()
}

func CreateReservation(reservation models.Reservation) (models.Reservation, error) {
	if reservation.ResourceID == 0 {
		return models.Reservation{}, errors.New("resourceId es obligatorio")
	}

	if reservation.Hour == "" {
		return models.Reservation{}, errors.New("hour es obligatorio")
	}

	if reservation.Title == "" {
		reservation.Title = "Reserva"
	}

	if reservation.Type == "" {
		reservation.Type = "normal"
	}

	existingReservations := repositories.GetAllReservations()

	for _, existing := range existingReservations {
		if existing.ResourceID == reservation.ResourceID &&
			existing.Hour == reservation.Hour {
			return models.Reservation{}, errors.New("el horario ya está reservado para este recurso")
		}
	}

	createdReservation := repositories.AddReservation(reservation)

	return createdReservation, nil
}
