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

func EnrollInWorkshop(
	workshopID int,
	user models.LocalAuthUser,
) (models.Workshop, error) {
	if workshopID <= 0 {
		return models.Workshop{}, errors.New("no se pudo identificar el taller")
	}

	if user.ID <= 0 {
		return models.Workshop{}, errors.New("usuario autenticado es obligatorio")
	}

	workshop, err := repositories.EnrollUserInWorkshop(
		workshopID,
		user.ID,
	)

	if err != nil {
		return models.Workshop{}, mapWorkshopError(err)
	}

	return workshop, nil
}

func mapWorkshopError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("taller no encontrado o no disponible")
	}

	var sqlErr mssql.Error

	if errors.As(err, &sqlErr) {
		switch sqlErr.Number {
		case 2601, 2627:
			return errors.New("ya estas inscrito en este taller")
		case 547:
			return errors.New("taller o usuario no existe")
		default:
			return errors.New(sqlErr.Message)
		}
	}

	return err
}
