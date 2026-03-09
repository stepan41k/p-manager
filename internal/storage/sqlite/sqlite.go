package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

func New(driverName string, connectionString string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open(driverName, connectionString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	storage := &Storage{db: db}
	if err := storage.migrate(); err != nil {
		return nil, fmt.Errorf("migrate failed: %w", err)
	}

	return storage, nil
}


var migrations = []string{
	`CREATE TABLE vault_metadata (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		salt BLOB NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE entries (
		id SERIAL PRIMARY KEY,
		service TEXT NOT NULL,
		username TEXT NOT NULL,
		encrypted_password BLOB NOT NULL,
		nonce BLOB NOT NULL
	);`,
}

func (s *Storage) migrate() error {
	var currentVersion int

	err := s.db.QueryRow("PRAGMA user_version").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("failed to get schema version: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	for i := currentVersion; i < len(migrations); i++ {
		_, err := tx.Exec(migrations[i])
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration v%d: %w", i+1, err)
		}
		
		_, err = tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", i+1))
		if err != nil {
			tx.Rollback()
			return err
		}
		fmt.Printf("Migrated database to version %d\n", i+1)
	}

	return tx.Commit()
}