//go:build integration

package repositories

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"poli-redi-api/internal/database"
)

func TestSEC006EntraIdentityPostgresIntegration(t *testing.T) {
	if os.Getenv("POLIREDI_3C_ISOLATED") != "1" ||
		os.Getenv("POLIREDI_SEC006_EPHEMERAL") != "1" {
		t.Skip("requiere PostgreSQL 16 efimero SEC-006")
	}

	rawURL := os.Getenv("DATABASE_URL")
	parsed, err := url.Parse(rawURL)
	if err != nil ||
		parsed.Hostname() != "127.0.0.1" ||
		parsed.Port() == "" {
		t.Fatal("se requiere PostgreSQL efimero local")
	}

	if err := database.Connect(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(database.Close)

	ctx := context.Background()
	suffix := fmt.Sprintf(
		"%d-%d", os.Getpid(), time.Now().UnixNano(),
	)

	email := func(kind string) string {
		return fmt.Sprintf(
			"sec006-%s-%s@example.invalid", kind, suffix,
		)
	}

	t.Cleanup(func() {
		_, err := database.DB.ExecContext(ctx,
			`DELETE FROM users WHERE email LIKE $1`,
			"sec006-%-"+suffix+"@example.invalid",
		)
		if err != nil {
			t.Errorf("limpieza SEC-006: %v", err)
		}
	})

	const tenant = "a1234567-1234-1234-1234-123456789abc"
	const otherTenant = "c1234567-1234-1234-1234-123456789abc"
	const oid = "b1234567-1234-1234-1234-123456789abc"
	const otherOID = "d1234567-1234-1234-1234-123456789abc"
	const concurrentOID = "e1234567-1234-1234-1234-123456789abc"

	// 1. Alta sin privilegios.
	first, err := ResolveEntraUser(
		tenant, oid, email("nuevo"), "Usuario SEC006",
	)
	if err != nil {
		t.Fatal(err)
	}

	if first.ID <= 0 || first.IsAdmin ||
		first.IsBlocked ||
		first.OID != oid || first.TenantID != tenant {
		t.Fatalf("alta insegura: %+v", first)
	}

	t.Run("alta_sin_privilegios", func(t *testing.T) {
		t.Log("Alta con tenant/OID y sin privilegios: PASS")
	})

	// 2. Cambio de correo: debe conservar users.id.
	t.Run("correo_cambiado", func(t *testing.T) {
		user, err := ResolveEntraUser(
			tenant, oid, email("nuevo-correo"),
			"Usuario actualizado",
		)
		if err != nil {
			t.Fatal(err)
		}
		if user.ID != first.ID ||
			user.Email != first.Email || user.IsAdmin {
			t.Fatalf("identidad modificada: %+v", user)
		}
	})

	// 3. Otro OID con el mismo correo no puede tomar la cuenta.
	t.Run("correo_reasignado", func(t *testing.T) {
		user, err := ResolveEntraUser(
			tenant, otherOID, first.Email, "Intruso",
		)
		if user != nil ||
			!errors.Is(err, ErrEntraIdentityLinkRequired) {
			t.Fatalf("colision no rechazada: %v", err)
		}
	})

	// 4. Otro tenant tampoco puede reutilizar el correo.
	t.Run("tenant_distinto", func(t *testing.T) {
		user, err := ResolveEntraUser(
			otherTenant, oid, first.Email, "Otro tenant",
		)
		if user != nil ||
			!errors.Is(err, ErrEntraIdentityLinkRequired) {
			t.Fatalf("colision de tenant: %v", err)
		}
	})

	// 5. Cuenta historica privilegiada: no vincular por email.
	t.Run("cuenta_historica_administradora", func(t *testing.T) {
		var legacyID int

		err := database.DB.QueryRowContext(ctx, `
			INSERT INTO users (
				email, full_name, is_admin, is_blocked
			)
			VALUES ($1, 'Administrador historico', true, false)
			RETURNING id
		`, email("historico")).Scan(&legacyID)

		if err != nil {
			t.Fatal(err)
		}

		user, err := ResolveEntraUser(
			tenant, otherOID,
			email("historico"), "Usuario externo",
		)
		if user != nil ||
			!errors.Is(err, ErrEntraIdentityLinkRequired) {
			t.Fatalf("vinculacion no autorizada: %v", err)
		}

		stored, err := getUserByID(ctx, legacyID)
		if err != nil {
			t.Fatal(err)
		}

		if !stored.IsAdmin ||
			stored.OID != "" ||
			stored.TenantID != "" {
			t.Fatalf("cuenta historica alterada: %+v", stored)
		}

		// Ni siquiera presentando el correo del administrador
		// una identidad ya vinculada puede adquirir sus permisos.
		resolved, err := ResolveEntraUser(
			tenant, oid, email("historico"), "Otro nombre",
		)
		if err != nil {
			t.Fatal(err)
		}
		if resolved.ID != first.ID || resolved.IsAdmin {
			t.Fatalf("privilegios transferidos: %+v", resolved)
		}
	})

	// 6. El bloqueo debe conservarse.
	t.Run("usuario_bloqueado", func(t *testing.T) {
		user, err := ResolveEntraUser(
			tenant, otherOID, email("bloqueado"), "Bloqueado",
		)
		if err != nil {
			t.Fatal(err)
		}

		_, err = database.DB.ExecContext(ctx, `
			UPDATE users SET is_blocked = true WHERE id = $1
		`, user.ID)
		if err != nil {
			t.Fatal(err)
		}

		again, err := ResolveEntraUser(
			tenant, otherOID,
			email("bloqueado-nuevo"), "Otro nombre",
		)
		if err != nil {
			t.Fatal(err)
		}

		if again.ID != user.ID || !again.IsBlocked {
			t.Fatalf("bloqueo perdido: %+v", again)
		}
	})

	// 7. Solicitudes simultaneas: una sola cuenta.
	t.Run("concurrencia", func(t *testing.T) {
		const workers = 16
		var wg sync.WaitGroup

		ids := make([]int, workers)
		errs := make([]error, workers)

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				user, err := ResolveEntraUser(
					tenant,
					concurrentOID,
					email("concurrente"),
					"Usuario concurrente",
				)
				errs[index] = err
				if err == nil && user != nil {
					ids[index] = user.ID
				}
			}(i)
		}

		wg.Wait()

		for i := range ids {
			if errs[i] != nil || ids[i] <= 0 ||
				ids[i] != ids[0] {
				t.Fatalf(
					"worker %d: id=%d err=%v",
					i, ids[i], errs[i],
				)
			}
		}

		var count int
		err := database.DB.QueryRowContext(ctx, `
			SELECT count(*)
			FROM users
			WHERE tenant_id = $1 AND entra_oid = $2
		`, tenant, concurrentOID).Scan(&count)

		if err != nil || count != 1 {
			t.Fatalf(
				"cuentas duplicadas: count=%d err=%v",
				count, err,
			)
		}
	})

	// 8. La cuenta original sigue intacta.
	t.Run("identidad_original_preservada", func(t *testing.T) {
		stored, err := getUserByID(ctx, first.ID)
		if err != nil {
			t.Fatal(err)
		}
		if stored.OID != oid ||
			stored.TenantID != tenant ||
			stored.IsAdmin ||
			stored.Email != first.Email {
			t.Fatalf("identidad original alterada: %+v", stored)
		}
	})
}
