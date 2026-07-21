package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

var ErrResourceNotAllowedByPolicy = errors.New("el recurso no esta permitido por la politica vigente")

func GetAllReservations() ([]models.Reservation, error) {
	rows, err := database.DB.QueryContext(
		context.Background(),
		`
		SELECT
			r.id,
			r.policy_id,
			r.user_id,
			r.resource_id,
			r.activity_id,
			r.start_time,
			r.duration_minutes,
			r.status,
			r.created_at,
			r.updated_at,
			COALESCE(a.name, 'Reserva') AS activity_name,
			res.name AS resource_name,
			COALESCE(u.full_name, '') AS user_full_name,
			COALESCE(u.email, '') AS user_email,
			COALESCE(u.rut, '') AS user_rut
		FROM dbo.reservations r
		INNER JOIN dbo.resources res
			ON res.id = r.resource_id
		INNER JOIN dbo.users u
			ON u.id = r.user_id
		LEFT JOIN dbo.activities a
			ON a.id = r.activity_id
		ORDER BY r.start_time ASC;
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return scanReservationRows(rows)
}

func GetReservationsByUserID(userID int) ([]models.Reservation, error) {
	rows, err := database.DB.QueryContext(
		context.Background(),
		`
		SELECT
			r.id,
			r.policy_id,
			r.user_id,
			r.resource_id,
			r.activity_id,
			r.start_time,
			r.duration_minutes,
			r.status,
			r.created_at,
			r.updated_at,
			COALESCE(a.name, 'Reserva') AS activity_name,
			res.name AS resource_name,
			COALESCE(u.full_name, '') AS user_full_name,
			COALESCE(u.email, '') AS user_email,
			COALESCE(u.rut, '') AS user_rut
		FROM dbo.reservations r
		INNER JOIN dbo.resources res
			ON res.id = r.resource_id
		INNER JOIN dbo.users u
			ON u.id = r.user_id
		LEFT JOIN dbo.activities a
			ON a.id = r.activity_id
		WHERE r.user_id = @p1
		ORDER BY r.start_time DESC;
		`,
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return scanReservationRows(rows)
}

func GetCurrentReservationPolicy() (models.ReservationPolicy, error) {
	return GetCurrentReservationPolicyComplete()
}

func GetLatestConsumingReservation(userID int) (time.Time, int, error) {
	var createdAt time.Time
	var frequencyDays int

	err := database.DB.QueryRowContext(
		context.Background(),
		`
		SELECT TOP (1)
			r.created_at,
			p.request_frequency_days
		FROM dbo.reservations r
		INNER JOIN dbo.reservation_policies p ON p.id = r.policy_id
		WHERE r.user_id = @p1
		  AND r.status IN ('PENDING', 'CONFIRMED')
		ORDER BY DATEADD(
			DAY,
			p.request_frequency_days,
			CONVERT(DATE, r.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Pacific SA Standard Time')
		) DESC, r.id DESC;
		`,
		userID,
	).Scan(&createdAt, &frequencyDays)

	return createdAt, frequencyDays, err
}

func scanReservationRows(rows *sql.Rows) ([]models.Reservation, error) {
	reservations := []models.Reservation{}

	for rows.Next() {
		var reservation models.Reservation
		var activityName string
		var resourceName string
		var userFullName string
		var userEmail string
		var userRUT string

		err := rows.Scan(
			&reservation.ID,
			&reservation.PolicyID,
			&reservation.UserID,
			&reservation.ResourceID,
			&reservation.ActivityID,
			&reservation.StartTime,
			&reservation.DurationMinutes,
			&reservation.Status,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
			&activityName,
			&resourceName,
			&userFullName,
			&userEmail,
			&userRUT,
		)

		if err != nil {
			return nil, err
		}

		// SQL Server DATETIME2 no tiene zona. start_time se guarda como hora de
		// muro institucional y recibe APP_TIMEZONE solo despues de escanearlo.
		reservation.StartTime = businessclock.FromDatabaseWallTime(reservation.StartTime)
		reservation.Hour = reservation.StartTime.Format("15:04")
		reservation.Title = activityName
		reservation.Type = mapReservationType(reservation.Status)
		reservation.ResourceName = resourceName
		reservation.UserFullName = userFullName
		reservation.UserEmail = userEmail
		reservation.UserRUT = userRUT

		reservations = append(reservations, reservation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reservations, nil
}

func AddReservationWithPolicy(reservation models.Reservation, validate func(models.ReservationPolicy) error) (models.Reservation, error) {
	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return models.Reservation{}, err
	}
	defer tx.Rollback()
	policy, err := scanPolicy(tx.QueryRowContext(ctx, `SELECT TOP (1) `+policyColumns+` FROM dbo.reservation_policies WITH (UPDLOCK, HOLDLOCK) WHERE is_published = 1 AND effective_from <= SYSUTCDATETIME() AND (effective_to IS NULL OR effective_to > SYSUTCDATETIME()) ORDER BY effective_from DESC, id DESC`))
	if err != nil {
		return models.Reservation{}, err
	}
	if err := loadPolicyCollections(ctx, tx, &policy); err != nil {
		return models.Reservation{}, err
	}
	if !policyAllowsResource(policy, reservation.ResourceID) {
		return models.Reservation{}, ErrResourceNotAllowedByPolicy
	}
	if err := validate(policy); err != nil {
		return models.Reservation{}, err
	}

	var activityName string
	var resourceName string
	var userFullName string
	var userEmail string
	var userRUT string

	err = tx.QueryRowContext(
		ctx,
		`
		DECLARE @created TABLE (id INT);
		INSERT INTO dbo.reservations (
			policy_id,
			user_id,
			resource_id,
			activity_id,
			start_time,
			duration_minutes,
			status
		)
		OUTPUT INSERTED.id INTO @created
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7);

		SELECT
			r.id,
			r.policy_id,
			r.user_id,
			r.resource_id,
			r.activity_id,
			r.start_time,
			r.duration_minutes,
			r.status,
			r.created_at,
			r.updated_at,
			COALESCE(a.name, 'Reserva') AS activity_name,
			res.name AS resource_name,
			COALESCE(u.full_name, '') AS user_full_name,
			COALESCE(u.email, '') AS user_email,
			COALESCE(u.rut, '') AS user_rut
		FROM dbo.reservations r
		INNER JOIN @created c
			ON c.id = r.id
		INNER JOIN dbo.resources res
			ON res.id = r.resource_id
		INNER JOIN dbo.users u
			ON u.id = r.user_id
		LEFT JOIN dbo.activities a
			ON a.id = r.activity_id;
		`,
		policy.ID,
		reservation.UserID,
		reservation.ResourceID,
		reservation.ActivityID,
		// Persistimos los campos de hora chilena que esperan los triggers de
		// reserva; timestamps como updated_at siguen usando SYSUTCDATETIME().
		businessclock.ToDatabaseWallTime(reservation.StartTime),
		reservation.DurationMinutes,
		reservation.Status,
	).Scan(
		&reservation.ID,
		&reservation.PolicyID,
		&reservation.UserID,
		&reservation.ResourceID,
		&reservation.ActivityID,
		&reservation.StartTime,
		&reservation.DurationMinutes,
		&reservation.Status,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
		&activityName,
		&resourceName,
		&userFullName,
		&userEmail,
		&userRUT,
	)

	if err != nil {
		return models.Reservation{}, err
	}
	if err := tx.Commit(); err != nil {
		return models.Reservation{}, err
	}

	reservation.StartTime = businessclock.FromDatabaseWallTime(reservation.StartTime)
	reservation.Hour = reservation.StartTime.Format("15:04")
	reservation.Title = activityName
	reservation.Type = mapReservationType(reservation.Status)
	reservation.ResourceName = resourceName
	reservation.UserFullName = userFullName
	reservation.UserEmail = userEmail
	reservation.UserRUT = userRUT

	return reservation, nil
}

func policyAllowsResource(policy models.ReservationPolicy, resourceID int) bool {
	for _, allowedID := range policy.ResourceIDs {
		if allowedID == resourceID {
			return true
		}
	}
	return false
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

func GetReservationCancellationSnapshot(id int) (int, string, time.Time, int, error) {
	var ownerID int
	var status string
	var startTime time.Time
	var durationMinutes int

	err := database.DB.QueryRowContext(
		context.Background(),
		`
		SELECT
			user_id,
			status,
			start_time,
			duration_minutes
		FROM dbo.reservations
		WHERE id = @p1;
		`,
		id,
	).Scan(
		&ownerID,
		&status,
		&startTime,
		&durationMinutes,
	)

	if err != nil {
		return 0, "", time.Time{}, 0, err
	}

	return ownerID, status, businessclock.FromDatabaseWallTime(startTime), durationMinutes, nil
}

func CancelReservation(id int) (models.Reservation, error) {
	var reservation models.Reservation
	var activityName string
	var resourceName string
	var userFullName string
	var userEmail string
	var userRUT string

	err := database.DB.QueryRowContext(
		context.Background(),
		`
		DECLARE @updated TABLE (id INT);

		UPDATE dbo.reservations
		SET
			status = 'CANCELLED',
			updated_at = SYSUTCDATETIME()
		OUTPUT INSERTED.id INTO @updated
		WHERE id = @p1
		  AND status IN ('CONFIRMED', 'PENDING');

		SELECT
			r.id,
			r.policy_id,
			r.user_id,
			r.resource_id,
			r.activity_id,
			r.start_time,
			r.duration_minutes,
			r.status,
			r.created_at,
			r.updated_at,
			COALESCE(a.name, 'Reserva') AS activity_name,
			res.name AS resource_name,
			COALESCE(usr.full_name, '') AS user_full_name,
			COALESCE(usr.email, '') AS user_email,
			COALESCE(usr.rut, '') AS user_rut
		FROM dbo.reservations r
		INNER JOIN @updated u
			ON u.id = r.id
		INNER JOIN dbo.resources res
			ON res.id = r.resource_id
		INNER JOIN dbo.users usr
			ON usr.id = r.user_id
		LEFT JOIN dbo.activities a
			ON a.id = r.activity_id;
		`,
		id,
	).Scan(
		&reservation.ID,
		&reservation.PolicyID,
		&reservation.UserID,
		&reservation.ResourceID,
		&reservation.ActivityID,
		&reservation.StartTime,
		&reservation.DurationMinutes,
		&reservation.Status,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
		&activityName,
		&resourceName,
		&userFullName,
		&userEmail,
		&userRUT,
	)

	if err != nil {
		return models.Reservation{}, err
	}

	reservation.StartTime = businessclock.FromDatabaseWallTime(reservation.StartTime)
	reservation.Hour = reservation.StartTime.Format("15:04")
	reservation.Title = activityName
	reservation.Type = mapReservationType(reservation.Status)
	reservation.ResourceName = resourceName
	reservation.UserFullName = userFullName
	reservation.UserEmail = userEmail
	reservation.UserRUT = userRUT

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
