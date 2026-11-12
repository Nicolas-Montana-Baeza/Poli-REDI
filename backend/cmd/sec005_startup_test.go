package main

import (
	"errors"
	"strings"
	"testing"
)

func TestSEC005StartupLogHidesSecrets(t *testing.T) {
	err := errors.New(
		"postgres://app:SECRET_PASSWORD@db.invalid/poliredi",
	)

	message := sec005StartupErrorMessage("postgres", err)

	if strings.Contains(message, "SECRET_PASSWORD") ||
		strings.Contains(message, "postgres://") {
		t.Fatalf("registro inseguro: %q", message)
	}

	if !strings.Contains(message, "component=postgres") ||
		!strings.Contains(message, "error_type=") {
		t.Fatalf("diagnostico insuficiente: %q", message)
	}
}
