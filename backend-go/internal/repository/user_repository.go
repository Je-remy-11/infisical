package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/infisical/api/internal/database/pg"
)

type User struct {
	ID             uuid.UUID
	Email          string
	Username       string
	FirstName      string
	LastName       string
	IsEmailVerified bool
	IsAccepted     bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type UserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}

type userRepository struct {
	querier pg.Querier
}

func NewUserRepository(querier pg.Querier) UserRepository {
	return &userRepository{
		querier: querier,
	}
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, username, first_name, last_name, is_email_verified, is_accepted, created_at, updated_at
		FROM users
		WHERE email = $1
		LIMIT 1
	`

	var user User
	err := r.querier.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.IsEmailVerified,
		&user.IsAccepted,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &user, nil
}