package repositories

import (
	"context"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

func GetAllReservations() ([]models.Reservation, error) {
	rows, err := database.DB.QueryContext(
		context.Background(),
		`
		SELECT
			r.id,
			r.user_id,
			r.resource_id,
			r.activity_id,
			r.start_time,
			r.duration_minutes,
			r.status,
			r.created_at,
			r.updated_at,
			COALESCE(a.name, 'Reserva') AS activity_name
		FROM dbo.reservations r
		LEFT JOIN dbo.activities a
			ON a.id = r.activity_id
		ORDER BY r.start_time ASC;
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	reservations := []models.Reservation{}

	for rows.Next() {
		var reservation models.Reservation
		var activityName string

		err := rows.Scan(
			&reservation.ID,
			&reservation.UserID,
			&reservation.ResourceID,
			&reservation.ActivityID,
			&reservation.StartTime,
			&reservation.DurationMinutes,
			&reservation.Status,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
			&activityName,
		)

		if err != nil {
			return nil, err
		}

		reservation.Hour = reservation.StartTime.Format("15:04")
		reservation.Title = activityName
		reservation.Type = mapReservationType(reservation.Status)

		reservations = append(reservations, reservation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reservations, nil
}

func AddReservation(reservation models.Reservation) (models.Reservation, error) {
	err := database.DB.QueryRowContext(
		context.Background(),
		`
		INSERT INTO dbo.reservations (
			user_id,
			resource_id,
			activity_id,
			start_time,
			duration_minutes,
			status
		)
		OUTPUT
			INSERTED.id,
			INSERTED.user_id,
			INSERTED.resource_id,
			INSERTED.activity_id,
			INSERTED.start_time,
			INSERTED.duration_minutes,
			INSERTED.status,
			INSERTED.created_at,
			INSERTED.updated_at
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6);
		`,
		reservation.UserID,
		reservation.ResourceID,
		reservation.ActivityID,
		reservation.StartTime,
		reservation.DurationMinutes,
		reservation.Status,
	).Scan(
		&reservation.ID,
		&reservation.UserID,
		&reservation.ResourceID,
		&reservation.ActivityID,
		&reservation.StartTime,
		&reservation.DurationMinutes,
		&reservation.Status,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
	)

	if err != nil {
		return models.Reservation{}, err
	}

	reservation.Hour = reservation.StartTime.Format("15:04")
	reservation.Title = "Reserva"
	reservation.Type = mapReservationType(reservation.Status)

	return reservation, nil
}

func mapReservationType(status string) string {
	switch status {
	case "CONFIRMED":
		return "normal"

	case "PENDING":
		return "pending"

	case "CANCELLED":
		return "cancelled"

	default:
		return "normal"
	}
}

func CancelReservation(id int) (models.Reservation, error) {
	var reservation models.Reservation

	err := database.DB.QueryRowContext(
		context.Background(),
		`
		UPDATE dbo.reservations
		SET
			status = 'CANCELLED',
			updated_at = SYSUTCDATETIME()
		OUTPUT
			INSERTED.id,
			INSERTED.user_id,
			INSERTED.resource_id,
			INSERTED.activity_id,
			INSERTED.start_time,
			INSERTED.duration_minutes,
			INSERTED.status,
			INSERTED.created_at,
			INSERTED.updated_at
		WHERE id = @p1
		  AND status <> 'CANCELLED';
		`,
		id,
	).Scan(
		&reservation.ID,
		&reservation.UserID,
		&reservation.ResourceID,
		&reservation.ActivityID,
		&reservation.StartTime,
		&reservation.DurationMinutes,
		&reservation.Status,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
	)

	if err != nil {
		return models.Reservation{}, err
	}

	reservation.Hour = reservation.StartTime.Format("15:04")
	reservation.Title = "Reserva cancelada"
	reservation.Type = mapReservationType(reservation.Status)

	return reservation, nil
}

func IsUserAdmin(userID int) (bool, error) {
	var isAdmin bool

	err := database.DB.QueryRowContext(
		context.Background(),
		`
		SELECT is_admin
		FROM dbo.users
		WHERE id = @p1
		  AND is_blocked = 0;
		`,
		userID,
	).Scan(&isAdmin)

	if err != nil {
		return false, err
	}

	return isAdmin, nil
}
