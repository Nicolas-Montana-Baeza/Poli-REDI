package repositories

import (
	"context"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func GetOrCreateUserByEmail(email string, fullName string) (*models.LocalAuthUser, error) {
	query := `
		INSERT INTO users (
			email,
			full_name,
			is_admin,
			is_blocked
		)
		VALUES ($1, $2, false, false)
		ON CONFLICT (email)
		DO UPDATE SET
			full_name = EXCLUDED.full_name,
			updated_at = CURRENT_TIMESTAMP
		RETURNING
			id,
			email,
			full_name,
			is_admin,
			is_blocked;
	`

	var user models.LocalAuthUser

	err := database.DB.QueryRow(
		context.Background(),
		query,
		email,
		fullName,
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
