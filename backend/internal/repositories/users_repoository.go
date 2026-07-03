package repositories

import (
	"context"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func GetOrCreateUserByEmail(email string, fullName string) (*models.LocalAuthUser, error) {
	query := `
		MERGE dbo.users WITH (HOLDLOCK) AS target
		USING (
			SELECT
				@p1 AS email,
				@p2 AS full_name
		) AS source
		ON target.email = source.email
		WHEN MATCHED THEN
			UPDATE SET
				full_name = source.full_name,
				updated_at = SYSUTCDATETIME()
		WHEN NOT MATCHED THEN
			INSERT (
				email,
				full_name,
				is_admin,
				is_blocked
			)
			VALUES (
				source.email,
				source.full_name,
				0,
				0
			)
		OUTPUT
			inserted.id,
			inserted.email,
			inserted.full_name,
			inserted.is_admin,
			inserted.is_blocked;
	`

	var user models.LocalAuthUser

	err := database.DB.QueryRowContext(
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
