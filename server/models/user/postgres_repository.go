package user

import (
	"database/sql"
	"time"

	"elotus_test/server/bsql"

	"github.com/lib/pq"
)

type PostgresRepository struct {
	db *bsql.DB
}

func NewPostgresRepository(db *bsql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateUser(username, hashedPassword string) (*User, error) {
	now := time.Now()

	id, err := r.db.Insert("users", map[string]interface{}{
		"username":   username,
		"password":   hashedPassword,
		"created_at": now,
	})

	if err != nil {
		// pq error code 23505 = unique_violation
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

func (r *PostgresRepository) UpdateLastLogin(userID int64) error {
	_, err := r.db.Exec(
		`UPDATE users SET last_login_at = $1 WHERE id = $2`,
		time.Now(), userID,
	)
	return err
}
