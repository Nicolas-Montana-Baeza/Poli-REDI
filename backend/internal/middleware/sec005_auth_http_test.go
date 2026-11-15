package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"poli-redi-api/internal/database"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func TestSEC005EntraAuthHTTPResponses(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	const issuer = "https://issuer.example.invalid/v2.0"
	const audience = "poliredi-test-api"

	signToken := func(
		key *rsa.PrivateKey,
		tokenIssuer, tokenAudience, email string,
		expiresAt time.Time,
		tokenTenants ...string,
	) string {
		t.Helper()

		tenantID := "a1234567-1234-1234-1234-123456789abc"
		if len(tokenTenants) > 0 {
			tenantID = tokenTenants[0]
		}

		claims := EntraClaims{
			PreferredUsername: email,
			OID:               "b1234567-1234-1234-1234-123456789abc",
			TID:               tenantID,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer: issuer,
				Audience: jwt.ClaimStrings{
					tokenAudience,
				},
				ExpiresAt: jwt.NewNumericDate(expiresAt),
			},
		}
		claims.Issuer = tokenIssuer

		token := jwt.NewWithClaims(
			jwt.SigningMethodRS256,
			claims,
		)

		signed, err := token.SignedString(key)
		if err != nil {
			t.Fatal(err)
		}
		return signed
	}

	validUntil := time.Now().Add(time.Hour)

	validToken := signToken(
		privateKey, issuer, audience,
		"user@example.invalid", validUntil,
	)

	noEmailToken := signToken(
		privateKey, issuer, audience,
		"", validUntil,
	)

	wrongSignature := signToken(
		otherKey, issuer, audience,
		"user@example.invalid", validUntil,
	)

	wrongIssuer := signToken(
		privateKey, "https://other.example.invalid",
		audience, "user@example.invalid", validUntil,
	)

	wrongAudience := signToken(
		privateKey, issuer, "other-api",
		"user@example.invalid", validUntil,
	)

	wrongTenant := signToken(
		privateKey, issuer, audience,
		"user@example.invalid", validUntil,
		"c1234567-1234-1234-1234-123456789abc",
	)

	expiredToken := signToken(
		privateKey, issuer, audience,
		"user@example.invalid",
		time.Now().Add(-time.Hour),
	)

	// Conexion local cerrada: no hay conexiones de red,
	// consultas reales ni modificaciones de datos.
	previousDB := database.DB

	closedDB, err := sql.Open(
		"pgx",
		"postgres://test:test@127.0.0.1:1/test?sslmode=disable",
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := closedDB.Close(); err != nil {
		t.Fatal(err)
	}

	database.DB = closedDB
	t.Cleanup(func() {
		database.DB = previousDB
	})

	verifyKey := func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	}

	app := fiber.New()

	app.Get(
		"/secure",
		requireEntraAuth(
			verifyKey,
			issuer,
			[]string{audience},
			"a1234567-1234-1234-1234-123456789abc",
		),
		func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		},
	)

	tests := []struct {
		name       string
		header     string
		wantStatus int
		wantError  string
	}{
		{
			"sin_token", "",
			401, "debes iniciar sesión para continuar",
		},
		{
			"cabecera_incorrecta", "Basic abc",
			401, "debes iniciar sesión para continuar",
		},
		{
			"jwt_malformado", "Bearer abc.def.ghi",
			401, "tu sesión expiró o no es válida",
		},
		{
			"firma_invalida", "Bearer " + wrongSignature,
			401, "tu sesión expiró o no es válida",
		},
		{
			"emisor_invalido", "Bearer " + wrongIssuer,
			401, "tu sesión expiró o no es válida",
		},
		{
			"audiencia_invalida", "Bearer " + wrongAudience,
			401, "tu sesión expiró o no es válida",
		},
		{
			"tenant_no_autorizado", "Bearer " + wrongTenant,
			403, "no se pudo validar tu cuenta institucional",
		},
		{
			"jwt_expirado", "Bearer " + expiredToken,
			401, "tu sesión expiró o no es válida",
		},
		{
			"sin_email", "Bearer " + noEmailToken,
			403, "no se pudo validar tu cuenta institucional",
		},
		{
			"repositorio_no_disponible", "Bearer " + validToken,
			500, "no se pudo cargar tu usuario",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(
				http.MethodGet,
				"/secure",
				nil,
			)

			if tc.header != "" {
				request.Header.Set(
					"Authorization",
					tc.header,
				)
			}

			response, err := app.Test(request)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}

			if response.StatusCode != tc.wantStatus {
				t.Fatalf(
					"HTTP %d, esperado %d: %s",
					response.StatusCode,
					tc.wantStatus,
					body,
				)
			}

			var payload map[string]json.RawMessage
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("JSON invalido: %v", err)
			}

			if len(payload) != 2 {
				t.Fatalf(
					"Respuesta con campos inesperados: %s",
					body,
				)
			}

			var code string
			if err := json.Unmarshal(payload["code"], &code); err != nil {
				t.Fatalf("codigo de error invalido: %v", err)
			}

			expectedCode := "INTERNAL_ERROR"
			if tc.wantStatus == fiber.StatusUnauthorized {
				expectedCode = "UNAUTHORIZED"
			} else if tc.wantStatus == fiber.StatusForbidden {
				expectedCode = "FORBIDDEN"
			}

			if code != expectedCode {
				t.Fatalf("codigo %q, esperado %q", code, expectedCode)
			}

			var publicError string
			if err := json.Unmarshal(
				payload["error"], &publicError,
			); err != nil {
				t.Fatalf("Campo error invalido: %v", err)
			}

			if publicError != tc.wantError {
				t.Fatalf(
					"Error publico %q, esperado %q",
					publicError,
					tc.wantError,
				)
			}
		})
	}
}
