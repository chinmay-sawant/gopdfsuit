package main

import (
	"database/sql"
	"errors"
	"strings"

	_ "modernc.org/sqlite" // pure-Go SQLite driver, no CGO
)

// ErrUserExists is returned when registering an email that already exists.
var ErrUserExists = errors.New("user already exists")

// ErrUserNotFound is returned when no account matches the lookup.
var ErrUserNotFound = errors.New("user not found")

// Store wraps the SQLite-backed users table.
type Store struct {
	db *sql.DB
}

// OpenStore opens (or creates) the database at dsn and runs migrations.
// Use ":memory:" for an isolated in-memory database in tests.
func OpenStore(dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// A single connection keeps an in-memory DB alive across queries.
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			email         TEXT NOT NULL UNIQUE COLLATE NOCASE,
			password_hash TEXT NOT NULL,
			created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`)
	return err
}

// Close releases the underlying database handle.
func (s *Store) Close() error { return s.db.Close() }

// CreateUser inserts a new account and returns its public record.
func (s *Store) CreateUser(email, passwordHash string) (User, error) {
	res, err := s.db.Exec(
		`INSERT INTO users (email, password_hash) VALUES (?, ?)`,
		email, passwordHash,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return User{}, ErrUserExists
		}
		return User{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return User{}, err
	}
	return s.userByID(id)
}

// Credentials returns the public user plus the stored password hash.
func (s *Store) Credentials(email string) (User, string, error) {
	var u User
	var hash string
	err := s.db.QueryRow(
		`SELECT id, email, password_hash, created_at FROM users WHERE email = ? COLLATE NOCASE`,
		email,
	).Scan(&u.ID, &u.Email, &hash, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, "", ErrUserNotFound
	}
	if err != nil {
		return User{}, "", err
	}
	return u, hash, nil
}

func (s *Store) userByID(id int64) (User, error) {
	var u User
	err := s.db.QueryRow(
		`SELECT id, email, created_at FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Email, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	return u, err
}
