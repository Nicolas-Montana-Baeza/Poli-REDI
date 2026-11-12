package handlers

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"
)

// Conserva "error" para mantener compatibilidad con el frontend.
func sec005WritePublicError(c *fiber.Ctx, status int, message string) error {
	var code string

	switch status {
	case fiber.StatusBadRequest:
		code = "VALIDATION_ERROR"
	case fiber.StatusUnauthorized:
		code = "UNAUTHORIZED"
	case fiber.StatusForbidden:
		code = "FORBIDDEN"
	case fiber.StatusNotFound:
		code = "NOT_FOUND"
	case fiber.StatusConflict:
		code = "CONFLICT"
	default:
		code = "INTERNAL_ERROR"
	}

	return c.Status(status).JSON(fiber.Map{
		"error": message,
		"code":  code,
	})
}

// Nunca registra err.Error(), SQL, tokens ni datos del usuario.
// operation debe ser un identificador constante definido en codigo.
func sec005LogInternalFailure(operation string, err error) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		log.Printf(
			"SEC-005 failure operation=%s type=postgres sqlstate=%s",
			operation,
			pgErr.Code,
		)
		return
	}

	log.Printf(
		"SEC-005 failure operation=%s error_type=%T",
		operation,
		err,
	)
}
