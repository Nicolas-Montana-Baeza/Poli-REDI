package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

var (
	ErrResourceNotAllowedByPolicy = errors.New("el recurso no esta permitido por la politica vigente")
	ErrReservationForbidden       = errors.New("no tienes permisos para cancelar esta reserva")
	ErrReservationNotCancellable  = errors.New("la reserva ya no se puede cancelar")
	ErrReservationFinalized       = errors.New("no puedes cancelar una reserva finalizada")
)

type RequestFrequencyError struct {
	NextDate time.Time
}

func (e RequestFrequencyError) Error() string {
	return fmt.Sprintf("ya tienes una solicitud vigente; proxima fecha permitida: %s", e.NextDate.Format("2006-01-02"))
}

const reservationColumns = `
	r.id, r.policy_id, r.user_id, r.resource_id, r.activity_id,
	r.start_time, r.duration_minutes, r.status, r.created_at, r.updated_at,
	COALESCE(a.name, 'Reserva') AS activity_name,
	res.name AS resource_name,
	COALESCE(u.full_name, '') AS user_full_name,
	COALESCE(u.email, '') AS user_email,
	COALESCE(u.rut, '') AS user_rut`

const reservationJoins = `
	FROM reservations r
	INNER JOIN resources res ON res.id = r.resource_id
	INNER JOIN users u ON u.id = r.user_id
	LEFT JOIN activities a ON a.id = r.activity_id`

type reservationScanner interface {
	Scan(...any) error
}

type reservationQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func scanReservation(scanner reservationScanner) (models.Reservation, error) {
	var reservation models.Reservation
	var activityName, resourceName, userFullName, userEmail, userRUT string
	err := scanner.Scan(
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

	reservation.StartTime = reservation.StartTime.In(businessclock.Location())
	reservation.CreatedAt = reservation.CreatedAt.In(businessclock.Location())
	reservation.UpdatedAt = reservation.UpdatedAt.In(businessclock.Location())
	reservation.Hour = reservation.StartTime.Format("15:04")
	reservation.Title = activityName
	reservation.Type = mapReservationType(reservation.Status)
	reservation.ResourceName = resourceName
	reservation.UserFullName = userFullName
	reservation.UserEmail = userEmail
	reservation.UserRUT = userRUT
	return reservation, nil
}

func scanReservationRows(rows *sql.Rows) ([]models.Reservation, error) {
	defer rows.Close()
	reservations := []models.Reservation{}
	for rows.Next() {
		reservation, err := scanReservation(rows)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, reservation)
	}
	return reservations, rows.Err()
}

func GetAllReservations() ([]models.Reservation, error) {
	rows, err := database.DB.QueryContext(context.Background(), `SELECT `+reservationColumns+reservationJoins+` ORDER BY r.start_time ASC`)
	if err != nil {
		return nil, err
	}
	return scanReservationRows(rows)
}

func GetReservationsByUserID(userID int) ([]models.Reservation, error) {
	rows, err := database.DB.QueryContext(context.Background(), `SELECT `+reservationColumns+reservationJoins+` WHERE r.user_id = $1 ORDER BY r.start_time DESC`, userID)
	if err != nil {
		return nil, err
	}
	return scanReservationRows(rows)
}

func GetReservationByID(id int) (models.Reservation, error) {
	return getReservationByID(context.Background(), database.DB, id)
}

func getReservationByID(ctx context.Context, q reservationQueryer, id int) (models.Reservation, error) {
	return scanReservation(q.QueryRowContext(ctx, `SELECT `+reservationColumns+reservationJoins+` WHERE r.id = $1`, id))
}

func GetActiveReservationsForAvailability(from, to time.Time, userID int, includeAllOpenUse bool) ([]models.Reservation, error) {
	rows, err := database.DB.QueryContext(context.Background(), `
		SELECT `+reservationColumns+reservationJoins+`
		WHERE r.status IN ('PENDING', 'CONFIRMED')
		  AND r.start_time < $2
		  AND r.end_time > $1
		  AND ($3::boolean OR r.reservation_mode_snapshot = 'RESERVABLE' OR r.user_id = $4)
		ORDER BY r.start_time ASC`, from, to, includeAllOpenUse, userID)
	if err != nil {
		return nil, err
	}
	return scanReservationRows(rows)
}

