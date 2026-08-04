package services

import (
	"database/sql"
	"errors"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"

	mssql "github.com/microsoft/go-mssqldb"
)

func GetWorkshops(userID int) ([]models.Workshop, error) {
	if userID <= 0 {
		return nil, errors.New("usuario autenticado es obligatorio")
	}

	return repositories.GetActiveWorkshopsForUser(userID)
}

func GetMyWorkshopEnrollments(
	userID int,
) ([]models.WorkshopEnrollmentHistory, error) {
	if userID <= 0 {
		return nil, errors.New("usuario autenticado es obligatorio")
	}
	return repositories.GetWorkshopEnrollmentsForUser(userID)
}

func EnrollInWorkshop(
	workshopID int,
	user models.LocalAuthUser,
) (models.Workshop, bool, error) {
	if workshopID <= 0 {
		return models.Workshop{}, false, errors.New("no se pudo identificar el taller")
	}

	if user.ID <= 0 {
		return models.Workshop{}, false, errors.New("usuario autenticado es obligatorio")
	}

	workshop, created, err := repositories.EnrollUserInWorkshop(
		workshopID,
		user.ID,
	)

	if err != nil {
		return models.Workshop{}, false, mapWorkshopError(err)
	}

	return workshop, created, nil
}

func WithdrawFromWorkshop(
	workshopID int,
	user models.LocalAuthUser,
) (models.WorkshopEnrollmentChange, error) {
	if workshopID <= 0 {
		return models.WorkshopEnrollmentChange{}, errors.New("no se pudo identificar el taller")
	}

	if user.ID <= 0 {
		return models.WorkshopEnrollmentChange{}, errors.New("usuario autenticado es obligatorio")
	}

	change, err := repositories.WithdrawUserFromWorkshop(workshopID, user.ID)
	if err != nil {
		return models.WorkshopEnrollmentChange{}, mapWorkshopError(err)
	}

	return change, nil
}

func mapWorkshopError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return repositories.ErrWorkshopNotFound
	}

	var sqlErr mssql.Error

	if errors.As(err, &sqlErr) {
		switch sqlErr.Number {
		case 2601, 2627:
			return repositories.ErrWorkshopInternal
		case 547:
			return repositories.ErrWorkshopNotFound
		case 51300:
			return repositories.ErrWorkshopScheduleInvalid
		case 51301:
			return repositories.ErrWorkshopCapacity
		case 1205:
			return repositories.ErrWorkshopInternal
		default:
			return repositories.ErrWorkshopInternal
		}
	}

	var conflict *repositories.WorkshopScheduleConflictError
	if errors.As(err, &conflict) ||
		errors.Is(err, repositories.ErrWorkshopNotFound) ||
		errors.Is(err, repositories.ErrWorkshopCapacity) ||
		errors.Is(err, repositories.ErrWorkshopScheduleInvalid) ||
		errors.Is(err, repositories.ErrWorkshopEnrollmentClosed) {
		return err
	}
	return repositories.ErrWorkshopInternal
}
