package services

import (
	"errors"
	"strings"
	"unicode/utf8"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
)

var ErrUserRoleInvalidReason = errors.New("el motivo debe contener entre 5 y 500 caracteres")
var ErrUserRoleInvalidRequest = errors.New("debes indicar el rol administrativo")

func validateUserRoleChange(actor models.LocalAuthUser, targetID int, req models.AdminUserRoleRequest) (bool, string, error) {
	if actor.ID <= 0 || !actor.IsAdmin || actor.IsBlocked {
		return false, "", repositories.ErrUserRoleUnauthorized
	}
	if targetID <= 0 {
		return false, "", repositories.ErrUserRoleNotFound
	}
	if targetID == actor.ID {
		return false, "", repositories.ErrUserRoleSelf
	}
	if req.IsAdmin == nil {
		return false, "", ErrUserRoleInvalidRequest
	}
	reason := strings.TrimSpace(req.Reason)
	if length := utf8.RuneCountInString(reason); length < 5 || length > 500 {
		return false, "", ErrUserRoleInvalidReason
	}
	return *req.IsAdmin, reason, nil
}

func SetAdminUserRole(actor models.LocalAuthUser, targetID int, req models.AdminUserRoleRequest) (models.AdminUserRoleResult, error) {
	grant, reason, err := validateUserRoleChange(actor, targetID, req)
	if err != nil {
		return models.AdminUserRoleResult{}, err
	}
	return repositories.ChangeAdminUserRole(actor.ID, targetID, grant, reason)
}

func GetAdminUserRoleHistory(actor models.LocalAuthUser, targetID int) ([]models.UserRoleAudit, error) {
	if actor.ID <= 0 || !actor.IsAdmin || actor.IsBlocked {
		return nil, repositories.ErrUserRoleUnauthorized
	}
	if targetID <= 0 {
		return nil, repositories.ErrUserRoleNotFound
	}
	return repositories.GetUserRoleAudit(targetID)
}
