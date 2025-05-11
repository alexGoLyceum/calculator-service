package migrate

import (
	"errors"
	"fmt"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/database/postgres"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(cfg *postgres.Config, migrationDir string) error {
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationDir),
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("error initializing migrations: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("error running migrations: %v", err)
	}

	return nil
}
