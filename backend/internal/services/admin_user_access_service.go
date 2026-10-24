package services

import (
	"errors"
	"strings"
	"unicode/utf8"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
)

var ErrUserAccessInvalidReason = errors.New("el motivo debe contener entre 5 y 500 caracteres")
var ErrUserAccessInvalidRequest = errors.New("debes indicar el estado de bloqueo")

func validateUserAccessChange(actor models.LocalAuthUser, targetID int, request models.AdminUserAccessRequest) (bool, string, error) {
	if actor.ID <= 0 || !actor.IsAdmin || actor.IsBlocked {
		return false, "", repositories.ErrUserAccessUnauthorized
	}
	if targetID <= 0 {
		return false, "", repositories.ErrUserAccessNotFound
	}
	if targetID == actor.ID {
		return false, "", repositories.ErrUserAccessSelf
	}
	if request.IsBlocked == nil {
		return false, "", ErrUserAccessInvalidRequest
	}
	reason := strings.TrimSpace(request.Reason)
	count := utf8.RuneCountInString(reason)
	if count < 5 || count > 500 {
		return false, "", ErrUserAccessInvalidReason
	}
	return *request.IsBlocked, reason, nil
}

func SetAdminUserAccess(actor models.LocalAuthUser, targetID int, req models.AdminUserAccessRequest) (models.AdminUserAccessResult, error) {
	blocked, reason, err := validateUserAccessChange(actor, targetID, req)
	if err != nil {
		return models.AdminUserAccessResult{}, err
	}
	return repositories.ChangeUserAccess(actor.ID, targetID, blocked, reason)
}

func GetAdminUserAccessHistory(actor models.LocalAuthUser, targetID int) ([]models.UserAccessAudit, error) {
	if actor.ID <= 0 || !actor.IsAdmin || actor.IsBlocked {
		return nil, repositories.ErrUserAccessUnauthorized
	}
	if targetID <= 0 {
		return nil, repositories.ErrUserAccessNotFound
	}
	return repositories.GetUserAccessAudit(targetID)
}
