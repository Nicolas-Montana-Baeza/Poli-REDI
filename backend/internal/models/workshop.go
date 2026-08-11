package models

import "time"

// Workshop represents a recurring sports class or workshop with a defined capacity.
// Users can enroll in it if there is availability.
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

type WorkshopEnrollmentHistory struct {
	ID             int       `json:"id"`
	WorkshopID     int       `json:"workshopId"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Location       string    `json:"location"`
	InstructorName string    `json:"instructorName"`
	DayText        string    `json:"dayText"`
	ScheduleText   string    `json:"scheduleText"`
	Status         string    `json:"status"`
	IsActive       bool      `json:"isActive"`
	EnrolledAt     time.Time `json:"enrolledAt"`
}

type WorkshopEnrollmentChange struct {
	WorkshopID    int  `json:"workshopId"`
	IsEnrolled    bool `json:"isEnrolled"`
	EnrolledCount int  `json:"enrolledCount"`
	Changed       bool `json:"changed"`
}
