package models

type Activity struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"isActive"`
}

type CreateActivityRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
