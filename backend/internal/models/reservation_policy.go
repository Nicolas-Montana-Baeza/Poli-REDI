package models

import "time"

type ReservationPolicy struct {
	ID                          int
	ReservableWindowDays        int
	RequestFrequencyDays        int
	ConfirmationDeadlineMinutes int
	MinimumParticipants         int
	EffectiveFrom               time.Time
	EffectiveTo                 *time.Time
}
