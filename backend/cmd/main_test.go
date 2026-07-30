package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func TestParseJoinCodeKeyVersion(t *testing.T) {
	if got, err := parseJoinCodeKeyVersion(" 1 "); err != nil || got != 1 {
		t.Fatalf("valid version = %d, %v", got, err)
	}
	for _, invalid := range []string{"", "0", "-1", `"1"`, "1:key", "one"} {
		if _, err := parseJoinCodeKeyVersion(invalid); err == nil {
			t.Fatalf("accepted invalid version %q", invalid)
		}
	}
}

func TestCORSPreflightAllowsIdempotencyKey(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://frontend.example.test")
	app := fiber.New()
	app.Use(cors.New(corsConfig()))
	app.Post("/api/admin/reservation-policies", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusCreated)
	})

	request := httptest.NewRequest(http.MethodOptions, "/api/admin/reservation-policies", nil)
	request.Header.Set("Origin", "https://frontend.example.test")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	request.Header.Set("Access-Control-Request-Headers", "content-type,idempotency-key")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusNoContent {
		t.Fatalf("preflight status = %d", response.StatusCode)
	}
	allowed := strings.ToLower(response.Header.Get("Access-Control-Allow-Headers"))
	if !strings.Contains(allowed, "idempotency-key") {
		t.Fatalf("preflight omitted Idempotency-Key: %q", allowed)
	}
}
