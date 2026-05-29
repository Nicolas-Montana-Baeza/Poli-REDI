package repositories

import (
	"context"
	"errors"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func GetUserByEmail(email string) (*models.LocalAuthUser, error) {
	query := `
		SELECT id, email, full_name, is_admin, is_blocked
		FROM users
		WHERE email = $1
		LIMIT 1;
	`

	var user models.LocalAuthUser

	err := database.DB.QueryRow(
		context.Background(),
		query,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.IsAdmin,
		&user.IsBlocked,
	)

	if err != nil {
		return nil, errors.New("usuario no registrado en Poli-REDI")
	}

	return &user, nil
}
