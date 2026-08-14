package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const driverName = "pgx"

var DB *sql.DB

func Connect() error {
	connectionString, err := connectionStringFromEnv()
	if err != nil {
		return err
	}

	db, err := sql.Open(driverName, connectionString)
	if err != nil {
		return fmt.Errorf("crear pool PostgreSQL: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("conectar a PostgreSQL: %w", err)
	}

	DB = db
	return nil
}

func connectionStringFromEnv() (string, error) {
	if rawURL := strings.TrimSpace(os.Getenv("DATABASE_URL")); rawURL != "" {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			return "", fmt.Errorf("DATABASE_URL invalida: %w", err)
		}
		if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
			return "", fmt.Errorf("DATABASE_URL debe usar el esquema postgres o postgresql")
		}
		if parsed.Hostname() == "" || strings.Trim(parsed.Path, "/") == "" {
			return "", fmt.Errorf("DATABASE_URL debe incluir host y base de datos")
		}
		if parsed.User == nil {
			return "", fmt.Errorf("DATABASE_URL debe incluir usuario")
		}
		if _, present := parsed.User.Password(); !present {
			return "", fmt.Errorf("DATABASE_URL debe incluir contrasena")
		}
		return rawURL, nil
	}

	host := envOrDefault("PGHOST", "127.0.0.1")
	port := envOrDefault("PGPORT", "5432")
	databaseName := envOrDefault("PGDATABASE", "poliredi")
	user := envOrDefault("PGUSER", "poliredi_app")
	password := strings.TrimSpace(os.Getenv("PGPASSWORD"))
	sslMode := envOrDefault("PGSSLMODE", "disable")

	parsedPort, err := strconv.Atoi(port)
	if err != nil || parsedPort < 1 || parsedPort > 65535 {
		return "", fmt.Errorf("PGPORT debe ser numerico y estar entre 1 y 65535")
	}
	if password == "" {
		return "", fmt.Errorf("PGPASSWORD no esta definido")
	}
	if !validSSLMode(sslMode) {
		return "", fmt.Errorf("PGSSLMODE no es valido")
	}

	connectionURL := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     net.JoinHostPort(host, port),
		Path:     databaseName,
		RawQuery: url.Values{"sslmode": []string{sslMode}}.Encode(),
	}

	return connectionURL.String(), nil
}

func validSSLMode(value string) bool {
	switch value {
	case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
		return true
	default:
		return false
	}
}

func envOrDefault(name string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func Close() {
	if DB != nil {
		_ = DB.Close()
		DB = nil
	}
}
