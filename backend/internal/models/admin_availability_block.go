package models

import "time"

// AdminAvailabilityBlock expone los detalles internos solo en rutas admin.
// La disponibilidad publica continua utilizando el DTO opaco AvailabilityItem.
type AdminAvailabilityBlock struct {
	ID              int       `json:"id"`
	ResourceID      int       `json:"resourceId"`
	ResourceName    string    `json:"resourceName"`
	CreatedByUserID int       `json:"createdByUserId"`
	BlockType       string    `json:"blockType"`
	Reason          string    `json:"reason"`
	StartTime       time.Time `json:"startTime"`
	EndTime         time.Time `json:"endTime"`
	IsActive        bool      `json:"isActive"`
	CreatedAt       time.Time `json:"createdAt"`
}
