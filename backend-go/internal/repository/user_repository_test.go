package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/infisical/api/internal/database/pg"
)

type mockRow struct {
	scanFn func(dest ...any) error
}

func (m *mockRow) Scan(dest ...any) error {
	if m.scanFn != nil {
		return m.scanFn(dest...)
	}
	return nil
}

type mockQuerier struct {
	queryRowFn func(ctx context.Context, sql string, args ...any) pgx.Row
	queryFn    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	execFn     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func (m *mockQuerier) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.queryRowFn != nil {
		return m.queryRowFn(ctx, sql, args...)
	}
	return &mockRow{}
}

func (m *mockQuerier) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.queryFn != nil {
		return m.queryFn(ctx, sql, args...)
	}
	return nil, nil
}

func (m *mockQuerier) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if m.execFn != nil {
		return m.execFn(ctx, sql, args...)
	}
	return pgconn.CommandTag{}, nil
}

var _ pg.Querier = (*mockQuerier)(nil)

func fixedUser() User {
	return User{
		ID:             uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Email:          "test@example.com",
		Username:       "testuser",
		FirstName:      "Test",
		LastName:       "User",
		IsEmailVerified: true,
		IsAccepted:     true,
		CreatedAt:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:      time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestGetUserByEmail(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		expectedUser := fixedUser()

		tests := []struct {
			name  string
			email string
			mock  *mockQuerier
		}{
			{
				name:  "find user by lowercase email",
				email: "test@example.com",
				mock: &mockQuerier{
					queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
						return &mockRow{
							scanFn: func(dest ...any) error {
								require.Len(t, dest, 9)
								*(dest[0].(*uuid.UUID)) = expectedUser.ID
								*(dest[1].(*string)) = expectedUser.Email
								*(dest[2].(*string)) = expectedUser.Username
								*(dest[3].(*string)) = expectedUser.FirstName
								*(dest[4].(*string)) = expectedUser.LastName
								*(dest[5].(*bool)) = expectedUser.IsEmailVerified
								*(dest[6].(*bool)) = expectedUser.IsAccepted
								*(dest[7].(*time.Time)) = expectedUser.CreatedAt
								*(dest[8].(*time.Time)) = expectedUser.UpdatedAt
								return nil
							},
						}
					},
				},
			},
			{
				name:  "find user by mixed case email",
				email: "Test@Example.Com",
				mock: &mockQuerier{
					queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
						return &mockRow{
							scanFn: func(dest ...any) error {
								*(dest[0].(*uuid.UUID)) = expectedUser.ID
								*(dest[1].(*string)) = expectedUser.Email
								*(dest[2].(*string)) = expectedUser.Username
								*(dest[3].(*string)) = expectedUser.FirstName
								*(dest[4].(*string)) = expectedUser.LastName
								*(dest[5].(*bool)) = expectedUser.IsEmailVerified
								*(dest[6].(*bool)) = expectedUser.IsAccepted
								*(dest[7].(*time.Time)) = expectedUser.CreatedAt
								*(dest[8].(*time.Time)) = expectedUser.UpdatedAt
								return nil
							},
						}
					},
				},
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				repo := NewUserRepository(tt.mock)
				user, err := repo.GetUserByEmail(context.Background(), tt.email)

				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, expectedUser.ID, user.ID)
				assert.Equal(t, expectedUser.Email, user.Email)
				assert.Equal(t, expectedUser.Username, user.Username)
				assert.Equal(t, expectedUser.FirstName, user.FirstName)
				assert.Equal(t, expectedUser.LastName, user.LastName)
				assert.Equal(t, expectedUser.IsEmailVerified, user.IsEmailVerified)
				assert.Equal(t, expectedUser.IsAccepted, user.IsAccepted)
				assert.True(t, expectedUser.CreatedAt.Equal(user.CreatedAt))
				assert.True(t, expectedUser.UpdatedAt.Equal(user.UpdatedAt))
			})
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name  string
			email string
			mock  *mockQuerier
		}{
			{
				name:  "user not found by email",
				email: "nonexistent@example.com",
				mock: &mockQuerier{
					queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
						return &mockRow{
							scanFn: func(dest ...any) error {
								return pgx.ErrNoRows
							},
						}
					},
				},
			},
			{
				name:  "user not found by empty email",
				email: "",
				mock: &mockQuerier{
					queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
						return &mockRow{
							scanFn: func(dest ...any) error {
								return pgx.ErrNoRows
							},
						}
					},
				},
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				repo := NewUserRepository(tt.mock)
				user, err := repo.GetUserByEmail(context.Background(), tt.email)

				require.Error(t, err)
				assert.True(t, errors.Is(err, sql.ErrNoRows))
				assert.Nil(t, user)
			})
		}
	})

	t.Run("database error", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name       string
			email      string
			mock       *mockQuerier
			errContains string
		}{
			{
				name:  "connection refused",
				email: "test@example.com",
				mock: &mockQuerier{
					queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
						return &mockRow{
							scanFn: func(dest ...any) error {
								return errors.New("connection refused")
							},
						}
					},
				},
				errContains: "connection refused",
			},
			{
				name:  "query timeout",
				email: "test@example.com",
				mock: &mockQuerier{
					queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
						return &mockRow{
							scanFn: func(dest ...any) error {
								return context.DeadlineExceeded
							},
						}
					},
				},
				errContains: "context deadline exceeded",
			},
			{
				name:  "column type mismatch",
				email: "test@example.com",
				mock: &mockQuerier{
					queryRowFn: func(ctx context.Context, sql string, args ...any) pgx.Row {
						return &mockRow{
							scanFn: func(dest ...any) error {
								return errors.New("cannot scan into *string: expected int4")
							},
						}
					},
				},
				errContains: "cannot scan",
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				repo := NewUserRepository(tt.mock)
				user, err := repo.GetUserByEmail(context.Background(), tt.email)

				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, user)
			})
		}
	})
}

func TestNewUserRepository(t *testing.T) {
	t.Parallel()

	querier := &mockQuerier{}
	repo := NewUserRepository(querier)

	assert.NotNil(t, repo)
	assert.Implements(t, (*UserRepository)(nil), repo)
}