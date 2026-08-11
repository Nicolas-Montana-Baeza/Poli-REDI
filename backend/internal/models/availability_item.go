package models

import "time"

// Availability kinds define the origin or type of a block in the calendar.
const (
	AvailabilityKindReservation      = "RESERVATION"
	AvailabilityKindGroupReservation = "GROUP_RESERVATION"
	AvailabilityKindScheduled        = "SCHEDULED_ACTIVITY"
	AvailabilityKindWorkshop         = "WORKSHOP"
	AvailabilityKindBlock            = "AVAILABILITY_BLOCK"
)

// AvailabilityItem represents a unified time block in the availability calendar,
// which can be a personal reservation, a group reservation, a workshop, or an institutional activity.
type AvailabilityItem struct {
	ID                    int        `json:"id"`
	AvailabilityKey       string     `json:"availabilityKey"`
	AvailabilityKind      string     `json:"availabilityKind"`
	UserID                int        `json:"userId,omitempty"`
	ResourceID            int        `json:"resourceId"`
	StartTime             time.Time  `json:"startTime"`
	DurationMinutes       int        `json:"durationMinutes"`
	Status                string     `json:"status"`
	Hour                  string     `json:"hour"`
	Title                 string     `json:"title"`
	Type                  string     `json:"type"`
	ItemType              string     `json:"itemType,omitempty"`
	ResourceName          string     `json:"resourceName"`
	UserFullName          string     `json:"userFullName,omitempty"`
	UserEmail             string     `json:"userEmail,omitempty"`
	UserRUT               string     `json:"userRut,omitempty"`
	IsScheduledActivity   bool       `json:"isScheduledActivity"`
	ActivityType          string     `json:"activityType,omitempty"`
	ParticipantCount      int        `json:"participantCount,omitempty"`
	MinimumParticipants   int        `json:"minimumParticipants,omitempty"`
	TargetParticipants    *int       `json:"targetParticipants,omitempty"`
	Capacity              *int       `json:"capacity,omitempty"`
	ConfirmationDeadline  *time.Time `json:"confirmationDeadline,omitempty"`
	CanEditTarget         bool       `json:"canEditTarget"`
	BlocksResource        bool       `json:"blocksResource"`
	IsCurrentUserConflict bool       `json:"isCurrentUserConflict"`
}

type ScheduledActivity struct {
	ID              int
	ResourceID      int
	Title           string
	ActivityType    string
	StartTime       time.Time
	EndTime         time.Time
	ResourceName    string
	ReservationMode string
	CreatedByUserID int
}

type AvailabilityReservation struct {
	Reservation
	ReservationMode       string
	IsCurrentUserConflict bool
}

type AvailabilityBlock struct {
	ID              int
	ResourceID      int
	CreatedByUserID int
	BlockType       string
	Reason          string
	StartTime       time.Time
	EndTime         time.Time
	ResourceName    string
}

type WorkshopAvailabilityOccurrence struct {
	ID              int
	WorkshopID      int
	ResourceID      int
	Title           string
	StartTime       time.Time
	EndTime         time.Time
	ResourceName    string
	ReservationMode string
}
