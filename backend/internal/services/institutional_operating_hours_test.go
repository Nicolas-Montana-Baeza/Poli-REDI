package services

import (
	"errors"
	"testing"

	"poli-redi-api/internal/models"
)

func TestInstitutionalScheduleOperatingHours(t *testing.T) {
	tests := []struct {
		name       string
		start, end string
		wantErr    bool
	}{
		{"SINGLE al límite", "08:00", "22:00", false},
		{"WEEKLY interior", "10:00", "11:00", false},
		{"apertura anticipada", "06:00", "11:00", true},
		{"cierre tardío", "10:00", "23:00", true},
		{"ambos extremos inválidos", "06:00", "23:00", true},
		{"fuera del día", "22:00", "23:00", true},
		{"rango invertido", "11:00", "10:00", true},
		{"hora inválida", "25:00", "26:00", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedules := []models.CreateInstitutionalScheduleRequest{{
				ScheduleType: models.InstitutionalScheduleTypeWeekly,
				StartTime:    tt.start, EndTime: tt.end,
			}}
			err := validateInstitutionalOperatingHours(schedules, 8*60, 22*60)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error=%v; wantErr=%v", err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, ErrInstitutionalScheduleInvalid) {
				t.Fatalf("el error debe clasificar como horario inválido: %v", err)
			}
		})
	}
}
