package models

import "time"

type AdminUserRoleRequest struct {
	IsAdmin *bool  `json:"isAdmin"`
	Reason  string `json:"reason"`
}

type UserRoleAudit struct {
	ID              int64     `json:"id"`
	TargetUserID    int       `json:"targetUserId"`
	ActorUserID     int       `json:"actorUserId"`
	ActorName       string    `json:"actorName,omitempty"`
	PreviousIsAdmin bool      `json:"previousIsAdmin"`
	NewIsAdmin      bool      `json:"newIsAdmin"`
	Reason          string    `json:"reason"`
	ChangedAt       time.Time `json:"changedAt"`
}

type AdminUserRoleResult struct {
	User  LocalAuthUser `json:"user"`
	Audit UserRoleAudit `json:"audit"`
}
