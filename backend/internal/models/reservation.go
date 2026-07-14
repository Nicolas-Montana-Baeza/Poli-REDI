package models

import "time"

const (
	ReservationStatusPending   = "PENDING"
	ReservationStatusConfirmed = "CONFIRMED"
	ReservationStatusCancelled = "CANCELLED"
	ReservationStatusRejected  = "REJECTED"
	ReservationStatusExpired   = "EXPIRED"
)

type Reservation struct {
	ID              int       `json:"id"`
	UserID          int       `json:"userId"`
	ResourceID      int       `json:"resourceId"`
	ActivityID      *int      `json:"activityId"`
	StartTime       time.Time `json:"startTime"`
	DurationMinutes int       `json:"durationMinutes"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`

	// Campos útiles para el frontend actual
	Hour         string `json:"hour"`
	Title        string `json:"title"`
	Type         string `json:"type"`
	ResourceName string `json:"resourceName"`
	UserFullName string `json:"userFullName"`
	UserEmail    string `json:"userEmail"`
	UserRUT      string `json:"userRut"`
}

type CreateReservationRequest struct {
	ResourceID      int    `json:"resourceId"`
	ActivityID      *int   `json:"activityId"`
	StartTime       string `json:"startTime"`
	DurationMinutes int    `json:"durationMinutes"`
}

type CancelReservationRequest struct {
	ReservationID int `json:"reservationId"`
}
