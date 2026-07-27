package repositories

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
	"poli-redi-api/internal/validators"

	mssql "github.com/microsoft/go-mssqldb"
)

var (
	ErrRUTInvalid    = errors.New("el RUT no es válido")
	ErrRUTAlreadySet = errors.New("el RUT ya fue registrado y no puede modificarse")
	ErrRUTDuplicate  = errors.New("el RUT ya está registrado por otra cuenta")
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
	rut = validators.NormalizeRUT(rut)
	if !validators.IsValidRUT(rut) {
		return nil, ErrRUTInvalid
	}

	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var current string
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(rut,'') FROM dbo.users WITH(UPDLOCK,HOLDLOCK) WHERE id=@p1`, userID).Scan(&current); err != nil {
		return nil, err
	}
	current = strings.TrimSpace(strings.ToUpper(current))
	if current != "" && current != rut {
		return nil, ErrRUTAlreadySet
	}
	if current == "" {
		result, updateErr := tx.ExecContext(ctx, `UPDATE dbo.users SET rut=@p1,updated_at=SYSUTCDATETIME() WHERE id=@p2 AND NULLIF(LTRIM(RTRIM(rut)),'') IS NULL`, rut, userID)
		err = updateErr
		var sqlErr mssql.Error
		if errors.As(err, &sqlErr) && (sqlErr.Number == 2601 || sqlErr.Number == 2627) {
			return nil, ErrRUTDuplicate
		}
		if errors.As(err, &sqlErr) && (sqlErr.Number == 51010 || sqlErr.Number == 55003) {
			return nil, ErrRUTAlreadySet
		}
		if err != nil {
			return nil, err
		}
		affected, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			return nil, rowsErr
		}
		if affected != 1 {
			return nil, ErrRUTAlreadySet
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	user, err := getUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
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
