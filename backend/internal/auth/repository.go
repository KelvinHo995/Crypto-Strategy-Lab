package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) *PostgresRepository { return &PostgresRepository{db: db} }
func (r *PostgresRepository) Create(ctx context.Context, u User) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO users(id,username,password_hash,created_at) VALUES($1,$2,$3,$4)`, u.ID, u.Username, u.PasswordHash, u.CreatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrUsernameTaken
	}
	return err
}
func (r *PostgresRepository) ByUsername(ctx context.Context, username string) (User, error) {
	var u User
	err := r.db.QueryRowContext(ctx, `SELECT id,username,password_hash,created_at FROM users WHERE username=$1`, username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrInvalidCredentials
	}
	return u, err
}

type MemoryRepository struct {
	mu    sync.RWMutex
	users map[string]User
}

func NewMemoryRepository() *MemoryRepository { return &MemoryRepository{users: map[string]User{}} }
func (r *MemoryRepository) Create(_ context.Context, u User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := strings.ToLower(u.Username)
	if _, ok := r.users[key]; ok {
		return ErrUsernameTaken
	}
	r.users[key] = u
	return nil
}
func (r *MemoryRepository) ByUsername(_ context.Context, name string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[strings.ToLower(name)]
	if !ok {
		return User{}, ErrInvalidCredentials
	}
	return u, nil
}
