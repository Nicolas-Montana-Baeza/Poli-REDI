package middleware

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"

	"poli-redi-api/internal/models"
	"poli-redi-api/internal/repositories"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type AuthUser struct {
	OID    string
	Name   string
	Email  string
	Tenant string
}

type EntraClaims struct {
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	UPN               string `json:"upn"`
	UniqueName        string `json:"unique_name"`
	OID               string `json:"oid"`
	TID               string `json:"tid"`
	Scopes            string `json:"scp"`
	jwt.RegisteredClaims
}

func resolveEmail(claims *EntraClaims) string {
	if claims.PreferredUsername != "" {
		return claims.PreferredUsername
	}

	if claims.Email != "" {
		return claims.Email
	}

	if claims.UPN != "" {
		return claims.UPN
	}

	if claims.UniqueName != "" {
		return claims.UniqueName
	}

	return ""
}

func RequireAuth() fiber.Handler {
	if strings.EqualFold(os.Getenv("DEV_AUTH_ENABLED"), "true") {
		log.Println("Autenticacion local de desarrollo habilitada")
		return requireDevAuth()
	}

	tenantID := os.Getenv("ENTRA_TENANT_ID")
	apiClientID := os.Getenv("ENTRA_API_CLIENT_ID")
	issuer := os.Getenv("ENTRA_ISSUER")

	if tenantID == "" || apiClientID == "" || issuer == "" {
		panic("configuracion de Microsoft Entra ID incompleta")
	}

	log.Println("Autenticacion Microsoft Entra ID habilitada")

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
				"error":  "debes iniciar sesión para continuar",
				"detail": err.Error(),
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
			detail := "token no valido"

			if err != nil {
				detail = err.Error()
			}

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":  "tu sesión expiró o no es válida",
				"detail": detail,
			})
		}

		email := resolveEmail(claims)
		authUser := AuthUser{
			OID:    claims.OID,
			Name:   claims.Name,
			Email:  email,
			Tenant: claims.TID,
		}

		if authUser.Email == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":  "no se pudo validar tu cuenta institucional",
				"detail": "El token no contiene email, preferred_username, upn ni unique_name.",
			})
		}

		fullName := authUser.Name

		if fullName == "" {
			fullName = authUser.Email
		}

		localUser, err := repositories.GetOrCreateUserByEmail(authUser.Email, fullName)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  "no se pudo cargar tu usuario",
				"detail": err.Error(),
			})
		}

		if localUser.IsBlocked {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":  "usuario bloqueado",
				"detail": "El usuario existe, pero está bloqueado en Poli-REDI.",
			})
		}

		localUser, err = repositories.UpdateUserEntraIdentity(
			localUser.ID,
			authUser.OID,
			authUser.Tenant,
		)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  "no se pudo sincronizar tu identidad institucional",
				"detail": err.Error(),
			})
		}

		c.Locals("authUser", authUser)
		c.Locals("localUser", *localUser)

		return c.Next()
	}
}

func requireDevAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		email := strings.TrimSpace(c.Get("X-Dev-Auth-Email"))
		fullName := strings.TrimSpace(c.Get("X-Dev-Auth-Name"))

		if email == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":  "falta iniciar sesión local de prueba",
				"detail": "Modo dev auth activo, pero falta X-Dev-Auth-Email.",
			})
		}

		if fullName == "" {
			fullName = email
		}

		localUser, err := repositories.GetOrCreateUserByEmail(email, fullName)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  "no se pudo cargar el usuario local de prueba",
				"detail": err.Error(),
			})
		}

		if localUser.IsBlocked {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":  "usuario bloqueado",
				"detail": "El usuario existe, pero está bloqueado en Poli-REDI.",
			})
		}

		if !localUser.IsAdmin && strings.EqualFold(c.Get("X-Dev-Reset-Rut"), "true") {
			localUser, err = repositories.UpdateUserRUT(localUser.ID, "")

			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":  "no se pudo preparar el RUT local de prueba",
					"detail": err.Error(),
				})
			}
		}

		authUser := AuthUser{
			OID:    "dev-local",
			Name:   fullName,
			Email:  email,
			Tenant: "dev-local",
		}

		c.Locals("authUser", authUser)
		c.Locals("localUser", *localUser)

		return c.Next()
	}
}

func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := GetLocalUser(c)

		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "usuario no autenticado",
			})
		}

		if !user.IsAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "no tienes permisos para acceder a esta sección",
			})
		}

		return c.Next()
	}
}

func extractBearerToken(c *fiber.Ctx) (string, error) {
	authHeader := c.Get("Authorization")

	if authHeader == "" {
		return "", errors.New("falta la cabecera Authorization")
	}

	parts := strings.SplitN(authHeader, " ", 2)

	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("formato de Authorization inválido")
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

func GetLocalUser(c *fiber.Ctx) (models.LocalAuthUser, bool) {
	user, ok := c.Locals("localUser").(models.LocalAuthUser)
	return user, ok
}
