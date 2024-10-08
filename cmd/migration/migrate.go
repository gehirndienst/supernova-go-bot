package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gehirndienst/supernova-go-bot/internal/botapi"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	direction := flag.String("direction", "up", "Migration direction: up or down")
	migrationPath := flag.String("migration-path", "../../migrations", "Path to migration files from the pwd(!)")
	envFile := flag.String("env-file", "../../.env", "Path to .env file from the pwd(!)")
	flag.Parse()

	err := godotenv.Load(*envFile)
	if err != nil {
		panic(fmt.Sprintf("Error loading .env file: %v", err))
	}

	logger := botapi.GetLogger()

	absMigrationPath, err := filepath.Abs(*migrationPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("Error getting absolute path to migration files")
		os.Exit(1)
	}
	absMigrationPath = fmt.Sprintf("file://%s", absMigrationPath)

	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		dbName,
	)

	db, err := sql.Open(dbName, connStr)
	if err != nil {
		logger.Fatal().Err(err).Msg("Error opening database")
		os.Exit(1)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Fatal().Err(err).Msg("Error creating driver")
		os.Exit(1)
	}

	m, err := migrate.NewWithDatabaseInstance(absMigrationPath, dbName, driver)
	if err != nil {
		logger.Fatal().Err(err).Msg("Error creating migration instance")
		os.Exit(1)
	}

	switch *direction {
	case "up":
		logger.Info().Msg("Running migration UP")
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			logger.Fatal().Err(err).Msg("Error running migration UP")
			os.Exit(1)
		}
		logger.Info().Msg("Migration UP completed successfully")

	case "down":
		logger.Info().Msg("Running migration DOWN")
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			logger.Fatal().Err(err).Msg("Error running migration DOWN")
			os.Exit(1)
		}
		logger.Info().Msg("Migration DOWN completed successfully")

	default:
		logger.Fatal().Msg("Invalid migration direction. Use 'up' or 'down'.")
		os.Exit(1)
	}
}
