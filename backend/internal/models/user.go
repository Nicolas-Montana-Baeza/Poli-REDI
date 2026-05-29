package models

type LocalAuthUser struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"fullName"`
	IsAdmin   bool   `json:"isAdmin"`
	IsBlocked bool   `json:"isBlocked"`
	OID       string `json:"oid"`
	TenantID  string `json:"tenantId"`
}
