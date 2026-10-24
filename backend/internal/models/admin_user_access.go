package models

import "time"

// AdminUserAccessRequest pide un cambio de estado con motivo auditable.
type AdminUserAccessRequest struct {
	IsBlocked *bool  `json:"isBlocked"`
	Reason    string `json:"reason"`
}

type UserAccessAudit struct {
	ID                int64     `json:"id"`
	TargetUserID      int       `json:"targetUserId"`
	ActorUserID       int       `json:"actorUserId"`
	ActorName         string    `json:"actorName,omitempty"`
	PreviousIsBlocked bool      `json:"previousIsBlocked"`
	NewIsBlocked      bool      `json:"newIsBlocked"`
	Reason            string    `json:"reason"`
	ChangedAt         time.Time `json:"changedAt"`
}

type AdminUserAccessResult struct {
	User  LocalAuthUser   `json:"user"`
	Audit UserAccessAudit `json:"audit"`
}
