package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"poli-redi-api/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func TestSEC005EntraLocalUserHTTP(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	const issuer = "https://issuer.example.invalid/v2.0"
	const audience = "poliredi-test-api"
	const email = "user@example.invalid"

	claims := EntraClaims{
		Name:              "Usuario Prueba",
		PreferredUsername: email,
		OID:               "test-oid",
		TID:               "test-tenant",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: issuer,
			Audience: jwt.ClaimStrings{
				audience,
			},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}

	verifyKey := func(_ *jwt.Token) (interface{}, error) {
		return &key.PublicKey, nil
	}

	tests := []struct {
		name       string
		lookupFail bool
		blocked    bool
		syncFail   bool
		status     int
		message    string
		wantSync   int
	}{
		{
			name:       "error_consulta_usuario",
			lookupFail: true,
			status:     500,
			message:    "no se pudo cargar tu usuario",
		},
		{
			name:    "usuario_bloqueado",
			blocked: true,
			status:  403,
			message: "usuario bloqueado",
		},
		{
			name:     "error_sincronizacion",
			syncFail: true,
			status:   500,
			message:  "no se pudo sincronizar tu identidad institucional",
			wantSync: 1,
		},
		{
			name:     "autenticacion_exitosa",
			status:   200,
			wantSync: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lookupCalls := 0
			syncCalls := 0

			users := entraUserOperations{
				getOrCreate: func(
					gotEmail, gotName string,
				) (*models.LocalAuthUser, error) {
					lookupCalls++

					if gotEmail != email ||
						gotName != "Usuario Prueba" {
						t.Fatalf(
							"identidad inesperada: %q / %q",
							gotEmail, gotName,
						)
					}

					if tc.lookupFail {
						return nil, errors.New(
							"SQLSTATE XX999 password=SECRET",
						)
					}

					return &models.LocalAuthUser{
						ID:        42,
						Email:     email,
						FullName:  gotName,
						IsBlocked: tc.blocked,
					}, nil
				},
				updateIdentity: func(
					id int, oid, tenant string,
				) (*models.LocalAuthUser, error) {
					syncCalls++

					if id != 42 ||
						oid != "test-oid" ||
						tenant != "test-tenant" {
						t.Fatalf(
							"sincronizacion inesperada: %d %q %q",
							id, oid, tenant,
						)
					}

					if tc.syncFail {
						return nil, errors.New(
							"postgres constraint password=SECRET",
						)
					}

					return &models.LocalAuthUser{
						ID:       42,
						Email:    email,
						FullName: "Usuario Prueba",
						OID:      oid,
						TenantID: tenant,
					}, nil
				},
			}

			app := fiber.New()

			app.Get(
				"/secure",
				requireEntraAuthWithUsers(
					verifyKey,
					issuer,
					[]string{audience},
					users,
				),
				func(c *fiber.Ctx) error {
					local, localOK := GetLocalUser(c)
					external, externalOK := GetAuthUser(c)

					if !localOK || !externalOK ||
						local.ID != 42 ||
						local.OID != "test-oid" ||
						local.TenantID != "test-tenant" ||
						external.Email != email {
						return c.SendStatus(500)
					}

					return c.JSON(fiber.Map{
						"authenticated": true,
					})
				},
			)

			req := httptest.NewRequest(
				http.MethodGet, "/secure", nil,
			)
			req.Header.Set(
				"Authorization", "Bearer "+signed,
			)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			if resp.StatusCode != tc.status {
				t.Fatalf(
					"HTTP %d, esperado %d: %s",
					resp.StatusCode, tc.status, body,
				)
			}

			if lookupCalls != 1 || syncCalls != tc.wantSync {
				t.Fatalf(
					"llamadas: consulta=%d sincronizacion=%d",
					lookupCalls, syncCalls,
				)
			}

			var payload map[string]json.RawMessage
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatal(err)
			}

			if tc.status == 200 {
				var authenticated bool
				if err := json.Unmarshal(
					payload["authenticated"],
					&authenticated,
				); err != nil || !authenticated {
					t.Fatalf("autenticacion no confirmada: %s", body)
				}
				return
			}

			if len(payload) != 1 {
				t.Fatalf("campos inesperados: %s", body)
			}

			var message string
			if err := json.Unmarshal(
				payload["error"], &message,
			); err != nil {
				t.Fatal(err)
			}

			if message != tc.message ||
				strings.Contains(string(body), "SECRET") ||
				strings.Contains(string(body), "SQLSTATE") {
				t.Fatalf("respuesta insegura: %s", body)
			}
		})
	}
}
