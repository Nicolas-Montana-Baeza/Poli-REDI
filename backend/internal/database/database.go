package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

var DB *sql.DB

func Connect() {
	connString := os.Getenv("AZURE_SQL_CONNECTION_STRING")

	if connString == "" {
		connString = buildConnectionString()
	}

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal("Error creando pool de conexion a Azure SQL:", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatal("No se pudo conectar a Azure SQL Database:", err)
	}

	DB = db

	log.Println("Conectado a Azure SQL Database correctamente")
}

func buildConnectionString() string {
	server := requiredEnv("DB_SERVER")
	user := requiredEnv("DB_USER")
	password := requiredEnv("DB_PASSWORD")
	databaseName := requiredEnv("DB_NAME")
	port := envOrDefault("DB_PORT", "1433")
	encrypt := envOrDefault("DB_ENCRYPT", "true")
	trustServerCertificate := envOrDefault("DB_TRUST_SERVER_CERTIFICATE", "false")

	if _, err := strconv.Atoi(port); err != nil {
		log.Fatal("DB_PORT debe ser numerico")
	}

	return fmt.Sprintf(
		"server=%s;user id=%s;password=%s;port=%s;database=%s;encrypt=%s;trustservercertificate=%s;",
		server,
		user,
		password,
		port,
		databaseName,
		encrypt,
		trustServerCertificate,
	)
}

func requiredEnv(name string) string {
	value := os.Getenv(name)

	if value == "" {
		log.Fatalf("%s no esta definido en .env", name)
	}

	return value
}

func envOrDefault(name string, fallback string) string {
	value := os.Getenv(name)

	if value == "" {
		return fallback
	}

	return value
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
