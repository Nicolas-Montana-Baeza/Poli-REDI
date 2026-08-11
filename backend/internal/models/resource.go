package models

// Resource represents a bookable physical space or asset in the sports center (e.g., Court 1, Multipurpose Room).
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
