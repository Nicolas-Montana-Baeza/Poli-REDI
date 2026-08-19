package services

import (
	"database/sql"
	"errors"
	"strings"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"

	"github.com/jackc/pgx/v5/pgconn"
)

// ============================================================================
// ERRORES DE DOMINIO
// ============================================================================
//
// Estos errores representan situaciones funcionales del módulo.
// Los handlers podrán traducirlos posteriormente a códigos HTTP sin depender
// de detalles internos de PostgreSQL.

var (
	ErrInstitutionalUnauthorized = errors.New(
		"no tienes permisos para gestionar programación institucional",
	)

	ErrInstitutionalUnitNotFound = errors.New(
		"unidad institucional no encontrada",
	)

	ErrInstitutionalUnitInactive = errors.New(
		"la unidad institucional está inactiva",
	)

	ErrInstitutionalUnitDuplicate = errors.New(
		"ya existe una unidad institucional con ese nombre o código",
	)

	ErrInstitutionalInvalidUnit = errors.New(
		"los datos de la unidad institucional no son válidos",
	)

	ErrInstitutionalInvalidMembership = errors.New(
		"los datos de la membresía institucional no son válidos",
	)

	ErrInstitutionalMembershipTargetNotFound = errors.New(
		"la unidad institucional o el usuario no existe",
	)
)

// ============================================================================
// CONSULTA DE UNIDADES DISPONIBLES PARA EL USUARIO
// ============================================================================

// GetInstitutionalUnitsForUser devuelve únicamente las unidades sobre las que
// el usuario autenticado puede actuar.
//
// Administrador global:
//
//	puede gestionar todas las unidades institucionales activas.
//
// MANAGER:
//
//	puede gestionar únicamente las unidades a las que pertenece como gestor.
//
// MEMBER:
//
//	no obtiene permisos de programación por el solo hecho de pertenecer a una
//	unidad.
func GetInstitutionalUnitsForUser(
	user models.LocalAuthUser,
) ([]models.InstitutionalUnit, error) {
	if err := validateInstitutionalActor(user); err != nil {
		return nil, err
	}

	if user.IsAdmin {
		return repositories.GetActiveInstitutionalUnits()
	}

	return repositories.GetManagedInstitutionalUnitsForUser(user.ID)
}

// ============================================================================
// CREACIÓN DE UNIDADES
// ============================================================================

// CreateInstitutionalUnit crea una nueva unidad institucional.
//
// Crear estructuras institucionales es una operación global del sistema.
// Un MANAGER administra programación de su unidad, pero no puede inventar o
// registrar nuevas carreras, departamentos o programas institucionales.
func CreateInstitutionalUnit(
	user models.LocalAuthUser,
	request models.CreateInstitutionalUnitRequest,
) (models.InstitutionalUnit, error) {
	if err := requireInstitutionalAdmin(user); err != nil {
		return models.InstitutionalUnit{}, err
	}

	normalizeInstitutionalUnitRequest(&request)

	if !validInstitutionalUnitRequest(request) {
		return models.InstitutionalUnit{}, ErrInstitutionalInvalidUnit
	}

	unit, err := repositories.CreateInstitutionalUnit(request)
	if err != nil {
		return models.InstitutionalUnit{}, mapInstitutionalUnitRepositoryError(err)
	}

	return unit, nil
}

// ============================================================================
// CONSULTA DE MEMBRESÍAS
// ============================================================================

// GetInstitutionalUnitMemberships permite consultar quién pertenece a una
// unidad.
//
// Tanto el administrador global como un MANAGER activo pueden consultar esta
// información. Modificar permisos, en cambio, seguirá siendo responsabilidad
// exclusiva del administrador global durante MVP2.
func GetInstitutionalUnitMemberships(
	user models.LocalAuthUser,
	unitID int,
) ([]models.InstitutionalUnitMembership, error) {
	if unitID <= 0 {
		return nil, ErrInstitutionalUnitNotFound
	}

	if err := EnsureInstitutionalUnitManager(user, unitID); err != nil {
		return nil, err
	}

	return repositories.GetInstitutionalUnitMemberships(unitID)
}

// ============================================================================
// ASIGNACIÓN DE MEMBRESÍAS
// ============================================================================

