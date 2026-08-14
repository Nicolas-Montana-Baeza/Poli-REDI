package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

const policyColumns = `id, reservable_window_days, request_frequency_days,
	confirmation_deadline_minutes, minimum_participants, opening_minute,
	closing_minute, slot_interval_minutes, effective_from, effective_to,
	created_by_user_id, created_at`

type policyScanner interface{ Scan(...any) error }

var ErrIdempotencyPayloadConflict = errors.New("idempotency payload conflict")
var ErrInvalidPolicyResource = errors.New("invalid policy resource")

func scanPolicy(row policyScanner) (models.ReservationPolicy, error) {
	var p models.ReservationPolicy
	var effectiveTo sql.NullTime
	var createdBy sql.NullInt64
	err := row.Scan(&p.ID, &p.ReservableWindowDays, &p.RequestFrequencyDays,
		&p.ConfirmationDeadlineMinutes, &p.MinimumParticipants, &p.OpeningMinute,
		&p.ClosingMinute, &p.SlotIntervalMinutes, &p.EffectiveFrom, &effectiveTo,
		&createdBy, &p.CreatedAt)
	if effectiveTo.Valid {
		p.EffectiveTo = &effectiveTo.Time
	}
	if createdBy.Valid {
		value := int(createdBy.Int64)
		p.CreatedByUserID = &value
	}
	return p, err
}

func loadPolicyCollections(ctx context.Context, q interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, p *models.ReservationPolicy) error {
	rows, err := q.QueryContext(ctx, `SELECT duration_minutes FROM reservation_policy_durations WHERE policy_id = $1 ORDER BY duration_minutes`, p.ID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var value int
		if err := rows.Scan(&value); err != nil {
			return err
		}
		p.AllowedDurations = append(p.AllowedDurations, value)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	rows.Close()
	rows, err = q.QueryContext(ctx, `SELECT resource_id FROM reservation_policy_resources WHERE policy_id = $1 ORDER BY resource_id`, p.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var value int
		if err := rows.Scan(&value); err != nil {
			return err
		}
		p.ResourceIDs = append(p.ResourceIDs, value)
	}
	return rows.Err()
}

func GetCurrentReservationPolicyComplete() (models.ReservationPolicy, error) {
	ctx := context.Background()
	p, err := scanPolicy(database.DB.QueryRowContext(ctx, `SELECT `+policyColumns+` FROM reservation_policies WHERE is_published = true AND effective_from <= CURRENT_TIMESTAMP AND (effective_to IS NULL OR effective_to > CURRENT_TIMESTAMP) ORDER BY effective_from DESC, id DESC LIMIT 1`))
	if err != nil {
		return p, err
	}
	err = loadPolicyCollections(ctx, database.DB, &p)
	return p, err
}

func GetReservationPolicyHistory() ([]models.ReservationPolicy, error) {
	ctx := context.Background()
	rows, err := database.DB.QueryContext(ctx, `SELECT `+policyColumns+` FROM reservation_policies ORDER BY effective_from DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	policies := []models.ReservationPolicy{}
	for rows.Next() {
		p, err := scanPolicy(rows)
		if err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	for index := range policies {
		if err := loadPolicyCollections(ctx, database.DB, &policies[index]); err != nil {
			return nil, err
		}
	}
	return policies, nil
}

func PublishReservationPolicy(request models.PublishReservationPolicyRequest, createdBy int, key, payloadHash string) (models.ReservationPolicy, bool, error) {
	ctx := context.Background()
	tx, err := database.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return models.ReservationPolicy{}, false, err
	}
	defer tx.Rollback()

	var existingHash string
	err = tx.QueryRowContext(ctx, `SELECT idempotency_payload_hash FROM dbo.reservation_policies WITH (UPDLOCK, HOLDLOCK) WHERE idempotency_key = @p1`, key).Scan(&existingHash)
	if err == nil {
		if existingHash != payloadHash {
			return models.ReservationPolicy{}, false, ErrIdempotencyPayloadConflict
		}
		existing, err := scanPolicy(tx.QueryRowContext(ctx, `SELECT `+policyColumns+` FROM dbo.reservation_policies WHERE idempotency_key = @p1`, key))
		if err != nil {
			return models.ReservationPolicy{}, false, err
		}
		if err := loadPolicyCollections(ctx, tx, &existing); err != nil {
			return models.ReservationPolicy{}, false, err
		}
		return existing, true, tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return models.ReservationPolicy{}, false, err
	}

	var id int
	err = tx.QueryRowContext(ctx, `
		DECLARE @now DATETIME2(0) = SYSUTCDATETIME();
		DECLARE @current_id INT;
		SELECT @current_id = id FROM dbo.reservation_policies WITH (UPDLOCK, HOLDLOCK) WHERE effective_to IS NULL;
		IF @current_id IS NOT NULL UPDATE dbo.reservation_policies SET effective_to = @now WHERE id = @current_id;
		INSERT INTO dbo.reservation_policies (reservable_window_days, request_frequency_days, confirmation_deadline_minutes, minimum_participants, opening_minute, closing_minute, slot_interval_minutes, effective_from, created_by_user_id, idempotency_key, idempotency_payload_hash, is_published)
		OUTPUT INSERTED.id VALUES (@p1,@p2,@p3,@p4,@p5,@p6,@p7,@now,@p8,@p9,@p10,0);`,
		request.ReservableWindowDays, request.RequestFrequencyDays, request.ConfirmationDeadlineMinutes,
		request.MinimumParticipants, request.OpeningMinute, request.ClosingMinute,
		request.SlotIntervalMinutes, createdBy, key, payloadHash).Scan(&id)
	if err != nil {
		return models.ReservationPolicy{}, false, err
	}
	for _, value := range request.AllowedDurations {
		if _, err := tx.ExecContext(ctx, `INSERT INTO dbo.reservation_policy_durations (policy_id, duration_minutes) VALUES (@p1,@p2)`, id, value); err != nil {
			return models.ReservationPolicy{}, false, err
		}
	}
	for _, value := range request.ResourceIDs {
		result, err := tx.ExecContext(ctx, `INSERT INTO dbo.reservation_policy_resources (policy_id, resource_id) SELECT @p1, id FROM dbo.resources WHERE id = @p2 AND is_active = 1`, id, value)
		if err != nil {
			return models.ReservationPolicy{}, false, err
		}
		count, _ := result.RowsAffected()
		if count != 1 {
			return models.ReservationPolicy{}, false, fmt.Errorf("%w: recurso %d no existe o no esta activo", ErrInvalidPolicyResource, value)
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE dbo.reservation_policies SET is_published = 1 WHERE id = @p1 AND is_published = 0`, id); err != nil {
		return models.ReservationPolicy{}, false, err
	}
	p, err := scanPolicy(tx.QueryRowContext(ctx, `SELECT `+policyColumns+` FROM dbo.reservation_policies WHERE id = @p1`, id))
	if err != nil {
		return models.ReservationPolicy{}, false, err
	}
	if err := loadPolicyCollections(ctx, tx, &p); err != nil {
		return models.ReservationPolicy{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return models.ReservationPolicy{}, false, err
	}
	return p, false, nil
}
