package services

import (
	"errors"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
)

// ============================================================================
// ERRORES DE DOMINIO
// ============================================================================

var (
	ErrInstitutionalWorkshopUnauthorized = errors.New(
		"usuario autenticado obligatorio para talleres",
	)

	ErrInstitutionalWorkshopInvalidID = errors.New(
		"no se pudo identificar el taller",
	)
)

// ============================================================================
// CONSULTA
// ============================================================================

func GetInstitutionalWorkshops(
	user models.LocalAuthUser,
) ([]models.InstitutionalActivity, error) {
	if user.ID <= 0 ||
		user.IsBlocked {

		return nil,
			ErrInstitutionalWorkshopUnauthorized
	}

	return repositories.
		GetScheduledInstitutionalWorkshopsForUser(
			user.ID,
		)
}

func GetInstitutionalWorkshop(
	workshopID int,
	user models.LocalAuthUser,
) (models.InstitutionalActivity, error) {
	if user.ID <= 0 ||
		user.IsBlocked {

		return models.InstitutionalActivity{},
			ErrInstitutionalWorkshopUnauthorized
	}

	if workshopID <= 0 {
		return models.InstitutionalActivity{},
			ErrInstitutionalWorkshopInvalidID
	}

	return repositories.
		GetScheduledInstitutionalWorkshopForUser(
			workshopID,
			user.ID,
		)
}

// ============================================================================
// INSCRIPCIÓN
// ============================================================================

func EnrollInInstitutionalWorkshop(
	workshopID int,
	user models.LocalAuthUser,
) (models.InstitutionalActivity, error) {
	if user.ID <= 0 ||
		user.IsBlocked {

		return models.InstitutionalActivity{},
			ErrInstitutionalWorkshopUnauthorized
	}

	if workshopID <= 0 {
		return models.InstitutionalActivity{},
			ErrInstitutionalWorkshopInvalidID
	}

	return repositories.
		EnrollUserInInstitutionalWorkshop(
			workshopID,
			user.ID,
		)
}

// ============================================================================
// RETIRO
// ============================================================================

func LeaveInstitutionalWorkshop(
	workshopID int,
	user models.LocalAuthUser,
) (models.InstitutionalActivity, error) {
	if user.ID <= 0 ||
		user.IsBlocked {

		return models.InstitutionalActivity{},
			ErrInstitutionalWorkshopUnauthorized
	}

	if workshopID <= 0 {
		return models.InstitutionalActivity{},
			ErrInstitutionalWorkshopInvalidID
	}

	return repositories.
		CancelUserInstitutionalWorkshopEnrollment(
			workshopID,
			user.ID,
		)
}
