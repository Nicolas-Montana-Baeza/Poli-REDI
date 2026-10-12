package repositories

import (
	"database/sql"
	"errors"
	"testing"
)

func TestNonGroupResourceAcceptsAnyValidOperationalMode(
	t *testing.T,
) {
	err := validateResourceGroupInvariant(
		sql.NullInt64{},
		"OPEN_USE",
		sql.NullInt64{},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGroupResourceMustRemainReservable(t *testing.T) {
	err := validateResourceGroupInvariant(
		sql.NullInt64{
			Int64: 10,
			Valid: true,
		},
		"OPEN_USE",
		sql.NullInt64{
			Int64: 22,
			Valid: true,
		},
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
}

func TestGroupResourceRejectsNullCapacity(t *testing.T) {
	err := validateResourceGroupInvariant(
		sql.NullInt64{
			Int64: 10,
			Valid: true,
		},
		"RESERVABLE",
		sql.NullInt64{},
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
}

func TestGroupResourceRejectsCapacityBelowMinimum(
	t *testing.T,
) {
	err := validateResourceGroupInvariant(
		sql.NullInt64{
			Int64: 10,
			Valid: true,
		},
		"RESERVABLE",
		sql.NullInt64{
			Int64: 9,
			Valid: true,
		},
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
}

func TestGroupResourceAcceptsCapacityEqualMinimum(
	t *testing.T,
) {
	err := validateResourceGroupInvariant(
		sql.NullInt64{
			Int64: 10,
			Valid: true,
		},
		"RESERVABLE",
		sql.NullInt64{
			Int64: 10,
			Valid: true,
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
