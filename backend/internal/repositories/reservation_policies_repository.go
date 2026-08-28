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

func PublishReservationPolicy(
	request models.PublishReservationPolicyRequest,
	createdBy int,
	key string,
	payloadHash string,
) (models.ReservationPolicy, bool, error) {
	ctx := context.Background()

	tx, err := database.DB.BeginTx(
		ctx,
		&sql.TxOptions{
			Isolation: sql.LevelSerializable,
		},
	)
	if err != nil {
		return models.ReservationPolicy{}, false, err
	}

	defer tx.Rollback()

	// Una sola publicacion administrativa puede modificar la linea temporal
	// de reservation_policies a la vez.
	//
	// SHARE ROW EXCLUSIVE permite lecturas normales y serializa publicadores
	// sin hints especificos de SQL Server.
	if _, err := tx.ExecContext(
		ctx,
		`LOCK TABLE reservation_policies
		 IN SHARE ROW EXCLUSIVE MODE`,
	); err != nil {
		return models.ReservationPolicy{}, false, err
	}

	var existingHash string

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT idempotency_payload_hash
		FROM reservation_policies
		WHERE idempotency_key = $1
		`,
		key,
	).Scan(&existingHash)

	if err == nil {
		if existingHash != payloadHash {
			return models.ReservationPolicy{},
				false,
				ErrIdempotencyPayloadConflict
		}

		existing, err := scanPolicy(
			tx.QueryRowContext(
				ctx,
				`
				SELECT `+policyColumns+`
				FROM reservation_policies
				WHERE idempotency_key = $1
				`,
				key,
			),
		)
		if err != nil {
			return models.ReservationPolicy{},
				false,
				err
		}

		if err := loadPolicyCollections(
			ctx,
			tx,
			&existing,
		); err != nil {
			return models.ReservationPolicy{},
				false,
				err
		}

		if err := tx.Commit(); err != nil {
			return models.ReservationPolicy{},
				false,
				err
		}

		return existing, true, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return models.ReservationPolicy{},
			false,
			err
	}

	// Conservamos la politica vigente para heredar configuracion grupal que
	// todavia no forma parte del contrato administrativo.
	//
	// Actualmente PublishReservationPolicyRequest no expone:
	//   - groupResourceIds;
	//   - lateWithdrawalMinutes;
	//   - groupRecoveryDeadlineMinutes.
	//
	// Hasta que producto cierre CAND-002, publicar una politica NO debe
	// resetear silenciosamente estos valores.
	currentPolicyID := 0
	currentLateWithdrawalMinutes := 60
	currentGroupRecoveryDeadlineMinutes := 0

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			late_withdrawal_minutes,
			group_recovery_deadline_minutes
		FROM reservation_policies
		WHERE is_published = true
		  AND effective_to IS NULL
		ORDER BY effective_from DESC, id DESC
		LIMIT 1
		FOR UPDATE
		`,
	).Scan(
		&currentPolicyID,
		&currentLateWithdrawalMinutes,
		&currentGroupRecoveryDeadlineMinutes,
	)

	if err != nil &&
		!errors.Is(err, sql.ErrNoRows) {
		return models.ReservationPolicy{},
			false,
			err
	}

	if currentPolicyID > 0 {
		if _, err := tx.ExecContext(
			ctx,
			`
			UPDATE reservation_policies
			SET effective_to = CURRENT_TIMESTAMP
			WHERE id = $1
			  AND effective_to IS NULL
			`,
			currentPolicyID,
		); err != nil {
			return models.ReservationPolicy{},
				false,
				err
		}
	}

	var id int

	err = tx.QueryRowContext(
		ctx,
		`
		INSERT INTO reservation_policies (
			reservable_window_days,
			request_frequency_days,
			confirmation_deadline_minutes,
			minimum_participants,
			opening_minute,
			closing_minute,
			slot_interval_minutes,
			effective_from,
			created_by_user_id,
			idempotency_key,
			idempotency_payload_hash,
			is_published,
			late_withdrawal_minutes,
			group_recovery_deadline_minutes
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			CURRENT_TIMESTAMP,
			$8,
			$9,
			$10,
			false,
			$11,
			$12
		)
		RETURNING id
		`,
		request.ReservableWindowDays,
		request.RequestFrequencyDays,
		request.ConfirmationDeadlineMinutes,
		request.MinimumParticipants,
		request.OpeningMinute,
		request.ClosingMinute,
		request.SlotIntervalMinutes,
		createdBy,
		key,
		payloadHash,
		currentLateWithdrawalMinutes,
		currentGroupRecoveryDeadlineMinutes,
	).Scan(&id)

	if err != nil {
		return models.ReservationPolicy{},
			false,
			err
	}

	for _, value := range request.AllowedDurations {
		if _, err := tx.ExecContext(
			ctx,
			`
			INSERT INTO reservation_policy_durations (
				policy_id,
				duration_minutes
			)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
			`,
			id,
			value,
		); err != nil {
			return models.ReservationPolicy{},
				false,
				err
		}
	}

	for _, value := range request.ResourceIDs {
		result, err := tx.ExecContext(
			ctx,
			`
			INSERT INTO reservation_policy_resources (
				policy_id,
				resource_id
			)
			SELECT
				$1,
				id
			FROM resources
			WHERE id = $2
			  AND is_active = true
			ON CONFLICT DO NOTHING
			`,
			id,
			value,
		)
		if err != nil {
			return models.ReservationPolicy{},
				false,
				err
		}

		count, err := result.RowsAffected()
		if err != nil {
			return models.ReservationPolicy{},
				false,
				err
		}

		if count != 1 {
			return models.ReservationPolicy{},
				false,
				fmt.Errorf(
					"%w: recurso %d no existe o no esta activo",
					ErrInvalidPolicyResource,
					value,
				)
		}
	}

	var hasGroupResourceTable bool

	if err := tx.QueryRowContext(
		ctx,
		`
		SELECT
			to_regclass(
				'public.reservation_policy_group_resources'
			) IS NOT NULL
		`,
	).Scan(&hasGroupResourceTable); err != nil {
		return models.ReservationPolicy{},
			false,
			err
	}

	if hasGroupResourceTable &&
		currentPolicyID > 0 {
		if _, err := tx.ExecContext(
			ctx,
			`
			INSERT INTO reservation_policy_group_resources (
				policy_id,
				resource_id,
				minimum_participants
			)
			SELECT
				$1,
				old_group.resource_id,
				old_group.minimum_participants
			FROM reservation_policy_group_resources old_group
			INNER JOIN reservation_policy_resources new_scope
				ON new_scope.policy_id = $1
				AND new_scope.resource_id =
					old_group.resource_id
			WHERE old_group.policy_id = $2
			ON CONFLICT DO NOTHING
			`,
			id,
			currentPolicyID,
		); err != nil {
			return models.ReservationPolicy{},
				false,
				err
		}
	}

	if _, err := tx.ExecContext(
		ctx,
		`
		UPDATE reservation_policies
		SET is_published = true
		WHERE id = $1
		  AND is_published = false
		`,
		id,
	); err != nil {
		return models.ReservationPolicy{},
			false,
			err
	}

	p, err := scanPolicy(
		tx.QueryRowContext(
			ctx,
			`
			SELECT `+policyColumns+`
			FROM reservation_policies
			WHERE id = $1
			`,
			id,
		),
	)
	if err != nil {
		return models.ReservationPolicy{},
			false,
			err
	}

	if err := loadPolicyCollections(
		ctx,
		tx,
		&p,
	); err != nil {
		return models.ReservationPolicy{},
			false,
			err
	}

	if err := tx.Commit(); err != nil {
		return models.ReservationPolicy{},
			false,
			err
	}

	return p, false, nil
}
