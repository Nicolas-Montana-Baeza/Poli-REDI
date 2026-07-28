package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"poli-redi-api/internal/businessclock"
	"poli-redi-api/internal/database"
	"poli-redi-api/internal/joinsecret"
	"poli-redi-api/internal/repositories"
	"poli-redi-api/internal/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	loadEnv()

	if err := businessclock.Configure(os.Getenv("APP_TIMEZONE")); err != nil {
		log.Fatal(err)
	}
	keyVersion, err := parseJoinCodeKeyVersion(os.Getenv("JOIN_CODE_KEY_VERSION"))
	if err != nil {
		log.Fatal(err)
	}
	if err := joinsecret.Configure(os.Getenv("JOIN_CODE_ENCRYPTION_KEYS"), keyVersion); err != nil {
		log.Fatal(err)
	}

	database.Connect()

	defer database.Close()
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := repositories.ExpirePendingGroupReservations(businessclock.Now()); err != nil {
				log.Println("No se pudieron expirar solicitudes grupales:", err)
			}
		}
	}()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: envOrDefault("CORS_ALLOWED_ORIGINS", "http://localhost:5173"),
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Dev-Auth-Email, X-Dev-Auth-Name, x-dev-auth-email, x-dev-auth-name",
	}))

	routes.RegisterRoutes(app)

	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	log.Println("Servidor iniciado en http://localhost:" + port)

	log.Fatal(app.Listen(":" + port))
}

func parseJoinCodeKeyVersion(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, errors.New("JOIN_CODE_KEY_VERSION es obligatorio; ejecute scripts/configure-join-code-encryption.ps1")
	}
	version, err := strconv.Atoi(value)
	if err != nil || version <= 0 {
		return 0, errors.New("JOIN_CODE_KEY_VERSION debe ser un entero positivo sin prefijo ni comillas")
	}
	return version, nil
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}

func loadEnv() {
	_, currentFile, _, ok := runtime.Caller(0)
	var lastErr error

	paths := []string{
		".env",
		"backend/.env",
		"../backend/.env",
	}

	if ok {
		backendDir := filepath.Dir(filepath.Dir(currentFile))
		paths = append([]string{filepath.Join(backendDir, ".env")}, paths...)
	}

	for _, path := range paths {
		if err := godotenv.Load(path); err == nil {
			log.Println("Variables de entorno cargadas desde", path)
			return
		} else {
			lastErr = err
		}
	}

	if lastErr == nil {
		lastErr = errors.New("sin detalle de error")
	}

	log.Println("No se pudo cargar archivo .env, usando variables del sistema:", lastErr)
}
