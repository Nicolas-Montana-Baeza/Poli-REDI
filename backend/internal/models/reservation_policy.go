package models

import "time"

// ReservationPolicy defines the global rules for creating reservations,
// including time windows, frequencies, and resource constraints.
type ReservationPolicy struct {
	ID                          int        `json:"id"`
	ReservableWindowDays        int        `json:"reservableWindowDays"`
	RequestFrequencyDays        int        `json:"requestFrequencyDays"`
	ConfirmationDeadlineMinutes int        `json:"confirmationDeadlineMinutes"`
	MinimumParticipants         int        `json:"minimumParticipants"`
	OpeningMinute               int        `json:"openingMinute"`
	ClosingMinute               int        `json:"closingMinute"`
	SlotIntervalMinutes         int        `json:"slotIntervalMinutes"`
	AllowedDurations            []int      `json:"allowedDurations"`
	ResourceIDs                 []int      `json:"resourceIds"`
	GroupResourceIDs            []int      `json:"groupResourceIds"`
	EffectiveFrom               time.Time  `json:"effectiveFrom"`
	EffectiveTo                 *time.Time `json:"effectiveTo,omitempty"`
	CreatedByUserID             *int       `json:"createdByUserId,omitempty"`
	CreatedAt                   time.Time  `json:"createdAt"`
}

type PublishReservationPolicyRequest struct {
	ReservableWindowDays        int   `json:"reservableWindowDays"`
	RequestFrequencyDays        int   `json:"requestFrequencyDays"`
	ConfirmationDeadlineMinutes int   `json:"confirmationDeadlineMinutes"`
	MinimumParticipants         int   `json:"minimumParticipants"`
	OpeningMinute               int   `json:"openingMinute"`
	ClosingMinute               int   `json:"closingMinute"`
	SlotIntervalMinutes         int   `json:"slotIntervalMinutes"`
	AllowedDurations            []int `json:"allowedDurations"`
	ResourceIDs                 []int `json:"resourceIds"`
	GroupResourceIDs            []int `json:"groupResourceIds"`
}

// CurrentReservationPolicy expone solo condiciones operativas. Identificadores,
// vigencias y autoría pertenecen al historial administrativo.
type CurrentReservationPolicy struct {
	ReservableWindowDays        int   `json:"reservableWindowDays"`
	RequestFrequencyDays        int   `json:"requestFrequencyDays"`
	ConfirmationDeadlineMinutes int   `json:"confirmationDeadlineMinutes"`
	MinimumParticipants         int   `json:"minimumParticipants"`
	OpeningMinute               int   `json:"openingMinute"`
	ClosingMinute               int   `json:"closingMinute"`
	SlotIntervalMinutes         int   `json:"slotIntervalMinutes"`
	AllowedDurations            []int `json:"allowedDurations"`
	ResourceIDs                 []int `json:"resourceIds"`
	GroupResourceIDs            []int `json:"groupResourceIds"`
}

func (p ReservationPolicy) Public() CurrentReservationPolicy {
	return CurrentReservationPolicy{
		ReservableWindowDays: p.ReservableWindowDays, RequestFrequencyDays: p.RequestFrequencyDays,
		ConfirmationDeadlineMinutes: p.ConfirmationDeadlineMinutes, MinimumParticipants: p.MinimumParticipants,
		OpeningMinute: p.OpeningMinute, ClosingMinute: p.ClosingMinute, SlotIntervalMinutes: p.SlotIntervalMinutes,
		AllowedDurations: p.AllowedDurations, ResourceIDs: p.ResourceIDs, GroupResourceIDs: p.GroupResourceIDs,
	}
}