// AddInstitutionalUnitMembership agrega o reactiva una membresía.
//
// En MVP2 solo el administrador global puede asignar MEMBER/MANAGER.
// Esto evita que un gestor pueda escalar privilegios o delegar administración
// de su unidad sin control institucional.
func AddInstitutionalUnitMembership(
	user models.LocalAuthUser,
	unitID int,
	request models.AddInstitutionalUnitMembershipRequest,
) (models.InstitutionalUnitMembership, error) {
	if err := requireInstitutionalAdmin(user); err != nil {
		return models.InstitutionalUnitMembership{}, err
	}

	if unitID <= 0 || request.UserID <= 0 {
		return models.InstitutionalUnitMembership{},
			ErrInstitutionalInvalidMembership
	}

	request.Role = strings.ToUpper(strings.TrimSpace(request.Role))

	if !validInstitutionalMembershipRole(request.Role) {
		return models.InstitutionalUnitMembership{},
			ErrInstitutionalInvalidMembership
	}

	unit, err := repositories.GetInstitutionalUnitByID(unitID)
	if err != nil {
		return models.InstitutionalUnitMembership{},
			mapInstitutionalUnitRepositoryError(err)
	}

	// No permitimos agregar nuevos gestores o miembros a una unidad que ya fue
	// desactivada. Las relaciones históricas existentes permanecen conservadas.
	if !unit.IsActive {
		return models.InstitutionalUnitMembership{},
			ErrInstitutionalUnitInactive
	}

	membership, err :=
		repositories.AddOrReactivateInstitutionalUnitMembership(
			unitID,
			request,
		)

	if err != nil {
		return models.InstitutionalUnitMembership{},
			mapInstitutionalUnitRepositoryError(err)
	}

	return membership, nil
}

// ============================================================================
// AUTORIZACIÓN REUTILIZABLE
// ============================================================================

// EnsureInstitutionalUnitManager centraliza la regla de autorización que será
// reutilizada posteriormente por actividades, horarios y talleres.
//
// La regla es:
//
//	administrador global
//	OR
//	MANAGER activo de la unidad activa
//
// Tener role=MEMBER no entrega permisos de programación.
func EnsureInstitutionalUnitManager(
	user models.LocalAuthUser,
	unitID int,
) error {
	if err := validateInstitutionalActor(user); err != nil {
		return err
	}

	if unitID <= 0 {
		return ErrInstitutionalUnitNotFound
	}

	unit, err := repositories.GetInstitutionalUnitByID(unitID)
	if err != nil {
		return mapInstitutionalUnitRepositoryError(err)
	}

	if !unit.IsActive {
		return ErrInstitutionalUnitInactive
	}

	// Un administrador global puede actuar sobre cualquier unidad activa sin
	// necesitar una membresía artificial en cada una de ellas.
	if user.IsAdmin {
		return nil
	}

	isManager, err := repositories.IsInstitutionalUnitManager(
		unitID,
		user.ID,
	)

	if err != nil {
		return err
	}

	if !isManager {
		return ErrInstitutionalUnauthorized
	}

	return nil
}

// ============================================================================
// VALIDACIONES
// ============================================================================

func validateInstitutionalActor(
	user models.LocalAuthUser,
) error {
	if user.ID <= 0 || user.IsBlocked {
		return ErrInstitutionalUnauthorized
	}

	return nil
}

func requireInstitutionalAdmin(
	user models.LocalAuthUser,
) error {
	if err := validateInstitutionalActor(user); err != nil {
		return err
	}

	if !user.IsAdmin {
		return ErrInstitutionalUnauthorized
	}

	return nil
}

func normalizeInstitutionalUnitRequest(
	request *models.CreateInstitutionalUnitRequest,
) {
	request.Name = strings.TrimSpace(request.Name)
	request.Code = strings.ToUpper(strings.TrimSpace(request.Code))
	request.UnitType = strings.ToUpper(strings.TrimSpace(request.UnitType))
}

func validInstitutionalUnitRequest(
	request models.CreateInstitutionalUnitRequest,
) bool {
	if request.Name == "" ||
		request.Code == "" {
		return false
	}

	switch request.UnitType {
	case models.InstitutionalUnitTypeAcademicProgram,
		models.InstitutionalUnitTypePostgraduateProgram,
		models.InstitutionalUnitTypeSportsUnit,
		models.InstitutionalUnitTypeAdministrativeUnit,
		models.InstitutionalUnitTypeOther:

		return true

	default:
		return false
	}
}

func validInstitutionalMembershipRole(
	role string,
) bool {
	switch role {
	case models.InstitutionalMembershipRoleMember,
		models.InstitutionalMembershipRoleManager:

		return true

	default:
		return false
	}
}

// ============================================================================
// MAPEO DE ERRORES POSTGRESQL
// ============================================================================

// mapInstitutionalUnitRepositoryError evita filtrar mensajes técnicos de la
// base de datos hacia handlers/frontend.
//
// SQLSTATE:
//
//	23505 -> unique_violation
//	23503 -> foreign_key_violation
//	23514 -> check_violation
func mapInstitutionalUnitRepositoryError(
	err error,
) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInstitutionalUnitNotFound
	}

	var pgErr *pgconn.PgError

	if !errors.As(err, &pgErr) {
		return err
	}

	switch pgErr.Code {

	case "23505":
		return ErrInstitutionalUnitDuplicate

	case "23503":
		return ErrInstitutionalMembershipTargetNotFound

	case "23514":
		return ErrInstitutionalInvalidUnit

	default:
		return err
	}
}
