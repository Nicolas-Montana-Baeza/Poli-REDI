package models

import "time"

type AvailabilityItem struct {
	ID                  int       `json:"id"`
	AvailabilityKey     string    `json:"availabilityKey"`
	UserID              int       `json:"userId"`
	ResourceID          int       `json:"resourceId"`
	StartTime           time.Time `json:"startTime"`
	DurationMinutes     int       `json:"durationMinutes"`
	Status              string    `json:"status"`
	Hour                string    `json:"hour"`
	Title               string    `json:"title"`
	Type                string    `json:"type"`
	ResourceName        string    `json:"resourceName"`
	UserFullName        string    `json:"userFullName"`
	UserEmail           string    `json:"userEmail"`
	UserRUT             string    `json:"userRut"`
	IsScheduledActivity bool      `json:"isScheduledActivity"`
	ActivityType        string    `json:"activityType,omitempty"`
}

type ScheduledActivity struct {
	ID              int
	ResourceID      int
	Title           string
	ActivityType    string
	StartTime       time.Time
	EndTime         time.Time
	ResourceName    string
	CreatedByUserID int
}
