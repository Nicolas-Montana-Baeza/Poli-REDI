package repositories

import (
	"context"
	"database/sql"
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/models"
)

var (
	ErrEntraIdentityInvalid = errors.New(
		"identidad institucional invalida",
	)

	ErrEntraIdentityLinkRequired = errors.New(
		"la cuenta requiere vinculacion institucional supervisada",
	)
)

func sec006ValidGUID(value string) bool {
	if len(value) != 36 {
		return false
	}

	for index := 0; index < len(value); index++ {
		ch := value[index]

		switch index {
		case 8, 13, 18, 23:
			if ch != '-' {
				return false
			}
		default:
			hex := (ch >= '0' && ch <= '9') ||
				(ch >= 'a' && ch <= 'f') ||
				(ch >= 'A' && ch <= 'F')

			if !hex {
				return false
			}
		}
	}

	return true
}

func sec006NormalizeIdentity(
	tenantID, oid string,
) (string, string, error) {
	tenantID = strings.ToLower(strings.TrimSpace(tenantID))
	oid = strings.ToLower(strings.TrimSpace(oid))

	if !sec006ValidGUID(tenantID) ||
		!sec006ValidGUID(oid) {
		return "", "", ErrEntraIdentityInvalid
	}

	return tenantID, oid, nil
}

// NormalizeEntraTenantID valida el tenant institucional.
func NormalizeEntraTenantID(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if !sec006ValidGUID(normalized) {
		return "", ErrEntraIdentityInvalid
	}
	return normalized, nil
}

// NormalizeEntraIdentity valida y normaliza ambos GUID.
func NormalizeEntraIdentity(
	tenantID, oid string,
) (string, string, error) {
	return sec006NormalizeIdentity(tenantID, oid)
}

// ResolveEntraUser usa exclusivamente tenant_id + entra_oid
// como identidad externa.
//
// El correo se utiliza solo al crear una identidad nueva.
// Nunca permite tomar una cuenta historica o ajena.
//
// No modifica email, RUT, permisos ni estado de bloqueo
// de usuarios previamente vinculados.
func ResolveEntraUser(
	tenantID, oid, email, fullName string,
) (*models.LocalAuthUser, error) {
	var err error
	tenantID, oid, err = sec006NormalizeIdentity(
		tenantID, oid,
	)
	if err != nil {
		return nil, err
	}

	email = strings.TrimSpace(email)
	fullName = strings.TrimSpace(fullName)

	address, parseErr := mail.ParseAddress(email)
	if parseErr != nil ||
		address.Address != email ||
		len(email) > 254 {
		return nil, ErrEntraIdentityInvalid
	}

	if fullName == "" {
		fullName = email
	}
	if utf8.RuneCountInString(fullName) > 150 {
		return nil, ErrEntraIdentityInvalid
	}

	ctx := context.Background()

	existing, err := sec006GetUserByIdentity(
		ctx, tenantID, oid,
	)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	var createdID int

	err = database.DB.QueryRowContext(ctx, `
		INSERT INTO users (
			email,
			full_name,
			is_admin,
			is_blocked,
			tenant_id,
			entra_oid
		)
		VALUES ($1, $2, false, false, $3, $4)
		ON CONFLICT DO NOTHING
		RETURNING id
	`, email, fullName, tenantID, oid).Scan(&createdID)

	if err == nil {
		return getUserByID(ctx, createdID)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Una insercion concurrente pudo crear la misma identidad.
	// Solo se recupera por tenant_id + entra_oid, nunca por email.
	existing, err = sec006GetUserByIdentity(
		ctx, tenantID, oid,
	)

	if err == nil {
		return existing, nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		// El correo puede pertenecer a una cuenta historica
		// o a una identidad externa diferente.
		return nil, ErrEntraIdentityLinkRequired
	}

	return nil, err
}

func sec006GetUserByIdentity(
	ctx context.Context,
	tenantID, oid string,
) (*models.LocalAuthUser, error) {
	var user models.LocalAuthUser

	err := database.DB.QueryRowContext(ctx, `
		SELECT
			id,
			email,
			full_name,
			COALESCE(rut, ''),
			is_admin,
			is_blocked,
			COALESCE(entra_oid, ''),
			COALESCE(tenant_id, '')
		FROM users
		WHERE tenant_id = $1
		  AND entra_oid = $2
	`, tenantID, oid).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.RUT,
		&user.IsAdmin,
		&user.IsBlocked,
		&user.OID,
		&user.TenantID,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
