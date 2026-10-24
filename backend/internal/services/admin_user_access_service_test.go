package services

import (
	"errors"
	"strings"
	"testing"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
)

func boolPointer(value bool) *bool { return &value }

func TestValidateUserAccessChange(t *testing.T) {
	admin := models.LocalAuthUser{ID: 3, IsAdmin: true}
	tests := []struct {
		name     string
		actor    models.LocalAuthUser
		target   int
		request  models.AdminUserAccessRequest
		expected error
	}{
		{"bloquear", admin, 5, models.AdminUserAccessRequest{IsBlocked: boolPointer(true), Reason: "  Conducta indebida  "}, nil},
		{"desbloquear", admin, 5, models.AdminUserAccessRequest{IsBlocked: boolPointer(false), Reason: "Se levantó la medida"}, nil},
		{"autobloqueo", admin, 3, models.AdminUserAccessRequest{IsBlocked: boolPointer(true), Reason: "Autobloqueo"}, repositories.ErrUserAccessSelf},
		{"no_admin", models.LocalAuthUser{ID: 3}, 5, models.AdminUserAccessRequest{IsBlocked: boolPointer(true), Reason: "Motivo válido"}, repositories.ErrUserAccessUnauthorized},
		{"admin_bloqueado", models.LocalAuthUser{ID: 3, IsAdmin: true, IsBlocked: true}, 5, models.AdminUserAccessRequest{IsBlocked: boolPointer(true), Reason: "Motivo válido"}, repositories.ErrUserAccessUnauthorized},
		{"sin_estado", admin, 5, models.AdminUserAccessRequest{Reason: "Motivo válido"}, ErrUserAccessInvalidRequest},
		{"sin_motivo", admin, 5, models.AdminUserAccessRequest{IsBlocked: boolPointer(true), Reason: "  "}, ErrUserAccessInvalidReason},
		{"motivo_corto", admin, 5, models.AdminUserAccessRequest{IsBlocked: boolPointer(true), Reason: "abc"}, ErrUserAccessInvalidReason},
		{"motivo_largo", admin, 5, models.AdminUserAccessRequest{IsBlocked: boolPointer(true), Reason: strings.Repeat("x", 501)}, ErrUserAccessInvalidReason},
		{"id_invalido", admin, 0, models.AdminUserAccessRequest{IsBlocked: boolPointer(true), Reason: "Motivo válido"}, repositories.ErrUserAccessNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := validateUserAccessChange(tt.actor, tt.target, tt.request)
			if !errors.Is(err, tt.expected) {
				t.Fatalf("esperado %v; obtenido %v", tt.expected, err)
			}
		})
	}
	blocked, reason, err := validateUserAccessChange(admin, 5, models.AdminUserAccessRequest{IsBlocked: boolPointer(false), Reason: "  Medida revocada  "})
	if err != nil || blocked || reason != "Medida revocada" {
		t.Fatalf("normalización incorrecta: %v %q %v", blocked, reason, err)
	}
}
