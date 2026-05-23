package models

type Reservation struct {
	ID         int    `json:"id"`
	ResourceID int    `json:"resourceId"`
	Hour       string `json:"hour"`
	Title      string `json:"title"`
	Type       string `json:"type"`
}
