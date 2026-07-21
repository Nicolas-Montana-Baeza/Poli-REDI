package repositories

import (
	"errors"
	"testing"
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
