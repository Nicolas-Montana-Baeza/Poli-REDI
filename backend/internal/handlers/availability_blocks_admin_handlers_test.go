package handlers

import (
	"testing"
	"time"
)

func TestParseAdminBlockTimes(t *testing.T) {
	cases := []struct {
		name, start, end string
		ok               bool
	}{
		{"valido", "2026-10-21T10:00:00-03:00", "2026-10-21T11:00:00-03:00", true},
		{"offset_obligatorio", "2026-10-21T10:00:00", "2026-10-21T11:00:00", false},
		{"cero", "2026-10-21T10:00:00Z", "2026-10-21T10:00:00Z", false},
		{"inverso", "2026-10-21T12:00:00Z", "2026-10-21T10:00:00Z", false},
		{"excede_maximo", "2026-10-01T10:00:00Z", "2026-11-02T10:00:00Z", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			from, to, err := parseAdminBlockTimes(tc.start, tc.end)
			if (err == nil) != tc.ok {
				t.Fatalf("parse times: %v", err)
			}
			if tc.ok && (from.IsZero() || !to.After(from) || to.Sub(from) != time.Hour) {
				t.Fatalf("rango equivocado: %v %v", from, to)
			}
		})
	}
}

func TestValidAdminBlockType(t *testing.T) {
	for _, valid := range []string{"MAINTENANCE", "ADMINISTRATIVE", "EVENT", "CLOSED", "OTHER"} {
		if !validAdminBlockType(valid) {
			t.Fatal("tipo valido rechazado:", valid)
		}
	}
	for _, invalid := range []string{"", "EVENTUAL", "maintenance"} {
		if validAdminBlockType(invalid) {
			t.Fatal("tipo invalido aceptado:", invalid)
		}
	}
}
