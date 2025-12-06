package user

import (
	"database/sql"
	"time"

	"elotus_test/server/bsql"

	"github.com/lib/pq"
)

// PostgresRepository handles user database operations
type PostgresRepository struct {
	db *bsql.DB
}

// NewPostgresRepository creates a new PostgresRepository
func NewPostgresRepository(db *bsql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// CreateUser inserts a new user into the database
func (r *PostgresRepository) CreateUser(username, hashedPassword string) (*User, error) {
	now := time.Now()

	id, err := r.db.Insert("users", map[string]interface{}{
		"username":   username,
		"password":   hashedPassword,
		"created_at": now,
	})

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return nil, ErrUserExists
			}
		}
		return nil, err
	}

	return &User{
		ID:        id,
		Username:  username,
		Password:  hashedPassword,
		CreatedAt: now,
	}, nil
}

// GetUserByUsername retrieves a user by username
func (r *PostgresRepository) GetUserByUsername(username string) (*User, bool) {
	var user User
	err := r.db.QueryRow(
		`SELECT id, username, password, created_at FROM users WHERE username = $1`,
		username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false
		}
		return nil, false
	}

	return &user, true
}

// GetUserByID retrieves a user by ID
func (r *PostgresRepository) GetUserByID(id int64) (*User, bool) {
	var user User
	err := r.db.QueryRow(
		`SELECT id, username, password, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false
		}
		return nil, false
	}

	return &user, true
}
