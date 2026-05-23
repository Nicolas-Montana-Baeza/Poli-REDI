package repositories

import "poli-redi-api/internal/models"

var reservations = []models.Reservation{
	{
		ID:         1,
		ResourceID: 1,
		Hour:       "18:00",
		Title:      "Entrenamiento",
		Type:       "normal",
	},
	{
		ID:         2,
		ResourceID: 2,
		Hour:       "20:00",
		Title:      "Campeonato",
		Type:       "priority",
	},
	{
		ID:         3,
		ResourceID: 4,
		Hour:       "17:00",
		Title:      "Mantención",
		Type:       "normal",
	},
}

func GetAllReservations() []models.Reservation {
	return reservations
}

func AddReservation(reservation models.Reservation) models.Reservation {
	reservation.ID = len(reservations) + 1

	reservations = append(
		reservations,
		reservation,
	)

	return reservation
}
