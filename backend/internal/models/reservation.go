package models

import "time"

const (
	ReservationStatusPending   = "PENDING"
	ReservationStatusConfirmed = "CONFIRMED"
	ReservationStatusCancelled = "CANCELLED"
	ReservationStatusRejected  = "REJECTED"
	ReservationStatusExpired   = "EXPIRED"
)

const CancellationReasonMinimumNotMet = "MINIMUM_NOT_MET"

const (
	GroupConditionPending  = "PENDING_MINIMUM"
	GroupConditionHealthy  = "HEALTHY"
	GroupConditionAtRisk   = "AT_RISK"
	GroupConditionInactive = "INACTIVE"
)

type Reservation struct {
	ID                 int       `json:"id"`
	PolicyID           int       `json:"policyId"`
	UserID             int       `json:"userId"`
	ResourceID         int       `json:"resourceId"`
	ActivityID         *int      `json:"activityId"`
	StartTime          time.Time `json:"startTime"`
	DurationMinutes    int       `json:"durationMinutes"`
	Status             string    `json:"status"`
	CancellationReason string    `json:"cancellationReason,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`

	// Datos utilizados por el frontend.
	Hour         string `json:"hour"`
	Title        string `json:"title"`
	Type         string `json:"type"`
	ResourceName string `json:"resourceName"`
	UserFullName string `json:"userFullName"`
	UserEmail    string `json:"userEmail"`
	UserRUT      string `json:"userRut"`

	// Datos de flujo grupal MVP2.
	//
	// JoinCode solo se entrega cuando corresponde mostrar el código
	// recién generado. La base de datos almacena únicamente su hash.
	JoinCode string `json:"joinCode,omitempty"`

	ParticipantCount    int    `json:"participantCount"`
	MinimumParticipants int    `json:"minimumParticipants"`
	Capacity            *int   `json:"capacity,omitempty"`
	GroupCondition      string `json:"groupCondition,omitempty"`
	IsGroupReservation  bool   `json:"isGroupReservation"`
}

type ReservationProgress struct {
	ReservationID       int    `json:"reservationId"`
	Status              string `json:"status"`
	GroupCondition      string `json:"groupCondition"`
	ParticipantCount    int    `json:"participantCount"`
	MinimumParticipants int    `json:"minimumParticipants"`
	Capacity            int    `json:"capacity"`
	IsMember            bool   `json:"isMember"`
	IsOwner             bool   `json:"isOwner"`
}

type ReservationParticipant struct {
	UserID      int        `json:"userId"`
	FullName    string     `json:"fullName"`
	Email       string     `json:"email"`
	RUT         string     `json:"rut"`
	IsOwner     bool       `json:"isOwner"`
	Status      string     `json:"status"`
	ConfirmedAt *time.Time `json:"confirmedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
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
