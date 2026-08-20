package repositories

import (
	"context"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func GetNotificationsByUserID(
	userID int,
) ([]models.Notification, error) {
	rows, err := database.DB.QueryContext(
		context.Background(),
		`
		SELECT
			id,
			user_id,
			reservation_id,
			title,
			message,
			type,
			is_read,
			created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY
			is_read ASC,
			created_at DESC
		LIMIT 20
		`,
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	notifications := []models.Notification{}

	for rows.Next() {
		var notification models.Notification

		if err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.ReservationID,
			&notification.Title,
			&notification.Message,
			&notification.Type,
			&notification.IsRead,
			&notification.CreatedAt,
		); err != nil {
			return nil, err
		}

		notifications = append(
			notifications,
			notification,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}
