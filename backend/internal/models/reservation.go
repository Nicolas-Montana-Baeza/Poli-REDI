package models

import "time"

// Reservation status constants define the lifecycle of a reservation.
const (
	ReservationStatusPending   = "PENDING"
	ReservationStatusConfirmed = "CONFIRMED"
	ReservationStatusCancelled = "CANCELLED"
	ReservationStatusRejected  = "REJECTED"
	ReservationStatusExpired   = "EXPIRED"
)

// Reservation represents a booking made by a user for a specific resource and time.
type Reservation struct {
	ID              int       `json:"id"`
	PolicyID        int       `json:"policyId"`
	UserID          int       `json:"userId,omitempty"`
	ResourceID      int       `json:"resourceId"`
	ActivityID      *int      `json:"activityId"`
	StartTime       time.Time `json:"startTime"`
	DurationMinutes int       `json:"durationMinutes"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`

	// Campos útiles para el frontend actual
	Hour                 string     `json:"hour"`
	Title                string     `json:"title"`
	Type                 string     `json:"type"`
	ResourceName         string     `json:"resourceName"`
	UserFullName         string     `json:"userFullName,omitempty"`
	UserEmail            string     `json:"userEmail,omitempty"`
	UserRUT              string     `json:"userRut,omitempty"`
	JoinCode             string     `json:"joinCode,omitempty"`
	ParticipantCount     int        `json:"participantCount"`
	MinimumParticipants  int        `json:"minimumParticipants"`
	Capacity             *int       `json:"capacity,omitempty"`
	TargetParticipants   *int       `json:"targetParticipants,omitempty"`
	ConfirmationDeadline *time.Time `json:"confirmationDeadline,omitempty"`
	CanEditTarget        bool       `json:"canEditTarget"`
}

type ReservationProgress struct {
	ReservationID        int       `json:"reservationId"`
	Status               string    `json:"status"`
	ParticipantCount     int       `json:"participantCount"`
	MinimumParticipants  int       `json:"minimumParticipants"`
	Capacity             int       `json:"capacity"`
	IsMember             bool      `json:"isMember"`
	TargetParticipants   int       `json:"targetParticipants"`
	ConfirmationDeadline time.Time `json:"confirmationDeadline"`
	CanEditTarget        bool      `json:"canEditTarget"`
	IsOwner              bool      `json:"isOwner"`
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
	ResourceID         int    `json:"resourceId"`
	ActivityID         *int   `json:"activityId"`
	StartTime          string `json:"startTime"`
	DurationMinutes    int    `json:"durationMinutes"`
	TargetParticipants *int   `json:"targetParticipants,omitempty"`
}

type UpdateTargetParticipantsRequest struct {
	TargetParticipants int `json:"targetParticipants"`
}

type JoinCodeResponse struct {
	JoinCode string `json:"joinCode"`
}

type CancelReservationRequest struct {
	ReservationID int `json:"reservationId"`
}
