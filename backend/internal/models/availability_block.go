package models

import "time"

type AvailabilityBlock struct {
	ID           int
	ResourceID   int
	BlockType    string
	Reason       string
	StartTime    time.Time
	EndTime      time.Time
	ResourceName string
}
