package middleware

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type AuthUser struct {
	OID    string
	Name   string
	Email  string
	Tenant string
	RawJWT string
}

type EntraClaims struct {
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	OID               string `json:"oid"`
	TID               string `json:"tid"`
	Scopes            string `json:"scp"`
	jwt.RegisteredClaims
}

func RequireAuth() fiber.Handler {
	tenantID := os.Getenv("ENTRA_TENANT_ID")
	apiClientID := os.Getenv("ENTRA_API_CLIENT_ID")
	issuer := os.Getenv("ENTRA_ISSUER")
	println("ENTRA_TENANT_ID:", tenantID)
	println("ENTRA_API_CLIENT_ID:", apiClientID)
	println("ENTRA_ISSUER:", issuer)

	jwksURL := "https://login.microsoftonline.com/" + tenantID + "/discovery/v2.0/keys"

	jwks, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		panic("no se pudo inicializar JWKS de Entra ID: " + err.Error())
	}

	expectedAudiences := []string{
		apiClientID,
		"api://" + apiClientID,
	}

	return func(c *fiber.Ctx) error {
		tokenString, err := extractBearerToken(c)

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		claims := &EntraClaims{}

		token, err := jwt.ParseWithClaims(
			tokenString,
			claims,
			jwks.Keyfunc,
			jwt.WithIssuer(issuer),
			jwt.WithAudience(expectedAudiences...),
		)

		if err != nil || !token.Valid {
			detail := "token no válido"

			if err != nil {
				detail = err.Error()
			}

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":  "token inválido o expirado",
				"detail": detail,
			})
		}

		authUser := AuthUser{
			OID:    claims.OID,
			Name:   claims.Name,
			Email:  claims.PreferredUsername,
			Tenant: claims.TID,
			RawJWT: tokenString,
		}

		c.Locals("authUser", authUser)

		return c.Next()
	}
}

func extractBearerToken(c *fiber.Ctx) (string, error) {
	authHeader := c.Get("Authorization")

	if authHeader == "" {
		return "", errors.New("falta header Authorization")
	}

	parts := strings.SplitN(authHeader, " ", 2)

	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("formato Authorization inválido")
	}

	return strings.TrimSpace(parts[1]), nil
}

func GetAuthUser(c *fiber.Ctx) (AuthUser, bool) {
	user, ok := c.Locals("authUser").(AuthUser)
	return user, ok
}

func ContextWithAuthUser(ctx context.Context, user AuthUser) context.Context {
	return context.WithValue(ctx, "authUser", user)
}
