//go:build integration

package repositories

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"poli-redi-api/internal/database"
)

func TestResourceOperationalSettingsPostgresIntegration(
	t *testing.T,
) {
	t.Setenv("MVP_SCOPE", "mvp3")

	if err := database.Connect(); err != nil {
		t.Fatalf("connect database: %v", err)
	}

	t.Cleanup(database.Close)

	ctx := context.Background()

	// ------------------------------------------------------------
	// 1. Recurso NO grupal: actualización válida y reversible.
	// ------------------------------------------------------------

	var (
		resourceID       int
		originalMode     string
		originalCapacity sql.NullInt64
		originalActive   bool
	)

	err := database.DB.QueryRowContext(
		ctx,
		`
		SELECT
			resource.id,
			resource.reservation_mode,
			resource.capacity,
			resource.is_active
		FROM resources resource
		WHERE NOT EXISTS (
			SELECT 1
			FROM reservation_policies policy
			INNER JOIN reservation_policy_group_resources group_resource
				ON group_resource.policy_id = policy.id
			   AND group_resource.resource_id = resource.id
			WHERE policy.is_published = true
			  AND policy.effective_from <= CURRENT_TIMESTAMP
			  AND (
				policy.effective_to IS NULL
				OR policy.effective_to > CURRENT_TIMESTAMP
			  )
		)
		ORDER BY resource.id
		LIMIT 1
		`,
	).Scan(
		&resourceID,
		&originalMode,
		&originalCapacity,
		&originalActive,
	)

	if err != nil {
		t.Fatalf("find non-group resource: %v", err)
	}

	// Restauración incluso si el test falla después del cambio.
	t.Cleanup(func() {
		var capacity any

		if originalCapacity.Valid {
			capacity = originalCapacity.Int64
		}

		_, restoreErr := database.DB.ExecContext(
			context.Background(),
			`
			UPDATE resources
			SET
				reservation_mode = $2,
				capacity = $3,
				is_active = $4
			WHERE id = $1
			`,
			resourceID,
			originalMode,
			capacity,
			originalActive,
		)

		if restoreErr != nil {
			t.Errorf(
				"restore resource %d: %v",
				resourceID,
				restoreErr,
			)
		}
	})

	newCapacity := 1

	if originalCapacity.Valid {
		newCapacity = int(originalCapacity.Int64) + 1
	}

	mode := originalMode
	active := originalActive

	updated, err := UpdateResourceOperationalSettings(
		resourceID,
		&mode,
		true,
		&newCapacity,
		&active,
	)

	if err != nil {
		t.Fatalf("valid operational update: %v", err)
	}

	if updated.Capacity == nil ||
		*updated.Capacity != newCapacity {

		t.Fatalf(
			"expected capacity %d, got %v",
			newCapacity,
			updated.Capacity,
		)
	}

	// ------------------------------------------------------------
	// 2. Recurso grupal vigente.
	// ------------------------------------------------------------

	var (
		groupResourceID int
		groupMode       string
		groupCapacity   sql.NullInt64
		groupActive     bool
		groupMinimum    int
	)

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT
			resource.id,
			resource.reservation_mode,
			resource.capacity,
			resource.is_active,
			group_resource.minimum_participants
		FROM reservation_policies policy
		INNER JOIN reservation_policy_group_resources group_resource
			ON group_resource.policy_id = policy.id
		INNER JOIN resources resource
			ON resource.id = group_resource.resource_id
		WHERE policy.is_published = true
		  AND policy.effective_from <= CURRENT_TIMESTAMP
		  AND (
			policy.effective_to IS NULL
			OR policy.effective_to > CURRENT_TIMESTAMP
		  )
		ORDER BY
			policy.effective_from DESC,
			policy.id DESC,
			resource.id
		LIMIT 1
		`,
	).Scan(
		&groupResourceID,
		&groupMode,
		&groupCapacity,
		&groupActive,
		&groupMinimum,
	)

	if err != nil {
		t.Fatalf("find current group resource: %v", err)
	}

	if groupMode != "RESERVABLE" {
		t.Fatalf(
			"group resource %d has unexpected mode %s",
			groupResourceID,
			groupMode,
		)
	}

	if !groupCapacity.Valid {
		t.Fatalf(
			"group resource %d has null capacity",
			groupResourceID,
		)
	}

	// ------------------------------------------------------------
	// 3. No puede convertirse a OPEN_USE.
	// ------------------------------------------------------------

	badMode := "OPEN_USE"

	_, err = UpdateResourceOperationalSettings(
		groupResourceID,
		&badMode,
		false,
		nil,
		nil,
	)

	if !errors.Is(
		err,
		ErrGroupResourceMustRemainReservable,
	) {
		t.Fatalf(
			"expected ErrGroupResourceMustRemainReservable, got %v",
			err,
		)
	}

	// ------------------------------------------------------------
	// 4. No puede quedar bajo el mínimo grupal.
	// ------------------------------------------------------------

	badCapacity := groupMinimum - 1

	_, err = UpdateResourceOperationalSettings(
		groupResourceID,
		nil,
		true,
		&badCapacity,
		nil,
	)

	if !errors.Is(
		err,
		ErrGroupResourceCapacityInvalid,
	) {
		t.Fatalf(
			"expected ErrGroupResourceCapacityInvalid, got %v",
			err,
		)
	}

	// ------------------------------------------------------------
	// 5. Los rechazos no modificaron PostgreSQL.
	// ------------------------------------------------------------

	var (
		finalMode     string
		finalCapacity sql.NullInt64
		finalActive   bool
	)

	err = database.DB.QueryRowContext(
		ctx,
		`
		SELECT
			reservation_mode,
			capacity,
			is_active
		FROM resources
		WHERE id = $1
		`,
		groupResourceID,
	).Scan(
		&finalMode,
		&finalCapacity,
		&finalActive,
	)

	if err != nil {
		t.Fatalf("reload group resource: %v", err)
	}

	if finalMode != groupMode {
		t.Fatalf(
			"group mode changed after rejected update: %s -> %s",
			groupMode,
			finalMode,
		)
	}

	capacityChanged :=
		finalCapacity.Valid != groupCapacity.Valid

	if finalCapacity.Valid && groupCapacity.Valid {
		capacityChanged =
			capacityChanged ||
				finalCapacity.Int64 != groupCapacity.Int64
	}

	if capacityChanged {
		t.Fatalf(
			"group capacity changed after rejected update",
		)
	}

	if finalActive != groupActive {
		t.Fatalf(
			"group active state changed unexpectedly",
		)
	}
}