func GetAvailabilityBlocks(from, to time.Time) ([]models.AvailabilityBlock, error) {
	rows, err := database.DB.QueryContext(context.Background(), `
		SELECT b.id, b.resource_id, b.block_type, COALESCE(b.reason, ''),
		       b.start_time, b.end_time, res.name
		FROM availability_blocks b
		INNER JOIN resources res ON res.id = b.resource_id
		WHERE b.is_active = true
		  AND b.start_time < $2
		  AND b.end_time > $1
		ORDER BY b.start_time ASC`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	blocks := []models.AvailabilityBlock{}
	for rows.Next() {
		var block models.AvailabilityBlock
		if err := rows.Scan(&block.ID, &block.ResourceID, &block.BlockType, &block.Reason, &block.StartTime, &block.EndTime, &block.ResourceName); err != nil {
			return nil, err
		}
		block.StartTime = block.StartTime.In(businessclock.Location())
		block.EndTime = block.EndTime.In(businessclock.Location())
		blocks = append(blocks, block)
	}
	return blocks, rows.Err()
}

func GetCurrentReservationPolicy() (models.ReservationPolicy, error) {
	return GetCurrentReservationPolicyComplete()
}

func GetLatestConsumingReservation(userID int) (time.Time, int, error) {
	var createdAt time.Time
	var frequencyDays int
	err := database.DB.QueryRowContext(context.Background(), `
		SELECT r.created_at, p.request_frequency_days
		FROM reservations r
		INNER JOIN reservation_policies p ON p.id = r.policy_id
		WHERE r.user_id = $1
		  AND r.status IN ('PENDING', 'CONFIRMED')
		  AND r.reservation_mode_snapshot = 'RESERVABLE'
		ORDER BY ((r.created_at AT TIME ZONE 'America/Santiago')::date + p.request_frequency_days) DESC, r.id DESC
		LIMIT 1`, userID).Scan(&createdAt, &frequencyDays)
	return createdAt.In(businessclock.Location()), frequencyDays, err
}

func AddReservationWithPolicy(reservation models.Reservation, validate func(models.ReservationPolicy) error) (models.Reservation, error) {
	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return models.Reservation{}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(73002, $1)`, reservation.UserID); err != nil {
		return models.Reservation{}, err
	}

	policy, err := scanPolicy(tx.QueryRowContext(ctx, `SELECT `+policyColumns+` FROM reservation_policies WHERE is_published = true AND effective_from <= CURRENT_TIMESTAMP AND (effective_to IS NULL OR effective_to > CURRENT_TIMESTAMP) ORDER BY effective_from DESC, id DESC LIMIT 1 FOR SHARE`))
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

	var mode string
	if err := tx.QueryRowContext(ctx, `SELECT reservation_mode FROM resources WHERE id = $1 AND is_active = true`, reservation.ResourceID).Scan(&mode); err != nil {
		return models.Reservation{}, err
	}
	if mode == "RESERVABLE" {
		var previousCreatedAt time.Time
		var frequencyDays int
		err := tx.QueryRowContext(ctx, `
			SELECT r.created_at, p.request_frequency_days
			FROM reservations r
			INNER JOIN reservation_policies p ON p.id = r.policy_id
			WHERE r.user_id = $1 AND r.status IN ('PENDING', 'CONFIRMED')
			  AND r.reservation_mode_snapshot = 'RESERVABLE'
			ORDER BY ((r.created_at AT TIME ZONE 'America/Santiago')::date + p.request_frequency_days) DESC, r.id DESC
			LIMIT 1`, reservation.UserID).Scan(&previousCreatedAt, &frequencyDays)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return models.Reservation{}, err
		}
		if err == nil {
			nextDate := time.Date(previousCreatedAt.In(businessclock.Location()).Year(), previousCreatedAt.In(businessclock.Location()).Month(), previousCreatedAt.In(businessclock.Location()).Day(), 0, 0, 0, 0, businessclock.Location()).AddDate(0, 0, frequencyDays)
			now := businessclock.Now()
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, businessclock.Location())
			if today.Before(nextDate) {
				return models.Reservation{}, RequestFrequencyError{NextDate: nextDate}
			}
		}
	}

	var id int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO reservations (policy_id, user_id, resource_id, activity_id, start_time, duration_minutes, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`, policy.ID, reservation.UserID, reservation.ResourceID,
		reservation.ActivityID, reservation.StartTime, reservation.DurationMinutes, reservation.Status).Scan(&id)
	if err != nil {
		return models.Reservation{}, err
	}
	created, err := getReservationByID(ctx, tx, id)
	if err != nil {
		return models.Reservation{}, err
	}
	if err := tx.Commit(); err != nil {
		return models.Reservation{}, err
	}
	return created, nil
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
	case "PENDING":
		return "pending"
	case "CANCELLED":
		return "cancelled"
	default:
		return "normal"
	}
}

