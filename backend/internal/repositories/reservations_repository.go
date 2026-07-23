package repositories

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

var ErrResourceNotAllowedByPolicy = errors.New("el recurso no esta permitido por la politica vigente")
var ErrTargetForNonGroup = errors.New("targetParticipants solo se permite para solicitudes grupales")
var ErrInvalidTargetParticipants = errors.New("targetParticipants debe estar entre el minimo y la capacidad")

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
			,COALESCE(r.target_participants,r.group_capacity_snapshot),r.group_capacity_snapshot,p.minimum_participants,
			(SELECT COUNT(*) FROM dbo.participants pa WHERE pa.reservation_id=r.id AND pa.status='CONFIRMED'),
			p.confirmation_deadline_minutes
		FROM dbo.reservations r
		INNER JOIN dbo.resources res
			ON res.id = r.resource_id
		INNER JOIN dbo.users u
			ON u.id = r.user_id
		INNER JOIN dbo.reservation_policies p ON p.id=r.policy_id
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
			,COALESCE(r.target_participants,r.group_capacity_snapshot),r.group_capacity_snapshot,p.minimum_participants,
			(SELECT COUNT(*) FROM dbo.participants pa WHERE pa.reservation_id=r.id AND pa.status='CONFIRMED'),
			p.confirmation_deadline_minutes
		FROM dbo.reservations r
		INNER JOIN dbo.resources res
			ON res.id = r.resource_id
		INNER JOIN dbo.users u
			ON u.id = r.user_id
		INNER JOIN dbo.reservation_policies p ON p.id=r.policy_id
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

	result, scanErr := scanReservationRows(rows)
	if scanErr == nil {
		for i := range result {
			result[i].CanEditTarget = result[i].TargetParticipants != nil && result[i].ConfirmationDeadline != nil && !businessclock.Now().After(*result[i].ConfirmationDeadline)
		}
	}
	return result, scanErr
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
		var target, capacity, minimum, count sql.NullInt64
		var deadlineMinutes sql.NullInt64

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
			&target, &capacity, &minimum, &count, &deadlineMinutes,
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
		if target.Valid {
			v := int(target.Int64)
			reservation.TargetParticipants = &v
		}
		if capacity.Valid {
			v := int(capacity.Int64)
			reservation.Capacity = &v
		}
		if minimum.Valid {
			reservation.MinimumParticipants = int(minimum.Int64)
		}
		if count.Valid {
			reservation.ParticipantCount = int(count.Int64)
		}
		if deadlineMinutes.Valid && target.Valid {
			v := businessclock.ConfirmationDeadline(reservation.StartTime, int(deadlineMinutes.Int64))
			reservation.ConfirmationDeadline = &v
		}

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
	isGroup := false
	for _, id := range policy.GroupResourceIDs {
		if id == reservation.ResourceID {
			isGroup = true
			break
		}
	}
	var joinHash any
	var capacitySnapshot any
	var frozenCapacity int
	if isGroup {
		reservation.Status = initialGroupReservationStatus(policy.MinimumParticipants)
		sum := sha256.Sum256([]byte(reservation.JoinCode))
		joinHash = hex.EncodeToString(sum[:])
		if err := tx.QueryRowContext(ctx, `SELECT capacity FROM dbo.resources WITH(UPDLOCK,HOLDLOCK) WHERE id=@p1 AND is_active=1 AND capacity IS NOT NULL`, reservation.ResourceID).Scan(&frozenCapacity); err != nil {
			return models.Reservation{}, err
		}
		if frozenCapacity < policy.MinimumParticipants {
			return models.Reservation{}, errors.New("la capacidad del recurso es menor al minimo de participantes")
		}
		capacitySnapshot = frozenCapacity
		target := policy.MinimumParticipants
		if reservation.TargetParticipants != nil {
			target = *reservation.TargetParticipants
		}
		if target < policy.MinimumParticipants || target > frozenCapacity {
			return models.Reservation{}, ErrInvalidTargetParticipants
		}
		reservation.TargetParticipants = &target
	} else {
		if reservation.TargetParticipants != nil {
			return models.Reservation{}, ErrTargetForNonGroup
		}
		reservation.Status = models.ReservationStatusConfirmed
		reservation.JoinCode = ""
		joinHash = nil
		capacitySnapshot = nil
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
			status, join_code_hash, group_capacity_snapshot, target_participants
		)
		OUTPUT INSERTED.id INTO @created
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9, @p10);
		INSERT INTO dbo.participants(reservation_id,user_id,status,confirmed_at,is_owner)
		SELECT id,@p2,'CONFIRMED',SYSUTCDATETIME(),1 FROM @created WHERE @p8 IS NOT NULL;
		INSERT INTO dbo.reservation_participant_audit(reservation_id,actor_user_id,participant_user_id,action,previous_status,new_status,previous_reservation_status,new_reservation_status)
		SELECT id,@p2,@p2,'REQUESTER_ADDED',NULL,'CONFIRMED',@p7,@p7 FROM @created WHERE @p8 IS NOT NULL;

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
		joinHash,
		capacitySnapshot,
		reservation.TargetParticipants,
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
	if isGroup {
		reservation.Capacity = &frozenCapacity
		progress := assembleReservationProgress(
			reservation.ID, reservation.Status, 1, policy.MinimumParticipants,
			*reservation.TargetParticipants, frozenCapacity, reservation.StartTime,
			policy.ConfirmationDeadlineMinutes, true, true,
		)
		reservation.ParticipantCount = progress.ParticipantCount
		reservation.MinimumParticipants = progress.MinimumParticipants
		reservation.ConfirmationDeadline = &progress.ConfirmationDeadline
		reservation.CanEditTarget = progress.CanEditTarget
	}

	return reservation, nil
}

func initialGroupReservationStatus(minimum int) string {
	if minimum <= 1 {
		return models.ReservationStatusConfirmed
	}
	return models.ReservationStatusPending
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
