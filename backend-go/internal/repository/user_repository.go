package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/infisical/api/internal/database/pg"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID    uuid.UUID
	Email string
	Name  string
}

type UserRepository struct {
	db pg.Querier
}

func NewUserRepository(db pg.Querier) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, name FROM users WHERE email = $1`
	row := r.db.QueryRow(ctx, query, email)

	var user User
	err := row.Scan(&user.ID, &user.Email, &user.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}
