package main

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// sec005NewApp configura la ultima linea de defensa HTTP.
// El middleware Recover convierte panicos en errores manejables.
func sec005NewApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: sec005GlobalErrorHandler,
	})

	app.Use(recover.New())

	return app
}

// Nunca se utiliza err.Error() como mensaje publico.
func sec005GlobalErrorHandler(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) && fiberErr != nil {
		if fiberErr.Code >= 400 && fiberErr.Code <= 599 {
			status = fiberErr.Code
		}
	}

	code := "INTERNAL_ERROR"
	message := "no se pudo procesar la solicitud"

	if status >= 500 {
		log.Printf(
			"SEC-005 failure operation=global_fiber_error error_type=%T",
			err,
		)
	} else {
		switch status {
		case fiber.StatusBadRequest:
			code = "VALIDATION_ERROR"
			message = "solicitud invalida"
		case fiber.StatusUnauthorized:
			code = "UNAUTHORIZED"
			message = "autenticacion requerida"
		case fiber.StatusForbidden:
			code = "FORBIDDEN"
			message = "acceso denegado"
		case fiber.StatusNotFound:
			code = "NOT_FOUND"
			message = "recurso no encontrado"
		case fiber.StatusMethodNotAllowed:
			code = "METHOD_NOT_ALLOWED"
			message = "metodo no permitido"
		case fiber.StatusConflict:
			code = "CONFLICT"
			message = "conflicto de solicitud"
		default:
			code = "HTTP_ERROR"
			message = "solicitud no valida"
		}
	}

	return c.Status(status).JSON(fiber.Map{
		"error": message,
		"code":  code,
	})
}
