package models

import "time"

type Notification struct {
	ID            int       `json:"id"`
	UserID        int       `json:"userId"`
	ReservationID *int      `json:"reservationId"`
	Title         string    `json:"title"`
	Message       string    `json:"message"`
	Type          string    `json:"type"`
	IsRead        bool      `json:"isRead"`
	CreatedAt     time.Time `json:"createdAt"`
}
