package repositories

import (
	"errors"
	"testing"
)

func TestSEC006NormalizeIdentity(t *testing.T) {
	tenant := "A1234567-1234-1234-1234-123456789ABC"
	oid := "B1234567-1234-1234-1234-123456789ABC"

	gotTenant, gotOID, err := sec006NormalizeIdentity(
		" "+tenant+" ",
		" "+oid+" ",
	)
	if err != nil {
		t.Fatal(err)
	}

	if gotTenant !=
		"a1234567-1234-1234-1234-123456789abc" ||
		gotOID !=
			"b1234567-1234-1234-1234-123456789abc" {
		t.Fatalf(
			"identidad normalizada incorrectamente: %q %q",
			gotTenant, gotOID,
		)
	}
}

func TestSEC006RejectsInvalidIdentity(t *testing.T) {
	valid := "a1234567-1234-1234-1234-123456789abc"

	tests := []struct {
		name   string
		tenant string
		oid    string
	}{
		{"sin_tenant", "", valid},
		{"sin_oid", valid, ""},
		{"tenant_invalido", "otro-tenant", valid},
		{"oid_invalido", valid, "otro-oid"},
		{"oid_con_caracteres_invalidos",
			valid, "z1234567-1234-1234-1234-123456789abc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := sec006NormalizeIdentity(
				tc.tenant, tc.oid,
			)

			if !errors.Is(
				err, ErrEntraIdentityInvalid,
			) {
				t.Fatalf(
					"se esperaba identidad invalida: %v",
					err,
				)
			}
		})
	}
}

func TestSEC006ResolveRejectsInvalidIDsBeforeDatabase(
	t *testing.T,
) {
	_, err := ResolveEntraUser(
		"tenant-invalido",
		"oid-invalido",
		"user@example.invalid",
		"Usuario",
	)

	if !errors.Is(err, ErrEntraIdentityInvalid) {
		t.Fatalf(
			"se esperaba rechazo antes de consultar BD: %v",
			err,
		)
	}
}
