package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
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
		log.Fatal("Error creando pool de conexion a SQL Server:", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatal("No se pudo conectar a SQL Server:", err)
	}

	DB = db

	log.Println("Conectado a SQL Server correctamente")
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

	if !strings.EqualFold(encrypt, "true") {
		log.Fatal("DB_ENCRYPT debe ser true. No use conexiones SQL sin cifrado.")
	}

	if strings.EqualFold(trustServerCertificate, "true") && !isLocalServer(server) {
		log.Fatal("DB_TRUST_SERVER_CERTIFICATE=true solo es seguro para conexiones locales (localhost, 127.0.0.1, . o (local)).")
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

func isLocalServer(server string) bool {
	server = strings.TrimSpace(server)
	if server == "" {
		return false
	}

	if strings.Contains(server, "\\") {
		server = strings.SplitN(server, "\\", 2)[0]
	}

	if strings.Contains(server, ",") {
		server = strings.SplitN(server, ",", 2)[0]
	}

	server = strings.TrimSpace(strings.ToLower(server))
	return server == "localhost" || server == "127.0.0.1" || server == "." || server == "(local)"
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