func GetReservationCancellationSnapshot(id int) (int, string, time.Time, int, error) {
	var ownerID, durationMinutes int
	var status string
	var startTime time.Time
	err := database.DB.QueryRowContext(context.Background(), `SELECT user_id, status, start_time, duration_minutes FROM reservations WHERE id = $1`, id).Scan(&ownerID, &status, &startTime, &durationMinutes)
	return ownerID, status, startTime.In(businessclock.Location()), durationMinutes, err
}

func CancelReservationAuthorized(id int, requestedBy models.LocalAuthUser, now time.Time) (models.Reservation, error) {
	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return models.Reservation{}, err
	}
	defer tx.Rollback()

	var ownerID int
	var status string
	var endTime time.Time
	if err := tx.QueryRowContext(ctx, `SELECT user_id, status, end_time FROM reservations WHERE id = $1 FOR UPDATE`, id).Scan(&ownerID, &status, &endTime); err != nil {
		return models.Reservation{}, err
	}
	if !requestedBy.IsAdmin && ownerID != requestedBy.ID {
		return models.Reservation{}, ErrReservationForbidden
	}
	if status != models.ReservationStatusConfirmed && status != models.ReservationStatusPending {
		return models.Reservation{}, ErrReservationNotCancellable
	}
	if !endTime.After(now) {
		return models.Reservation{}, ErrReservationFinalized
	}
	result, err := tx.ExecContext(ctx, `UPDATE reservations SET status = 'CANCELLED' WHERE id = $1 AND status IN ('CONFIRMED', 'PENDING') AND end_time > $2`, id, now)
	if err != nil {
		return models.Reservation{}, err
	}
	if count, _ := result.RowsAffected(); count != 1 {
		return models.Reservation{}, ErrReservationNotCancellable
	}
	cancelled, err := getReservationByID(ctx, tx, id)
	if err != nil {
		return models.Reservation{}, err
	}
	if err := tx.Commit(); err != nil {
		return models.Reservation{}, err
	}
	return cancelled, nil
}

func CancelReservation(id int) (models.Reservation, error) {
	var updatedID int
	err := database.DB.QueryRowContext(context.Background(), `UPDATE reservations SET status = 'CANCELLED' WHERE id = $1 AND status IN ('CONFIRMED', 'PENDING') AND end_time > CURRENT_TIMESTAMP RETURNING id`, id).Scan(&updatedID)
	if err != nil {
		return models.Reservation{}, err
	}
	return GetReservationByID(updatedID)
}

func IsUserAdmin(userID int) (bool, error) {
	var isAdmin bool
	err := database.DB.QueryRowContext(context.Background(), `SELECT is_admin FROM users WHERE id = $1 AND is_blocked = false`, userID).Scan(&isAdmin)
	return isAdmin, err
}
