package store

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"interviewsw/internal/domain"

	_ "github.com/lib/pq"
)

type PostgresUserStore struct {
	db *sql.DB
}

func NewPostgresUserStore(databaseURL string) (*PostgresUserStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &PostgresUserStore{db: db}, nil
}

func (s *PostgresUserStore) Close() error {
	return s.db.Close()
}

func (s *PostgresUserStore) EnsureSchema() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id BIGINT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		)
	`)
	return err
}

func (s *PostgresUserStore) SeedIfEmpty(users []domain.User) error {
	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO users (id, email, name, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, user := range users {
		if _, err = stmt.Exec(user.ID, user.Email, user.Name, user.PasswordHash, user.CreatedAt, user.UpdatedAt); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (s *PostgresUserStore) GetByID(id int64) (domain.User, error) {
	return s.scanOne(s.db.QueryRow(`
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id))
}

func (s *PostgresUserStore) GetByEmail(email string) (domain.User, error) {
	normalized := strings.TrimSpace(strings.ToLower(email))
	return s.scanOne(s.db.QueryRow(`
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM users
		WHERE lower(email) = $1
	`, normalized))
}

func (s *PostgresUserStore) UpdateName(id int64, name string) (domain.User, error) {
	now := time.Now().UTC()
	trimmedName := strings.TrimSpace(name)

	row := s.db.QueryRow(`
		UPDATE users
		SET name = $2, updated_at = $3
		WHERE id = $1
		RETURNING id, email, name, password_hash, created_at, updated_at
	`, id, trimmedName, now)

	return s.scanOne(row)
}

func (s *PostgresUserStore) ListByEmailQuery(query string) []domain.User {
	trimmed := strings.TrimSpace(strings.ToLower(query))

	rows, err := s.db.Query(`
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM users
		WHERE ($1 = '' OR lower(email) LIKE '%' || $1 || '%')
		ORDER BY id ASC
	`, trimmed)
	if err != nil {
		return []domain.User{}
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		var user domain.User
		if scanErr := rows.Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt); scanErr != nil {
			continue
		}
		users = append(users, user)
	}

	return users
}

func (s *PostgresUserStore) scanOne(row *sql.Row) (domain.User, error) {
	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, err
	}
	return user, nil
}
