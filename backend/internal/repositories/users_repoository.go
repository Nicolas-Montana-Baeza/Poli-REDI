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

	return &user, nil
}
