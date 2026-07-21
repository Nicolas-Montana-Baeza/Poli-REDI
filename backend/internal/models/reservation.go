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
	PolicyID        int       `json:"policyId"`
	UserID          int       `json:"userId"`
	ResourceID      int       `json:"resourceId"`
	ActivityID      *int      `json:"activityId"`
	StartTime       time.Time `json:"startTime"`
	DurationMinutes int       `json:"durationMinutes"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`

	// Campos útiles para el frontend actual
	Hour                string `json:"hour"`
	Title               string `json:"title"`
	Type                string `json:"type"`
	ResourceName        string `json:"resourceName"`
	UserFullName        string `json:"userFullName"`
	UserEmail           string `json:"userEmail"`
	UserRUT             string `json:"userRut"`
	JoinCode            string `json:"joinCode,omitempty"`
	ParticipantCount    int    `json:"participantCount"`
	MinimumParticipants int    `json:"minimumParticipants"`
	Capacity            *int   `json:"capacity,omitempty"`
}

type ReservationProgress struct {
	ReservationID       int    `json:"reservationId"`
	Status              string `json:"status"`
	ParticipantCount    int    `json:"participantCount"`
	MinimumParticipants int    `json:"minimumParticipants"`
	Capacity            int    `json:"capacity"`
	IsMember            bool   `json:"isMember"`
}
type ReservationParticipant struct {
	UserID   int    `json:"userId"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	RUT      string `json:"rut"`
	IsOwner  bool   `json:"isOwner"`
	Status   string `json:"status"`
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
