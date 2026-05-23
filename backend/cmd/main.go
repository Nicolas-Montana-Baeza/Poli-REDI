package main

import (
	"log"
	"os"

	"poli-redi-api/internal/database"
	"poli-redi-api/internal/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Println("No se encontró archivo .env, usando variables del sistema")
	}

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
