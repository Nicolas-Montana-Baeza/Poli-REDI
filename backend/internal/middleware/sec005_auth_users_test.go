package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"

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
	const tenant = "a1234567-1234-1234-1234-123456789abc"
	const oid = "b1234567-1234-1234-1234-123456789abc"
	const otherOID = "c1234567-1234-1234-1234-123456789abc"
	const originalEmail = "user@example.invalid"

	verifyKey := func(_ *jwt.Token) (interface{}, error) {
		return &key.PublicKey, nil
	}

	tests := []struct {
		name               string
		dbErr              error
		blocked            bool
		mismatchedIdentity bool
		changedEmail       bool
		missingEmail       bool
		missingTenant      bool
		wrongTenant        bool
		missingOID         bool
		invalidOID         bool
		status             int
		message            string
	}{
		{
			name:    "error_consulta_usuario",
			dbErr:   errors.New("SQLSTATE XX999 password=SECRET"),
			status:  500,
			message: "no se pudo cargar tu usuario",
		},
		{
			name:    "vinculacion_pendiente",
			dbErr:   repositories.ErrEntraIdentityLinkRequired,
			status:  403,
			message: "tu cuenta requiere vinculación institucional supervisada",
		},
		{
			name:    "identidad_invalida_en_repositorio",
			dbErr:   repositories.ErrEntraIdentityInvalid,
			status:  403,
			message: "no se pudo validar tu cuenta institucional",
		},
		{
			name:    "usuario_bloqueado",
			blocked: true,
			status:  403,
			message: "usuario bloqueado",
		},
		{
			name:               "identidad_local_distinta",
			mismatchedIdentity: true,
			status:             403,
			message:            "no se pudo validar tu cuenta institucional",
		},
		{
			name:   "autenticacion_exitosa",
			status: 200,
		},
		{
			name:         "correo_cambiado_misma_identidad",
			changedEmail: true,
			status:       200,
		},
		{
			name:         "sin_email",
			missingEmail: true,
			status:       403,
			message:      "no se pudo validar tu cuenta institucional",
		},
		{
			name:        "tenant_no_autorizado",
			wrongTenant: true,
			status:      403,
			message:     "no se pudo validar tu cuenta institucional",
		},
		{
			name:          "sin_tenant",
			missingTenant: true,
			status:        403,
			message:       "no se pudo validar tu cuenta institucional",
		},
		{
			name:       "sin_oid",
			missingOID: true,
			status:     403,
			message:    "no se pudo validar tu cuenta institucional",
		},
		{
			name:       "oid_malformado",
			invalidOID: true,
			status:     403,
			message:    "no se pudo validar tu cuenta institucional",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tokenTenant := tenant
			tokenOID := oid
			tokenEmail := originalEmail

			if tc.wrongTenant {
				tokenTenant = "c1234567-1234-1234-1234-123456789abc"
			}
			if tc.missingTenant {
				tokenTenant = ""
			}
			if tc.missingOID {
				tokenOID = ""
			}
			if tc.invalidOID {
				tokenOID = "invalid-oid"
			}
			if tc.missingEmail {
				tokenEmail = ""
			}
			if tc.changedEmail {
				tokenEmail = "nuevo@example.invalid"
			}

			claims := EntraClaims{
				Name:              "Usuario Prueba",
				PreferredUsername: tokenEmail,
				OID:               tokenOID,
				TID:               tokenTenant,
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:   issuer,
					Audience: jwt.ClaimStrings{audience},
					ExpiresAt: jwt.NewNumericDate(
						time.Now().Add(time.Hour),
					),
				},
			}

			token := jwt.NewWithClaims(
				jwt.SigningMethodRS256, claims,
			)

			signed, err := token.SignedString(key)
			if err != nil {
				t.Fatal(err)
			}

			calls := 0

			users := entraUserOperations{
				resolve: func(
					gotTenant, gotOID, gotEmail, gotName string,
				) (*models.LocalAuthUser, error) {
					calls++

					if gotTenant != tenant ||
						gotOID != oid ||
						gotEmail != tokenEmail ||
						gotName != "Usuario Prueba" {
						t.Fatalf(
							"resolucion incorrecta: %q %q %q %q",
							gotTenant, gotOID, gotEmail, gotName,
						)
					}

					if tc.dbErr != nil {
						return nil, tc.dbErr
					}

					storedOID := oid
					if tc.mismatchedIdentity {
						storedOID = otherOID
					}

					return &models.LocalAuthUser{
						ID:        42,
						Email:     originalEmail,
						FullName:  "Usuario Prueba",
						IsBlocked: tc.blocked,
						OID:       storedOID,
						TenantID:  tenant,
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
					tenant,
					users,
				),
				func(c *fiber.Ctx) error {
					local, localOK := GetLocalUser(c)
					external, externalOK := GetAuthUser(c)

					if !localOK || !externalOK ||
						local.ID != 42 ||
						local.Email != originalEmail ||
						local.OID != oid ||
						local.TenantID != tenant ||
						external.Email != tokenEmail {
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
			req.Header.Set("Authorization", "Bearer "+signed)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.status {
				t.Fatalf(
					"HTTP %d, esperado %d",
					resp.StatusCode, tc.status,
				)
			}

			expectedCalls := 1
			if tc.missingTenant || tc.missingOID ||
				tc.invalidOID || tc.missingEmail || tc.wrongTenant {
				expectedCalls = 0
			}

			if calls != expectedCalls {
				t.Fatalf(
					"consultas: %d, esperado %d",
					calls, expectedCalls,
				)
			}

			var body map[string]json.RawMessage
			if err := json.NewDecoder(resp.Body).Decode(
				&body,
			); err != nil {
				t.Fatal(err)
			}

			if tc.status == 200 {
				var authenticated bool
				if err := json.Unmarshal(
					body["authenticated"],
					&authenticated,
				); err != nil || !authenticated {
					t.Fatalf("autenticacion no confirmada: %v", body)
				}
				return
			}

			if len(body) != 2 {
				t.Fatalf("contrato inseguro: %v", body)
			}

			var publicError, code string
			if err := json.Unmarshal(
				body["error"], &publicError,
			); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(
				body["code"], &code,
			); err != nil {
				t.Fatal(err)
			}

			expectedCode := "FORBIDDEN"
			if tc.status == 500 {
				expectedCode = "INTERNAL_ERROR"
			}

			if publicError != tc.message ||
				code != expectedCode ||
				strings.Contains(publicError, "SECRET") ||
				strings.Contains(publicError, "SQLSTATE") {
				t.Fatalf(
					"respuesta incorrecta o insegura: %v", body,
				)
			}
		})
	}
}
