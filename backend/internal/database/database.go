package database

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Connect() {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		log.Fatal("DATABASE_URL no está definido en .env")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)

	if err != nil {
		log.Fatal("Error conectando a PostgreSQL:", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("No se pudo hacer ping a PostgreSQL:", err)
	}

	DB = pool

	log.Println("Conectado a PostgreSQL correctamente")
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
