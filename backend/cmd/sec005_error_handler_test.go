package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestSEC005GlobalErrorHandler(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		handler fiber.Handler
		status  int
		code    string
		message string
	}{
		{
			name: "error_SQL_inesperado",
			path: "/test",
			handler: func(c *fiber.Ctx) error {
				return errors.New("SQLSTATE XX999 password=SECRET")
			},
			status:  500,
			code:    "INTERNAL_ERROR",
			message: "no se pudo procesar la solicitud",
		},
		{
			name: "error_Fiber_500",
			path: "/test",
			handler: func(c *fiber.Ctx) error {
				return fiber.NewError(
					500, "JWT SECRET_INTERNAL_MESSAGE",
				)
			},
			status:  500,
			code:    "INTERNAL_ERROR",
			message: "no se pudo procesar la solicitud",
		},
		{
			name: "error_Fiber_503",
			path: "/test",
			handler: func(c *fiber.Ctx) error {
				return fiber.NewError(
					503, "database SECRET unavailable",
				)
			},
			status:  503,
			code:    "INTERNAL_ERROR",
			message: "no se pudo procesar la solicitud",
		},
		{
			name: "validacion_envuelta",
			path: "/test",
			handler: func(c *fiber.Ctx) error {
				return fmt.Errorf(
					"SQLSTATE SECRET: %w", fiber.ErrBadRequest,
				)
			},
			status:  400,
			code:    "VALIDATION_ERROR",
			message: "solicitud invalida",
		},
		{
			name: "acceso_denegado",
			path: "/test",
			handler: func(c *fiber.Ctx) error {
				return fiber.NewError(
					403, "internal SECRET permissions",
				)
			},
			status:  403,
			code:    "FORBIDDEN",
			message: "acceso denegado",
		},
		{
			name: "metodo_no_permitido",
			path: "/test",
			handler: func(c *fiber.Ctx) error {
				return fiber.NewError(
					405, "internal SECRET method",
				)
			},
			status:  405,
			code:    "METHOD_NOT_ALLOWED",
			message: "metodo no permitido",
		},
		{
			name: "panico_recuperado",
			path: "/test",
			handler: func(c *fiber.Ctx) error {
				panic("SECRET_PANIC_DATABASE")
			},
			status:  500,
			code:    "INTERNAL_ERROR",
			message: "no se pudo procesar la solicitud",
		},
		{
			name:    "ruta_inexistente",
			path:    "/not-found",
			status:  404,
			code:    "NOT_FOUND",
			message: "recurso no encontrado",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := sec005NewApp()

			if tc.handler != nil {
				app.Get("/test", tc.handler)
			}

			req := httptest.NewRequest(
				http.MethodGet, tc.path, nil,
			)

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

			var payload map[string]string

			if err := json.NewDecoder(resp.Body).Decode(
				&payload,
			); err != nil {
				t.Fatal(err)
			}

			if len(payload) != 2 ||
				payload["error"] != tc.message ||
				payload["code"] != tc.code {
				t.Fatalf(
					"contrato HTTP inesperado: %v", payload,
				)
			}

			for _, value := range payload {
				if strings.Contains(value, "SECRET") ||
					strings.Contains(value, "SQLSTATE") {
					t.Fatalf(
						"detalle interno expuesto: %v", payload,
					)
				}
			}
		})
	}
}
