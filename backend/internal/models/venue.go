package models

type Venue struct {
	ID                  int      `json:"id"`
	Name                string   `json:"name"`
	AddressLine         *string  `json:"addressLine"`
	Commune             *string  `json:"commune"`
	City                *string  `json:"city"`
	Region              *string  `json:"region"`
	Country             string   `json:"country"`
	Latitude            *float64 `json:"latitude"`
	Longitude           *float64 `json:"longitude"`
	IsActive            bool     `json:"isActive"`
	ResourceCount       int64    `json:"resourceCount"`
	ActiveResourceCount int64    `json:"activeResourceCount"`
}
