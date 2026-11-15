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

	expectedTenantID, tenantErr :=
		repositories.NormalizeEntraTenantID(tenantID)
	if tenantErr != nil {
		panic("configuracion de Microsoft Entra ID invalida")
	}

	log.Println("Autenticacion Microsoft Entra ID habilitada")

	jwksURL := "https://login.microsoftonline.com/" + expectedTenantID + "/discovery/v2.0/keys"

	jwks, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		panic("no se pudo inicializar JWKS de Entra ID")
	}

	expectedAudiences := []string{
		apiClientID,
		"api://" + apiClientID,
	}

	return requireEntraAuth(jwks.Keyfunc, issuer, expectedAudiences, expectedTenantID)
}

// requireEntraAuth permite probar la validacion JWT con claves
// efimeras. RequireAuth sigue utilizando JWKS real en produccion.
type entraUserOperations struct {
	resolve func(
		string, string, string, string,
	) (*models.LocalAuthUser, error)
}

func requireEntraAuth(
	verifyKey jwt.Keyfunc,
	issuer string,
	expectedAudiences []string,
	expectedTenantID string,
) fiber.Handler {
	return requireEntraAuthWithUsers(
		verifyKey,
		issuer,
		expectedAudiences,
		expectedTenantID,
		entraUserOperations{
			resolve: repositories.ResolveEntraUser,
		},
	)
}

func requireEntraAuthWithUsers(
	verifyKey jwt.Keyfunc,
	issuer string,
	expectedAudiences []string,
	expectedTenantID string,

	users entraUserOperations,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString, err := extractBearerToken(c)

		if err != nil {
			return sec005AuthPublicError(c, fiber.StatusUnauthorized, "debes iniciar sesión para continuar")
		}

		claims := &EntraClaims{}

		token, err := jwt.ParseWithClaims(
			tokenString,
			claims,
			verifyKey,
			jwt.WithIssuer(issuer),
			jwt.WithAudience(expectedAudiences...),
		)

		if err != nil || !token.Valid {
			return sec005AuthPublicError(c, fiber.StatusUnauthorized, "tu sesión expiró o no es válida")
		}

		tenantID, oid, identityErr :=
			repositories.NormalizeEntraIdentity(
				claims.TID, claims.OID,
			)

		if identityErr != nil || tenantID != expectedTenantID {
			return sec005AuthPublicError(
				c,
				fiber.StatusForbidden,
				"no se pudo validar tu cuenta institucional",
			)
		}

		email := resolveEmail(claims)

		authUser := AuthUser{
			OID:    oid,
			Name:   claims.Name,
			Email:  email,
			Tenant: tenantID,
		}

		if authUser.Email == "" {
			return sec005AuthPublicError(
				c,
				fiber.StatusForbidden,
				"no se pudo validar tu cuenta institucional",
			)
		}

		fullName := authUser.Name
		if fullName == "" {
			fullName = authUser.Email
		}

		// La identidad externa se resuelve solo por tenant y OID.
		// Nunca se vincula una cuenta existente por correo.
		localUser, err := users.resolve(
			authUser.Tenant,
			authUser.OID,
			authUser.Email,
			fullName,
		)

		if err != nil {
			switch {
			case errors.Is(err, repositories.ErrEntraIdentityInvalid):
				return sec005AuthPublicError(
					c,
					fiber.StatusForbidden,
					"no se pudo validar tu cuenta institucional",
				)
			case errors.Is(err, repositories.ErrEntraIdentityLinkRequired):
				return sec005AuthPublicError(
					c,
					fiber.StatusForbidden,
					"tu cuenta requiere vinculación institucional supervisada",
				)
			default:
				return sec005AuthPublicError(
					c,
					fiber.StatusInternalServerError,
					"no se pudo cargar tu usuario",
				)
			}
		}

		if localUser == nil {
			return sec005AuthPublicError(
				c,
				fiber.StatusInternalServerError,
				"no se pudo cargar tu usuario",
			)
		}

		// Defensa adicional: no autorizar un registro local
		// cuya identidad no coincida con el JWT validado.
		if localUser.ID <= 0 ||
			localUser.TenantID != authUser.Tenant ||
			localUser.OID != authUser.OID {
			return sec005AuthPublicError(
				c,
				fiber.StatusForbidden,
				"no se pudo validar tu cuenta institucional",
			)
		}

		if localUser.IsBlocked {
			return sec005AuthPublicError(
				c,
				fiber.StatusForbidden,
				"usuario bloqueado",
			)
		}

		c.Locals("authUser", authUser)
		c.Locals("localUser", *localUser)

		return c.Next()
	}
}

// sec005AuthPublicError conserva el campo error y agrega
// un codigo estable sin exponer diagnosticos internos.
func sec005AuthPublicError(c *fiber.Ctx, status int, message string) error {
	code := "INTERNAL_ERROR"

	switch status {
	case fiber.StatusUnauthorized:
		code = "UNAUTHORIZED"
	case fiber.StatusForbidden:
		code = "FORBIDDEN"
	}

	return c.Status(status).JSON(fiber.Map{
		"error": message,
		"code":  code,
	})
}

func requireDevAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		email := strings.TrimSpace(c.Get("X-Dev-Auth-Email"))
		fullName := strings.TrimSpace(c.Get("X-Dev-Auth-Name"))

		if email == "" {
			return sec005AuthPublicError(c, fiber.StatusUnauthorized, "falta iniciar sesión local de prueba")
		}

		if fullName == "" {
			fullName = email
		}

		localUser, err := repositories.GetOrCreateUserByEmail(email, fullName)

		if err != nil {
			return sec005AuthPublicError(c, fiber.StatusInternalServerError, "no se pudo cargar el usuario local de prueba")
		}

		if localUser.IsBlocked {
			return sec005AuthPublicError(c, fiber.StatusForbidden, "usuario bloqueado")
		}

		// X-Dev-Reset-Rut existe solo para probar localmente el flujo de RUT
		// obligatorio. Depende de DEV_AUTH_ENABLED y nunca aplica a administradores.
		if !localUser.IsAdmin && strings.EqualFold(c.Get("X-Dev-Reset-Rut"), "true") {
			localUser, err = repositories.UpdateUserRUT(localUser.ID, "")

			if err != nil {
				return sec005AuthPublicError(c, fiber.StatusInternalServerError, "no se pudo preparar el RUT local de prueba")
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
			return sec005AuthPublicError(c, fiber.StatusUnauthorized, "usuario no autenticado")
		}

		if !user.IsAdmin {
			return sec005AuthPublicError(c, fiber.StatusForbidden, "no tienes permisos para acceder a esta sección")
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
