package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/infisical/api/internal/database/pg"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID                            uuid.UUID
	Email                         *string
	AuthMethods                   []string
	SuperAdmin                    bool
	FirstName                     *string
	LastName                      *string
	IsAccepted                    bool
	IsMfaEnabled                  bool
	MfaMethods                    []string
	Devices                       interface{}
	CreatedAt                     time.Time
	UpdatedAt                     time.Time
	IsGhost                       bool
	Username                      string
	IsEmailVerified               bool
	ConsecutiveFailedMfaAttempts  int
	IsLocked                      bool
	TemporaryLockDateEnd          *time.Time
	ConsecutiveFailedPasswordAttempts int
	SelectedMfaMethod             *string
	HashedPassword                *string
	IsGoogleVerified              bool
	IsGitHubVerified              bool
	IsGitLabVerified              bool
	LastSeenAnnouncementId        *string
}

type UserRepository struct {
	querier pg.Querier
}

func NewUserRepository(querier pg.Querier) *UserRepository {
	return &UserRepository{querier: querier}
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	row := r.querier.QueryRow(ctx, `
		SELECT id, email, auth_methods, super_admin, first_name, last_name, is_accepted, 
		       is_mfa_enabled, mfa_methods, devices, created_at, updated_at, is_ghost, 
		       username, is_email_verified, consecutive_failed_mfa_attempts, is_locked, 
		       temporary_lock_date_end, consecutive_failed_password_attempts, 
		       selected_mfa_method, hashed_password, is_google_verified, 
		       is_git_hub_verified, is_git_lab_verified, last_seen_announcement_id
		FROM users
		WHERE email = $1 AND is_ghost = false
	`, email)

	var user User
	err := row.Scan(
		&user.ID, &user.Email, &user.AuthMethods, &user.SuperAdmin, &user.FirstName, 
		&user.LastName, &user.IsAccepted, &user.IsMfaEnabled, &user.MfaMethods, 
		&user.Devices, &user.CreatedAt, &user.UpdatedAt, &user.IsGhost, &user.Username, 
		&user.IsEmailVerified, &user.ConsecutiveFailedMfaAttempts, &user.IsLocked, 
		&user.TemporaryLockDateEnd, &user.ConsecutiveFailedPasswordAttempts, 
		&user.SelectedMfaMethod, &user.HashedPassword, &user.IsGoogleVerified, 
		&user.IsGitHubVerified, &user.IsGitLabVerified, &user.LastSeenAnnouncementId,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}
