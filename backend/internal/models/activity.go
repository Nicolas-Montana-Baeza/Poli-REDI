package models

// Activity represents a specific sport or activity that can be practiced in a resource (e.g., Football, Basketball).
// It acts as a catalog entity to standardize reservation purposes.
type Activity struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"isActive"`
}
