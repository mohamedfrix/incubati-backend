package repository

import (
	"database/sql"

	_ "github.com/lib/pq" // Import the PostgreSQL driver
)

type DB struct {
    db *sql.DB
}

// GetDB returns the underlying sql.DB instance
func (d *DB) GetDB() *sql.DB {
    return d.db
}

func OpenDB(DSN string) (*DB, error) {
    db, err := sql.Open("postgres", DSN)
    if err != nil {
        return nil, err
    }

    if err := db.Ping(); err != nil {
        return nil, err
    }

    return &DB{
        db: db,
    }, nil
}