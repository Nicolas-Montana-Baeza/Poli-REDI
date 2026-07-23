package repositories

import (
	"errors"
	"testing"
	"time"
)

func TestParticipantTransition(t *testing.T) {
	cases := []struct {
		name, old, wantP, wantR            string
		exists, owner, confirm, wantMutate bool
		count, capacity, minimum           int
		wantErr                            error
	}{
		{"new", "", "CONFIRMED", "PENDING", false, false, true, true, 1, 10, 3, nil}, {"threshold", "", "CONFIRMED", "CONFIRMED", false, false, true, true, 2, 10, 3, nil}, {"duplicate", "CONFIRMED", "CONFIRMED", "CONFIRMED", true, false, true, false, 3, 10, 3, nil}, {"reconfirm", "CANCELLED", "CONFIRMED", "CONFIRMED", true, false, true, true, 2, 10, 3, nil}, {"withdraw", "CONFIRMED", "CANCELLED", "PENDING", true, false, false, true, 3, 10, 3, nil}, {"owner", "CONFIRMED", "CONFIRMED", "", true, true, false, false, 3, 10, 3, ErrOwnerCannotWithdraw}, {"capacity", "", "", "", false, false, true, false, 10, 10, 3, ErrGroupCapacity}}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m, p, r, e := participantTransition(c.exists, c.old, c.owner, c.count, c.capacity, c.minimum, c.confirm)
			if m != c.wantMutate || p != c.wantP || r != c.wantR || !errors.Is(e, c.wantErr) {
				t.Fatalf("got %v %q %q %v", m, p, r, e)
			}
		})
	}
}

func TestTargetValidationAndInclusiveDeadline(t *testing.T) {
	if err := validateTargetChange(10, 10, 22, 9); err != nil {
		t.Fatalf("valid target: %v", err)
	}
	if !errors.Is(validateTargetChange(9, 10, 22, 9), ErrInvalidTargetParticipants) {
		t.Fatal("minimum not enforced")
	}
	if !errors.Is(validateTargetChange(23, 10, 22, 9), ErrInvalidTargetParticipants) {
		t.Fatal("capacity not enforced")
	}
	if !errors.Is(validateTargetChange(11, 10, 22, 12), ErrTargetBelowConfirmed) {
		t.Fatal("confirmed count not enforced")
	}
	deadline := time.Date(2026, time.September, 6, 0, 30, 0, 0, time.FixedZone("America/Santiago", -4*60*60))
	if !targetDeadlineOpen(deadline, deadline) {
		t.Fatal("exact deadline must be inclusive")
	}
	if targetDeadlineOpen(deadline.Add(time.Nanosecond), deadline) {
		t.Fatal("after deadline must be rejected")
	}
}

func TestAssembleReservationProgressReturnsCompleteOwnerProgress(t *testing.T) {
	start := time.Date(2099, time.July, 10, 18, 0, 0, 0, time.UTC)
	progress := assembleReservationProgress(7, "PENDING", 1, 4, 8, 12, start, 60, true, true)
	if progress.ParticipantCount != 1 || progress.MinimumParticipants != 4 ||
		progress.TargetParticipants != 8 || progress.Capacity != 12 {
		t.Fatalf("incomplete progress: %+v", progress)
	}
	if progress.ConfirmationDeadline.IsZero() || !progress.CanEditTarget || !progress.IsOwner {
		t.Fatalf("owner/deadline progress is incoherent: %+v", progress)
	}
}
