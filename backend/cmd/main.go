package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	loadEnv()

	database.Connect()

	defer database.Close()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	routes.RegisterRoutes(app)

	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	log.Println("Servidor iniciado en http://localhost:" + port)

	log.Fatal(app.Listen(":" + port))
}

func loadEnv() {
	_, currentFile, _, ok := runtime.Caller(0)
	var lastErr error

	paths := []string{
		".env",
		"backend/.env",
		"../backend/.env",
		"../.env",
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
			log.Println("No se pudo cargar", path+":", err)
			lastErr = err
		}
	}

	if lastErr == nil {
		lastErr = errors.New("sin detalle de error")
	}

	log.Println("No se pudo cargar archivo .env, usando variables del sistema:", lastErr)
}
