package models

import "time"

type Workshop struct {
	ID             int       `json:"id"`
	ResourceID     int       `json:"resourceId"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Location       string    `json:"location"`
	InstructorName string    `json:"instructorName"`
	DayText        string    `json:"dayText"`
	ScheduleText   string    `json:"scheduleText"`
	Capacity       int       `json:"capacity"`
	EnrolledCount  int       `json:"enrolledCount"`
	IsActive       bool      `json:"isActive"`
	IsEnrolled     bool      `json:"isEnrolled"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type WorkshopEnrollment struct {
	ID         int       `json:"id"`
	WorkshopID int       `json:"workshopId"`
	UserID     int       `json:"userId"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
}
