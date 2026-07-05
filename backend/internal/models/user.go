package models

type LocalAuthUser struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"fullName"`
	RUT       string `json:"rut"`
	IsAdmin   bool   `json:"isAdmin"`
	IsBlocked bool   `json:"isBlocked"`
	OID       string `json:"oid"`
	TenantID  string `json:"tenantId"`
}

type UpdateRUTRequest struct {
	RUT string `json:"rut"`
}
