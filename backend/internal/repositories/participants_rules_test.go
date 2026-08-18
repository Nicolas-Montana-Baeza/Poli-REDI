package repositories

import (
	"errors"
	"testing"
)

func TestParticipantTransitionPendingReachesMinimum(t *testing.T) {
	mutate, participantStatus, reservationStatus, err :=
		participantTransition(
			false,
			"",
			false,
			9,
			22,
			10,
			"PENDING",
			true,
		)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mutate {
		t.Fatal("expected mutation")
	}

	if participantStatus != "CONFIRMED" {
		t.Fatalf(
			"expected participant CONFIRMED, got %s",
			participantStatus,
		)
	}

	if reservationStatus != "CONFIRMED" {
		t.Fatalf(
			"expected reservation CONFIRMED, got %s",
			reservationStatus,
		)
	}
}

func TestConfirmedReservationDoesNotReturnToPending(t *testing.T) {
	mutate, participantStatus, reservationStatus, err :=
		participantTransition(
			true,
			"CONFIRMED",
			false,
			10,
			22,
			10,
			"CONFIRMED",
			false,
		)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mutate {
		t.Fatal("expected mutation")
	}

	if participantStatus != "CANCELLED" {
		t.Fatalf(
			"expected participant CANCELLED, got %s",
			participantStatus,
		)
	}

	// Regla B.
	if reservationStatus != "CONFIRMED" {
		t.Fatalf(
			"expected reservation to remain CONFIRMED, got %s",
			reservationStatus,
		)
	}

	condition := participantGroupCondition(
		reservationStatus,
		9,
		10,
	)

	if condition != GroupConditionAtRisk {
		t.Fatalf(
			"expected AT_RISK, got %s",
			condition,
		)
	}
}

func TestOwnerCannotWithdraw(t *testing.T) {
	_, _, _, err :=
		participantTransition(
			true,
			"CONFIRMED",
			true,
			10,
			22,
			10,
			"CONFIRMED",
			false,
		)

	if !errors.Is(err, ErrOwnerCannotWithdraw) {
		t.Fatalf(
			"expected ErrOwnerCannotWithdraw, got %v",
			err,
		)
	}
}

func TestGroupCapacityIsEnforced(t *testing.T) {
	_, _, _, err :=
		participantTransition(
			false,
			"",
			false,
			22,
			22,
			10,
			"CONFIRMED",
			true,
		)

	if !errors.Is(err, ErrGroupCapacity) {
		t.Fatalf(
			"expected ErrGroupCapacity, got %v",
			err,
		)
	}
}

func TestPendingReservationRemainsPendingBelowMinimum(t *testing.T) {
	_, _, reservationStatus, err :=
		participantTransition(
			false,
			"",
			false,
			8,
			22,
			10,
			"PENDING",
			true,
		)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reservationStatus != "PENDING" {
		t.Fatalf(
			"expected PENDING, got %s",
			reservationStatus,
		)
	}
}

func TestAlreadyConfirmedParticipationIsIdempotent(t *testing.T) {
	mutate, participantStatus, reservationStatus, err :=
		participantTransition(
			true,
			"CONFIRMED",
			false,
			10,
			22,
			10,
			"CONFIRMED",
			true,
		)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mutate {
		t.Fatal("expected idempotent operation")
	}

	if participantStatus != "CONFIRMED" {
		t.Fatalf(
			"expected participant CONFIRMED, got %s",
			participantStatus,
		)
	}

	if reservationStatus != "CONFIRMED" {
		t.Fatalf(
			"expected reservation CONFIRMED, got %s",
			reservationStatus,
		)
	}
}

func TestInvalidGroupConfiguration(t *testing.T) {
	_, _, _, err :=
		participantTransition(
			false,
			"",
			false,
			0,
			5,
			10,
			"PENDING",
			true,
		)

	if !errors.Is(err, ErrInvalidGroupConfig) {
		t.Fatalf(
			"expected ErrInvalidGroupConfig, got %v",
			err,
		)
	}
}
