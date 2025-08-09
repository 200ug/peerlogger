package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/rs/zerolog/log"
)

type Database struct {
	*sql.DB
}

type Config struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func New(config Config) (*Database, error) {
	db, err := sql.Open("postgres", config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	log.Info().Msg("Database connection established successfully")
	return &Database{db}, nil
}

func (d *Database) Close() error {
	log.Info().Msg("Closing database connection")
	return d.DB.Close()
}

/*
	golang-migrate commands:

	- up: 		migrate -database $DATABASE_URL -path internal/db/migrations up
	- down: 	migrate -database $DATABASE_URL -path internal/db/migrations down 1
*/

func (d *Database) RunMigrations(migrationsPath string) error {
	// https://github.com/golang-migrate/migrate/tree/master/database/postgres
	driver, err := postgres.WithInstance(d.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	log.Info().Msg("Database migrations completed successfully")
	return nil
}

func (d *Database) HealthCheck(ctx context.Context) error {
	return d.PingContext(ctx)
}
