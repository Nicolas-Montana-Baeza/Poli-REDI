package repositories

import (
	"context"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

// ============================================================================
// UNIDADES INSTITUCIONALES
// ============================================================================

// GetActiveInstitutionalUnits devuelve las unidades que actualmente pueden
// participar en la programación institucional.
//
// Las unidades inactivas se conservan en la base de datos para mantener
// trazabilidad histórica, pero no deben ofrecerse para crear programación nueva.
func GetActiveInstitutionalUnits() ([]models.InstitutionalUnit, error) {
	rows, err := database.DB.QueryContext(
		context.Background(),
		`
		SELECT
			id,
			name,
			code,
			unit_type,
			is_active,
			created_at,
			updated_at
		FROM institutional_units
		WHERE is_active = true
		ORDER BY name ASC;
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	units := []models.InstitutionalUnit{}

	for rows.Next() {
		var unit models.InstitutionalUnit

		if err := rows.Scan(
			&unit.ID,
			&unit.Name,
			&unit.Code,
			&unit.UnitType,
			&unit.IsActive,
			&unit.CreatedAt,
			&unit.UpdatedAt,
		); err != nil {
			return nil, err
		}

		units = append(units, unit)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return units, nil
}

// GetInstitutionalUnitByID obtiene una unidad incluso si actualmente está
// inactiva. Esto permite consultar información histórica sin perder la
// referencia de actividades creadas anteriormente.
func GetInstitutionalUnitByID(
	unitID int,
) (models.InstitutionalUnit, error) {
	var unit models.InstitutionalUnit

	err := database.DB.QueryRowContext(
		context.Background(),
		`
		SELECT
			id,
			name,
			code,
			unit_type,
			is_active,
			created_at,
			updated_at
		FROM institutional_units
		WHERE id = $1;
		`,
		unitID,
	).Scan(
		&unit.ID,
		&unit.Name,
		&unit.Code,
		&unit.UnitType,
		&unit.IsActive,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)

	if err != nil {
		return models.InstitutionalUnit{}, err
	}

	return unit, nil
}

// CreateInstitutionalUnit crea una nueva unidad institucional.
//
// La unicidad case-insensitive de name y code está protegida por PostgreSQL.
// El service traducirá esos errores técnicos a mensajes de dominio.
func CreateInstitutionalUnit(
	request models.CreateInstitutionalUnitRequest,
) (models.InstitutionalUnit, error) {
	var unit models.InstitutionalUnit

	err := database.DB.QueryRowContext(
		context.Background(),
		`
		INSERT INTO institutional_units (
			name,
			code,
			unit_type
		)
		VALUES (
			$1,
			$2,
			$3
		)
		RETURNING
			id,
			name,
			code,
			unit_type,
			is_active,
			created_at,
			updated_at;
		`,
		request.Name,
		request.Code,
		request.UnitType,
	).Scan(
		&unit.ID,
		&unit.Name,
		&unit.Code,
		&unit.UnitType,
		&unit.IsActive,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)

	if err != nil {
		return models.InstitutionalUnit{}, err
	}

	return unit, nil
}

// ============================================================================
// MEMBRESÍAS / GESTORES
// ============================================================================

// GetInstitutionalUnitMemberships devuelve los usuarios asociados a una unidad.
//
// Se incluyen también membresías inactivas para que la administración pueda
// conocer el historial y reactivar posteriormente una relación existente.
func GetInstitutionalUnitMemberships(
	unitID int,
) ([]models.InstitutionalUnitMembership, error) {
	rows, err := database.DB.QueryContext(
		context.Background(),
		`
		SELECT
			membership.id,
			membership.unit_id,
			membership.user_id,
			membership.role,
			membership.is_active,
			membership.created_at,
			membership.updated_at,
			unit.name,
			user_account.full_name,
			user_account.email
		FROM institutional_unit_memberships membership

		INNER JOIN institutional_units unit
			ON unit.id = membership.unit_id

		INNER JOIN users user_account
			ON user_account.id = membership.user_id

		WHERE membership.unit_id = $1

		ORDER BY
			CASE
				WHEN membership.role = 'MANAGER' THEN 0
				ELSE 1
			END,
			user_account.full_name ASC;
		`,
		unitID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	memberships := []models.InstitutionalUnitMembership{}

	for rows.Next() {
		var membership models.InstitutionalUnitMembership

		if err := rows.Scan(
			&membership.ID,
			&membership.UnitID,
			&membership.UserID,
			&membership.Role,
			&membership.IsActive,
			&membership.CreatedAt,
			&membership.UpdatedAt,
			&membership.UnitName,
			&membership.UserFullName,
			&membership.UserEmail,
		); err != nil {
			return nil, err
		}

		memberships = append(memberships, membership)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return memberships, nil
}

// AddOrReactivateInstitutionalUnitMembership asigna un usuario a una unidad.
//
// Si la relación ya existía, no creamos una segunda fila: actualizamos el rol
// y reactivamos la membresía. Esto conserva una identidad estable para la
// relación y evita duplicados históricos.
func AddOrReactivateInstitutionalUnitMembership(
	unitID int,
	request models.AddInstitutionalUnitMembershipRequest,
) (models.InstitutionalUnitMembership, error) {
	var membership models.InstitutionalUnitMembership

	err := database.DB.QueryRowContext(
		context.Background(),
		`
		INSERT INTO institutional_unit_memberships (
			unit_id,
			user_id,
			role,
			is_active
		)
		VALUES (
			$1,
			$2,
			$3,
			true
		)

		ON CONFLICT (unit_id, user_id)
		DO UPDATE SET
			role = EXCLUDED.role,
			is_active = true

		RETURNING
			id,
			unit_id,
			user_id,
			role,
			is_active,
			created_at,
			updated_at;
		`,
		unitID,
		request.UserID,
		request.Role,
	).Scan(
		&membership.ID,
		&membership.UnitID,
		&membership.UserID,
		&membership.Role,
		&membership.IsActive,
		&membership.CreatedAt,
		&membership.UpdatedAt,
	)

	if err != nil {
		return models.InstitutionalUnitMembership{}, err
	}

	// Enriquecemos la respuesta después de guardar para mantener una única
	// representación pública de la membresía.
	err = database.DB.QueryRowContext(
		context.Background(),
		`
		SELECT
			unit.name,
			user_account.full_name,
			user_account.email
		FROM institutional_units unit

		INNER JOIN users user_account
			ON user_account.id = $2

		WHERE unit.id = $1;
		`,
		membership.UnitID,
		membership.UserID,
	).Scan(
		&membership.UnitName,
		&membership.UserFullName,
		&membership.UserEmail,
	)

	if err != nil {
		return models.InstitutionalUnitMembership{}, err
	}

	return membership, nil
}

// IsInstitutionalUnitManager responde únicamente a la pertenencia institucional.
//
// La condición de administrador global NO se evalúa aquí. Esa autorización
// pertenece al service/middleware:
//
//	administrador global OR manager activo de la unidad
//
// Mantener esa separación evita mezclar privilegios globales con relaciones
// específicas del dominio institucional.
func IsInstitutionalUnitManager(
	unitID int,
	userID int,
) (bool, error) {
	var isManager bool

	err := database.DB.QueryRowContext(
		context.Background(),
		`
		SELECT EXISTS (
			SELECT 1

			FROM institutional_unit_memberships membership

			INNER JOIN institutional_units unit
				ON unit.id = membership.unit_id

			WHERE membership.unit_id = $1
			  AND membership.user_id = $2
			  AND membership.role = 'MANAGER'
			  AND membership.is_active = true
			  AND unit.is_active = true
		);
		`,
		unitID,
		userID,
	).Scan(&isManager)

	if err != nil {
		return false, err
	}

	return isManager, nil
}

// GetManagedInstitutionalUnitsForUser devuelve las unidades sobre las cuales
// un usuario tiene autorización institucional activa como MANAGER.
//
// Será útil posteriormente para limitar la UI de programación sin exponer
// unidades que el usuario no administra.
func GetManagedInstitutionalUnitsForUser(
	userID int,
) ([]models.InstitutionalUnit, error) {
	rows, err := database.DB.QueryContext(
		context.Background(),
		`
		SELECT
			unit.id,
			unit.name,
			unit.code,
			unit.unit_type,
			unit.is_active,
			unit.created_at,
			unit.updated_at

		FROM institutional_units unit

		INNER JOIN institutional_unit_memberships membership
			ON membership.unit_id = unit.id

		WHERE membership.user_id = $1
		  AND membership.role = 'MANAGER'
		  AND membership.is_active = true
		  AND unit.is_active = true

		ORDER BY unit.name ASC;
		`,
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	units := []models.InstitutionalUnit{}

	for rows.Next() {
		var unit models.InstitutionalUnit

		if err := rows.Scan(
			&unit.ID,
			&unit.Name,
			&unit.Code,
			&unit.UnitType,
			&unit.IsActive,
			&unit.CreatedAt,
			&unit.UpdatedAt,
		); err != nil {
			return nil, err
		}

		units = append(units, unit)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return units, nil
}
