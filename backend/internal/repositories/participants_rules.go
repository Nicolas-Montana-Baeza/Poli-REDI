package repositories

import "time"

func participantTransition(exists bool, old string, owner bool, count, capacity, minimum int, confirm bool) (bool, string, string, error) {
	next := "CANCELLED"
	if confirm {
		next = "CONFIRMED"
	}
	if !confirm && owner {
		return false, old, "", ErrOwnerCannotWithdraw
	}
	if exists && old == next {
		return false, next, participantReservationStatus(count, minimum), nil
	}
	if !exists && !confirm {
		return false, next, participantReservationStatus(count, minimum), nil
	}
	if confirm && old != "CONFIRMED" && count >= capacity {
		return false, old, "", ErrGroupCapacity
	}
	if confirm {
		count++
	} else if old == "CONFIRMED" {
		count--
	}
	return true, next, participantReservationStatus(count, minimum), nil
}
func participantReservationStatus(count, minimum int) string {
	if count >= minimum {
		return "CONFIRMED"
	}
	return "PENDING"
}

func validateTargetChange(target, minimum, capacity, confirmed int) error {
	if target < minimum || target > capacity {
		return ErrInvalidTargetParticipants
	}
	if target < confirmed {
		return ErrTargetBelowConfirmed
	}
	return nil
}

func targetDeadlineOpen(now, deadline time.Time) bool {
	return !now.After(deadline)
}

func participationDeadlineClosed(now, deadline time.Time) bool {
	return now.After(deadline)
}
