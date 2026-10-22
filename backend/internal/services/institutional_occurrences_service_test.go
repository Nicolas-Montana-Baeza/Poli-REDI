package services

import (
	"errors"
	"poli-redi-api/internal/repositories"
	"testing"
)

func TestInstitutionalOccurrenceValidation(t *testing.T) {
	date := "2026-11-03"
	start := "10:00"
	end := "11:00"
	tests := []struct {
		name    string
		request InstitutionalOccurrenceRequest
		valid   bool
	}{
		{"cancel semanal válido", InstitutionalOccurrenceRequest{ScheduleID: 1, OriginalDate: date, Action: "CANCEL", Reason: "Cambio de recinto"}, true},
		{"reprogramación válida", InstitutionalOccurrenceRequest{ScheduleID: 1, OriginalDate: date, Action: "RESCHEDULE", Reason: "Cambio de jornada", NewDate: &date, NewStartTime: &start, NewEndTime: &end}, true},
		{"sin motivo", InstitutionalOccurrenceRequest{ScheduleID: 1, OriginalDate: date, Action: "CANCEL"}, false},
		{"fecha incorrecta", InstitutionalOccurrenceRequest{ScheduleID: 1, OriginalDate: "2026-55-03", Action: "CANCEL", Reason: "Sesión suspendida"}, false},
		{"horas faltantes", InstitutionalOccurrenceRequest{ScheduleID: 1, OriginalDate: date, Action: "RESCHEDULE", Reason: "Sesión movida"}, false},
		{"cancel con nueva fecha", InstitutionalOccurrenceRequest{ScheduleID: 1, OriginalDate: date, Action: "CANCEL", Reason: "Sesión suspendida", NewDate: &date}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOccurrenceRequest(&tt.request)
			if tt.valid && err != nil {
				t.Fatal(err)
			}
			if !tt.valid && !errors.Is(err, repositories.ErrInstitutionalOccurrenceInvalid) {
				t.Fatalf("expected invalid request, got %v", err)
			}
		})
	}
}
