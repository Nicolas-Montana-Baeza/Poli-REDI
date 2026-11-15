package middleware

import (
	"fmt"
	"strings"
	"testing"
)

func TestSEC006RejectsInvalidTenantConfig(t *testing.T) {
	t.Setenv("DEV_AUTH_ENABLED", "false")
	t.Setenv("ENTRA_TENANT_ID", "common")
	t.Setenv("ENTRA_API_CLIENT_ID", "api-test")
	t.Setenv(
		"ENTRA_ISSUER",
		"https://issuer.example.invalid/v2.0",
	)

	defer func() {
		value := recover()
		if value == nil {
			t.Fatal("se esperaba rechazo del tenant")
		}
		if !strings.Contains(
			fmt.Sprint(value),
			"configuracion de Microsoft Entra ID invalida",
		) {
			t.Fatalf("rechazo inesperado: %v", value)
		}
	}()

	RequireAuth()
}
