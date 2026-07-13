package models

type Resource struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	ReservationMode string `json:"reservationMode"`
	ImageURL        string `json:"imageUrl"`
	Capacity        *int   `json:"capacity"`
	IsActive        bool   `json:"isActive"`
	Status          string `json:"status"`
}
