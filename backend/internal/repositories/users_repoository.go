package repositories

import (
	"context"
	"database/sql"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func GetOrCreateUserByEmail(email string, fullName string) (*models.LocalAuthUser, error) {
	ctx := context.Background()

	tx, err := database.DB.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	result, err := tx.ExecContext(
		ctx,
		`
		UPDATE dbo.users
		SET
			full_name = @p2,
			updated_at = SYSUTCDATETIME()
		WHERE email = @p1;
		`,
		email,
		fullName,
	)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		_, err = tx.ExecContext(
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
			return nil, err
		}
	}

	var user models.LocalAuthUser

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			email,
			full_name,
			is_admin,
			is_blocked
		FROM dbo.users
		WHERE email = @p1;
		`,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.IsAdmin,
		&user.IsBlocked,
	)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &user, nil
}
