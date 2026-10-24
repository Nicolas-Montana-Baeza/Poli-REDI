package services

import (
	"errors"
	"strings"
	"testing"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"
)

func boolPtrRole(v bool) *bool { return &v }

func TestValidateUserRoleChange(t *testing.T) {
	admin := models.LocalAuthUser{ID: 3, IsAdmin: true}
	cases := []struct {
		name     string
		actor    models.LocalAuthUser
		id       int
		req      models.AdminUserRoleRequest
		expected error
	}{
		{"promover", admin, 5, models.AdminUserRoleRequest{IsAdmin: boolPtrRole(true), Reason: "  Autorización institucional  "}, nil},
		{"revocar", admin, 5, models.AdminUserRoleRequest{IsAdmin: boolPtrRole(false), Reason: "Revisión de permisos"}, nil},
		{"self", admin, 3, models.AdminUserRoleRequest{IsAdmin: boolPtrRole(false), Reason: "Motivo válido"}, repositories.ErrUserRoleSelf},
		{"no_admin", models.LocalAuthUser{ID: 3}, 5, models.AdminUserRoleRequest{IsAdmin: boolPtrRole(true), Reason: "Motivo válido"}, repositories.ErrUserRoleUnauthorized},
		{"admin_bloqueado", models.LocalAuthUser{ID: 3, IsAdmin: true, IsBlocked: true}, 5, models.AdminUserRoleRequest{IsAdmin: boolPtrRole(true), Reason: "Motivo válido"}, repositories.ErrUserRoleUnauthorized},
		{"sin_estado", admin, 5, models.AdminUserRoleRequest{Reason: "Motivo válido"}, ErrUserRoleInvalidRequest},
		{"sin_motivo", admin, 5, models.AdminUserRoleRequest{IsAdmin: boolPtrRole(true), Reason: " "}, ErrUserRoleInvalidReason},
		{"corto", admin, 5, models.AdminUserRoleRequest{IsAdmin: boolPtrRole(true), Reason: "abc"}, ErrUserRoleInvalidReason},
		{"largo", admin, 5, models.AdminUserRoleRequest{IsAdmin: boolPtrRole(true), Reason: strings.Repeat("a", 501)}, ErrUserRoleInvalidReason},
		{"id_invalido", admin, 0, models.AdminUserRoleRequest{IsAdmin: boolPtrRole(true), Reason: "Motivo válido"}, repositories.ErrUserRoleNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := validateUserRoleChange(tc.actor, tc.id, tc.req)
			if !errors.Is(err, tc.expected) {
				t.Fatalf("esperaba %v, obtuvo %v", tc.expected, err)
			}
		})
	}
	granted, why, err := validateUserRoleChange(admin, 5,
		models.AdminUserRoleRequest{IsAdmin: boolPtrRole(true), Reason: "  Acceso confirmado  "})
	if err != nil || !granted || why != "Acceso confirmado" {
		t.Fatal("normalizacion incorrecta", err)
	}
}
