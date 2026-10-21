package services

import (
	"database/sql"
	"errors"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
)

var ErrInstitutionalActivityNotFound = errors.New("actividad institucional no encontrada")

func institutionalActivityForManagement(user models.LocalAuthUser, id int) (models.InstitutionalActivity, error) {
	if id <= 0 {
		return models.InstitutionalActivity{}, ErrInstitutionalActivityNotFound
	}
	activity, err := repositories.GetInstitutionalActivityByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		return models.InstitutionalActivity{}, ErrInstitutionalActivityNotFound
	}
	if err != nil {
		return models.InstitutionalActivity{}, err
	}
	if err = EnsureInstitutionalUnitManager(user, activity.UnitID); err != nil {
		return models.InstitutionalActivity{}, err
	}
	return activity, nil
}

func CancelInstitutionalActivity(user models.LocalAuthUser, id int) (models.InstitutionalActivity, error) {
	if _, err := institutionalActivityForManagement(user, id); err != nil {
		return models.InstitutionalActivity{}, err
	}
	return repositories.CancelInstitutionalActivity(id, user.ID)
}

func UpdateInstitutionalActivity(user models.LocalAuthUser, id int, request models.CreateInstitutionalActivityRequest) (models.InstitutionalActivity, error) {
	activity, err := institutionalActivityForManagement(user, id)
	if err != nil {
		return models.InstitutionalActivity{}, err
	}
	if request.UnitID != activity.UnitID || request.ResourceID != activity.ResourceID {
		return models.InstitutionalActivity{}, repositories.ErrInstitutionalActivityResourceChange
	}
	if err = normalizeInstitutionalActivityRequest(&request); err != nil {
		return models.InstitutionalActivity{}, err
	}
	if err = validateInstitutionalActivityRequest(request); err != nil {
		return models.InstitutionalActivity{}, err
	}
	if err = validateInstitutionalScheduleSelfOverlap(request.Schedules); err != nil {
		return models.InstitutionalActivity{}, err
	}
	if err := ensureInstitutionalOperatingHours(request.Schedules); err != nil {
		return models.InstitutionalActivity{}, err
	}
	result, err := repositories.UpdateInstitutionalActivityWithSchedules(id, user.ID, request)
	if err != nil {
		return models.InstitutionalActivity{}, mapInstitutionalActivityRepositoryError(err)
	}
	return result, nil
}
