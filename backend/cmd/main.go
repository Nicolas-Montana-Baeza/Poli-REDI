package main

import (
	"log"

	"poli-redi-api/internal/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	routes.RegisterRoutes(app)

	log.Println("Servidor iniciado en http://localhost:3000")

	log.Fatal(app.Listen(":3000"))
}
