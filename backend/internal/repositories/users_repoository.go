package repositories

import (
	"context"
	"database/sql"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func GetOrCreateUserByEmail(email string, fullName string) (*models.LocalAuthUser, error) {
	ctx := context.Background()
	user, err := getUserByEmail(ctx, email)

	if err == nil {
		return user, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	_, err = database.DB.ExecContext(
		ctx,
		`
		INSERT INTO dbo.users (
			email,
			full_name,
			is_admin,
			is_blocked
		)
		VALUES (@p1, @p2, 0, 0);
		`,
		email,
		fullName,
	)

	if err != nil {
		if existingUser, selectErr := getUserByEmail(ctx, email); selectErr == nil {
			return existingUser, nil
		}

		return nil, err
	}

	return getUserByEmail(ctx, email)
}

func getUserByEmail(ctx context.Context, email string) (*models.LocalAuthUser, error) {
	var user models.LocalAuthUser

	err := database.DB.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			email,
			full_name,
			COALESCE(rut, '') AS rut,
			is_admin,
			is_blocked,
			COALESCE(entra_oid, '') AS entra_oid,
			COALESCE(tenant_id, '') AS tenant_id
		FROM dbo.users
		WHERE email = @p1;
		`,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.RUT,
		&user.IsAdmin,
		&user.IsBlocked,
		&user.OID,
		&user.TenantID,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func UpdateUserEntraIdentity(userID int, oid string, tenantID string) (*models.LocalAuthUser, error) {
	ctx := context.Background()

	_, err := database.DB.ExecContext(
		ctx,
		`
		UPDATE dbo.users
		SET
			entra_oid = NULLIF(@p1, ''),
			tenant_id = NULLIF(@p2, ''),
			updated_at = SYSUTCDATETIME()
		WHERE id = @p3;
		`,
		oid,
		tenantID,
		userID,
	)

	if err != nil {
		return nil, err
	}

	return getUserByID(ctx, userID)
}

func GetAllUsers() ([]models.LocalAuthUser, error) {
	rows, err := database.DB.QueryContext(
		context.Background(),
		`
		SELECT
			id,
			email,
			full_name,
			COALESCE(rut, '') AS rut,
			is_admin,
			is_blocked,
			COALESCE(entra_oid, '') AS entra_oid,
			COALESCE(tenant_id, '') AS tenant_id
		FROM dbo.users
		ORDER BY full_name ASC, email ASC;
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []models.LocalAuthUser{}

	for rows.Next() {
		var user models.LocalAuthUser

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FullName,
			&user.RUT,
			&user.IsAdmin,
			&user.IsBlocked,
			&user.OID,
			&user.TenantID,
		)

		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func UpdateUserRUT(userID int, rut string) (*models.LocalAuthUser, error) {
	ctx := context.Background()
	var rutValue any = rut

	if rut == "" {
		rutValue = nil
	}

	_, err := database.DB.ExecContext(
		ctx,
		`
		UPDATE dbo.users
		SET
			rut = @p1,
			updated_at = SYSUTCDATETIME()
		WHERE id = @p2;
		`,
		rutValue,
		userID,
	)

	if err != nil {
		return nil, err
	}

	return getUserByID(ctx, userID)
}

func getUserByID(ctx context.Context, userID int) (*models.LocalAuthUser, error) {
	var user models.LocalAuthUser

	err := database.DB.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			email,
			full_name,
			COALESCE(rut, '') AS rut,
			is_admin,
			is_blocked,
			COALESCE(entra_oid, '') AS entra_oid,
			COALESCE(tenant_id, '') AS tenant_id
		FROM dbo.users
		WHERE id = @p1;
		`,
		userID,
	).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.RUT,
		&user.IsAdmin,
		&user.IsBlocked,
		&user.OID,
		&user.TenantID,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
