package migrations

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/moulaybdl/incubAT/project_service/internal/logger"
)

// SimpleMigrator handles database migrations with file system access
type SimpleMigrator struct {
	db        *sql.DB
	tableName string
}

// NewSimpleMigrator creates a new simple migration instance
func NewSimpleMigrator(db *sql.DB) *SimpleMigrator {
	return &SimpleMigrator{
		db:        db,
		tableName: "schema_migrations",
	}
}

// RunMigrations runs all pending migrations from the migrations directory
func (m *SimpleMigrator) RunMigrations(migrationsDir string) error {
	logger.Logger.Info("Starting database migrations", "dir", migrationsDir)

	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(); err != nil {
		return err
	}

	// Get migration files
	migrationFiles, err := m.getMigrationFiles(migrationsDir)
	if err != nil {
		return err
	}

	if len(migrationFiles) == 0 {
		logger.Logger.Info("No migration files found")
		return nil
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	// Apply pending migrations
	appliedCount := 0
	for _, file := range migrationFiles {
		version := m.extractVersion(file)
		if applied[version] {
			logger.Logger.Debug("Migration already applied", "version", version, "file", file)
			continue
		}

		logger.Logger.Info("Applying migration", "version", version, "file", file)

		if err := m.applyMigration(file, version, migrationsDir); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", file, err)
		}

		appliedCount++
		logger.Logger.Info("Migration applied successfully", "version", version, "file", file)
	}

	if appliedCount == 0 {
		logger.Logger.Info("No pending migrations")
	} else {
		logger.Logger.Info("Database migrations completed", "applied", appliedCount)
	}

	return nil
}

// createMigrationsTable creates the schema_migrations table if it doesn't exist
func (m *SimpleMigrator) createMigrationsTable() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version VARCHAR(255) PRIMARY KEY,
			filename VARCHAR(255) NOT NULL,
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

// getMigrationFiles returns sorted list of .up.sql migration files
func (m *SimpleMigrator) getMigrationFiles(migrationsDir string) ([]string, error) {
	pattern := filepath.Join(migrationsDir, "*.up.sql")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration directory: %w", err)
	}

	// Sort files to ensure proper order
	sort.Strings(files)

	// Convert to just filenames
	var filenames []string
	for _, file := range files {
		filenames = append(filenames, filepath.Base(file))
	}

	logger.Logger.Debug("Found migration files", "count", len(filenames), "files", filenames)
	return filenames, nil
}

// extractVersion extracts version from migration filename
func (m *SimpleMigrator) extractVersion(filename string) string {
	// Extract version from filename like "000001_create_tables.up.sql"
	parts := strings.Split(filename, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return filename
}

// getAppliedMigrations returns a map of applied migration versions
func (m *SimpleMigrator) getAppliedMigrations() (map[string]bool, error) {
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

	logger.Logger.Debug("Found applied migrations", "count", len(applied))
	return applied, nil
}

// applyMigration applies a single migration file
func (m *SimpleMigrator) applyMigration(filename, version, migrationsDir string) error {
	// Read migration file
	filepath := filepath.Join(migrationsDir, filename)
	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %w", filepath, err)
	}

	sql := string(content)
	if strings.TrimSpace(sql) == "" {
		logger.Logger.Warn("Migration file is empty", "file", filename)
		// Still record it as applied to avoid re-running
		return m.recordMigration(version, filename)
	}

	// Begin transaction
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Execute migration SQL
	_, err = tx.Exec(sql)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration as applied
	recordQuery := fmt.Sprintf(`
		INSERT INTO %s (version, filename, applied_at) 
		VALUES ($1, $2, $3)
	`, m.tableName)
	
	_, err = tx.Exec(recordQuery, version, filename, time.Now())
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

// recordMigration records a migration as applied (for empty files)
func (m *SimpleMigrator) recordMigration(version, filename string) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (version, filename, applied_at) 
		VALUES ($1, $2, $3)
	`, m.tableName)

	_, err := m.db.Exec(query, version, filename, time.Now())
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}

// GetMigrationStatus returns the status of migrations
func (m *SimpleMigrator) GetMigrationStatus(migrationsDir string) error {
	files, err := m.getMigrationFiles(migrationsDir)
	if err != nil {
		return err
	}

	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	logger.Logger.Info("Migration Status:")
	for _, file := range files {
		version := m.extractVersion(file)
		status := "PENDING"
		if applied[version] {
			status = "APPLIED"
		}
		logger.Logger.Info("Migration", "version", version, "file", file, "status", status)
	}

	return nil
}
