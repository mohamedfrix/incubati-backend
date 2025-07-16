package migrations

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/moulaybdl/incubAT/project_service/internal/logger"
)

// Migration represents a single database migration
type Migration struct {
	Version string
	Name    string
	UpSQL   string
	DownSQL string
}

// Migrator handles database migrations
type Migrator struct {
	db          *sql.DB
	migrations  []Migration
	tableName   string
}

// NewMigrator creates a new migration instance
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{
		db:        db,
		tableName: "schema_migrations",
	}
}

// LoadMigrationsFromDir loads migrations from the migrations directory
func (m *Migrator) LoadMigrationsFromDir(migrationsDir string) error {
	logger.Logger.Info("Loading migrations from directory", "dir", migrationsDir)
	
	// Read migration files
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	sort.Strings(files)

	for _, upFile := range files {
		// Extract version and name from filename
		// Expected format: 000001_create_tables.up.sql
		filename := filepath.Base(upFile)
		parts := strings.Split(filename, "_")
		if len(parts) < 2 {
			logger.Logger.Warn("Skipping migration file with invalid format", "file", filename)
			continue
		}

		version := parts[0]
		name := strings.Join(parts[1:], "_")
		name = strings.TrimSuffix(name, ".up.sql")

		// Read up migration
		upSQL, err := m.readFile(upFile)
		if err != nil {
			return fmt.Errorf("failed to read up migration %s: %w", upFile, err)
		}

		// Read corresponding down migration
		downFile := strings.Replace(upFile, ".up.sql", ".down.sql", 1)
		downSQL, err := m.readFile(downFile)
		if err != nil {
			logger.Logger.Warn("No down migration found", "file", downFile)
			downSQL = "" // Down migration is optional
		}

		migration := Migration{
			Version: version,
			Name:    name,
			UpSQL:   upSQL,
			DownSQL: downSQL,
		}

		m.migrations = append(m.migrations, migration)
		logger.Logger.Debug("Loaded migration", "version", version, "name", name)
	}

	logger.Logger.Info("Loaded migrations", "count", len(m.migrations))
	return nil
}

// readFile reads the content of a file
func (m *Migrator) readFile(filepath string) (string, error) {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}



// createMigrationsTable creates the schema_migrations table if it doesn't exist
func (m *Migrator) createMigrationsTable() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`, m.tableName)

	_, err := m.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	logger.Logger.Debug("Migrations table created or verified")
	return nil
}

// getAppliedMigrations returns a map of applied migration versions
func (m *Migrator) getAppliedMigrations() (map[string]bool, error) {
	query := fmt.Sprintf("SELECT version FROM %s", m.tableName)
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	return applied, nil
}

// recordMigration records a migration as applied
func (m *Migrator) recordMigration(migration Migration) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (version, name, applied_at) 
		VALUES ($1, $2, $3)
	`, m.tableName)

	_, err := m.db.Exec(query, migration.Version, migration.Name, time.Now())
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}

// Up runs all pending migrations
func (m *Migrator) Up() error {
	logger.Logger.Info("Starting database migrations")

	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(); err != nil {
		return err
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	// Apply pending migrations
	appliedCount := 0
	for _, migration := range m.migrations {
		if applied[migration.Version] {
			logger.Logger.Debug("Migration already applied", "version", migration.Version, "name", migration.Name)
			continue
		}

		logger.Logger.Info("Applying migration", "version", migration.Version, "name", migration.Name)

		// Begin transaction
		tx, err := m.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", migration.Version, err)
		}

		// Execute migration
		if migration.UpSQL != "" {
			_, err = tx.Exec(migration.UpSQL)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to execute migration %s: %w", migration.Version, err)
			}
		}

		// Record migration
		recordQuery := fmt.Sprintf(`
			INSERT INTO %s (version, name, applied_at) 
			VALUES ($1, $2, $3)
		`, m.tableName)
		
		_, err = tx.Exec(recordQuery, migration.Version, migration.Name, time.Now())
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
		}

		// Commit transaction
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration.Version, err)
		}

		appliedCount++
		logger.Logger.Info("Migration applied successfully", "version", migration.Version, "name", migration.Name)
	}

	if appliedCount == 0 {
		logger.Logger.Info("No pending migrations")
	} else {
		logger.Logger.Info("Database migrations completed", "applied", appliedCount)
	}

	return nil
}

// GetMigrationStatus returns the status of all migrations
func (m *Migrator) GetMigrationStatus() ([]MigrationStatus, error) {
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return nil, err
	}

	var status []MigrationStatus
	for _, migration := range m.migrations {
		status = append(status, MigrationStatus{
			Version: migration.Version,
			Name:    migration.Name,
			Applied: applied[migration.Version],
		})
	}

	return status, nil
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	Version string
	Name    string
	Applied bool
}
