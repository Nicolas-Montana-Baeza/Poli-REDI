package repositories

import (
	"database/sql"
	"strings"
	"testing"
)

func ptrOccurrence(s string) *string { return &s }

func TestInstitutionalOccurrenceNotificationSpec(t *testing.T) {
	date, start, end := "2026-11-04", "10:30", "11:30"
	cancel := institutionalOccurrenceNotificationSpec("Spinning", "2026-11-03", "2026-11-03", "CANCEL", nil, nil, nil)
	if cancel.Type != "INSTITUTIONAL_SESSION_CANCELLED" || !strings.Contains(cancel.Message, "suspendida") || !strings.Contains(cancel.Message, "inscripción") {
		t.Fatalf("unexpected cancel notification: %+v", cancel)
	}
	moved := institutionalOccurrenceNotificationSpec("Spinning", "2026-11-03", "2026-11-03", "RESCHEDULE", &date, &start, &end)
	if moved.Type != "INSTITUTIONAL_SESSION_RESCHEDULED" || !strings.Contains(moved.Message, date) || !strings.Contains(moved.Message, start) {
		t.Fatalf("unexpected reschedule notification: %+v", moved)
	}
	if strings.Contains(moved.Message, "motivo de gestión interna") {
		t.Fatal("internal administrative reason leaked to participant")
	}
}

func TestInstitutionalOccurrenceNoopDoesNotRenotify(t *testing.T) {
	date, start, end := "2026-11-04", "10:30", "11:30"
	p := InstitutionalOccurrencePatch{Action: "RESCHEDULE", NewDate: &date, NewStartTime: &start, NewEndTime: &end}
	if !occurrencePatchUnchanged(true, "RESCHEDULE", sql.NullString{String: date, Valid: true}, sql.NullString{String: start, Valid: true}, sql.NullString{String: end, Valid: true}, p) {
		t.Fatal("same effective schedule should be a no-op")
	}
	if occurrencePatchUnchanged(false, "RESCHEDULE", sql.NullString{}, sql.NullString{}, sql.NullString{}, p) {
		t.Fatal("first patch cannot be treated as existing")
	}
	if occurrencePatchUnchanged(true, "CANCEL", sql.NullString{}, sql.NullString{}, sql.NullString{}, p) {
		t.Fatal("reactivating a previously cancelled occurrence is a real change")
	}
	if !occurrencePatchUnchanged(true, "CANCEL", sql.NullString{}, sql.NullString{}, sql.NullString{}, InstitutionalOccurrencePatch{Action: "CANCEL"}) {
		t.Fatal("duplicate cancellation must not send a second notification")
	}
	p.NewDate = ptrOccurrence("2026-11-05")
	if occurrencePatchUnchanged(true, "RESCHEDULE", sql.NullString{String: date, Valid: true}, sql.NullString{String: start, Valid: true}, sql.NullString{String: end, Valid: true}, p) {
		t.Fatal("moving again to a different date is a real change")
	}
}
