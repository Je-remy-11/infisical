package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/infisical/api/internal/database/pg"
)

type mockQuerier struct {
	queryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
}

func (m *mockQuerier) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, nil
}

func (m *mockQuerier) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return m.queryRowFunc(ctx, sql, args...)
}

func (m *mockQuerier) Exec(ctx context.Context, sql string, args ...any) (interface{}, error) {
	return nil, nil
}

type mockRow struct {
	scanResult interface{}
	scanErr    error
}

func (m *mockRow) Scan(dest ...any) error {
	if m.scanErr != nil {
		return m.scanErr
	}

	if m.scanResult != nil {
		if user, ok := m.scanResult.(*User); ok {
			// Fill the destination pointers with user data
			if len(dest) >= 25 {
				if dest[0] != nil {
					*(dest[0].(*uuid.UUID)) = user.ID
				}
				if dest[1] != nil {
					*(dest[1].(**string)) = user.Email
				}
				if dest[2] != nil {
					*(dest[2].(*[]string)) = user.AuthMethods
				}
				if dest[3] != nil {
					*(dest[3].(*bool)) = user.SuperAdmin
				}
				if dest[4] != nil {
					*(dest[4].(**string)) = user.FirstName
				}
				if dest[5] != nil {
					*(dest[5].(**string)) = user.LastName
				}
				if dest[6] != nil {
					*(dest[6].(*bool)) = user.IsAccepted
				}
				if dest[7] != nil {
					*(dest[7].(*bool)) = user.IsMfaEnabled
				}
				if dest[8] != nil {
					*(dest[8].(*[]string)) = user.MfaMethods
				}
				if dest[9] != nil {
					*(dest[9].(*interface{})) = user.Devices
				}
				if dest[10] != nil {
					*(dest[10].(*time.Time)) = user.CreatedAt
				}
				if dest[11] != nil {
					*(dest[11].(*time.Time)) = user.UpdatedAt
				}
				if dest[12] != nil {
					*(dest[12].(*bool)) = user.IsGhost
				}
				if dest[13] != nil {
					*(dest[13].(*string)) = user.Username
				}
				if dest[14] != nil {
					*(dest[14].(*bool)) = user.IsEmailVerified
				}
				if dest[15] != nil {
					*(dest[15].(*int)) = user.ConsecutiveFailedMfaAttempts
				}
				if dest[16] != nil {
					*(dest[16].(*bool)) = user.IsLocked
				}
				if dest[17] != nil {
					*(dest[17].(**time.Time)) = user.TemporaryLockDateEnd
				}
				if dest[18] != nil {
					*(dest[18].(*int)) = user.ConsecutiveFailedPasswordAttempts
				}
				if dest[19] != nil {
					*(dest[19].(**string)) = user.SelectedMfaMethod
				}
				if dest[20] != nil {
					*(dest[20].(**string)) = user.HashedPassword
				}
				if dest[21] != nil {
					*(dest[21].(*bool)) = user.IsGoogleVerified
				}
				if dest[22] != nil {
					*(dest[22].(*bool)) = user.IsGitHubVerified
				}
				if dest[23] != nil {
					*(dest[23].(*bool)) = user.IsGitLabVerified
				}
				if dest[24] != nil {
					*(dest[24].(**string)) = user.LastSeenAnnouncementId
				}
			}
		}
	}
	return nil
}

func TestGetUserByEmail(t *testing.T) {
	ctx := context.Background()
	testEmail := "test@example.com"
	testUserID := uuid.New()
	now := time.Now()

	// Create a complete test user for the success case
	email := testEmail
	firstName := "Test"
	lastName := "User"
	selectedMfaMethod := "totp"
	hashedPassword := "hashed"
	expectedUser := &User{
		ID:                            testUserID,
		Email:                         &email,
		AuthMethods:                   []string{"password"},
		SuperAdmin:                    false,
		FirstName:                     &firstName,
		LastName:                      &lastName,
		IsAccepted:                    true,
		IsMfaEnabled:                  false,
		MfaMethods:                    []string{},
		Devices:                       nil,
		CreatedAt:                     now,
		UpdatedAt:                     now,
		IsGhost:                       false,
		Username:                      "testuser",
		IsEmailVerified:               true,
		ConsecutiveFailedMfaAttempts:  0,
		IsLocked:                      false,
		TemporaryLockDateEnd:          nil,
		ConsecutiveFailedPasswordAttempts: 0,
		SelectedMfaMethod:             &selectedMfaMethod,
		HashedPassword:                &hashedPassword,
		IsGoogleVerified:              false,
		IsGitHubVerified:              false,
		IsGitLabVerified:              false,
		LastSeenAnnouncementId:        nil,
	}

	testCases := []struct {
		name         string
		setupMock    func() *mockQuerier
		expectedUser *User
		expectedErr  error
	}{
		{
			name: "success - user found",
			setupMock: func() *mockQuerier {
				return &mockQuerier{
					queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
						assert.Equal(t, testEmail, args[0])
						assert.Contains(t, sql, "SELECT")
						assert.Contains(t, sql, "FROM users")
						assert.Contains(t, sql, "WHERE email = $1")
						assert.Contains(t, sql, "is_ghost = false")
						
						return &mockRow{
							scanResult: expectedUser,
							scanErr:    nil,
						}
					},
				}
			},
			expectedUser: expectedUser,
			expectedErr:  nil,
		},
		{
			name: "failure - user not found",
			setupMock: func() *mockQuerier {
				return &mockQuerier{
					queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
						assert.Equal(t, testEmail, args[0])
						return &mockRow{
							scanErr: pgx.ErrNoRows,
						}
					},
				}
			},
			expectedUser: nil,
			expectedErr:  ErrUserNotFound,
		},
		{
			name: "failure - database error",
			setupMock: func() *mockQuerier {
				return &mockQuerier{
					queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
						assert.Equal(t, testEmail, args[0])
						return &mockRow{
							scanErr: errors.New("database connection error"),
						}
					},
				}
			},
			expectedUser: nil,
			expectedErr:  errors.New("database connection error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := tc.setupMock()
			repo := NewUserRepository(mock)
			
			user, err := repo.GetUserByEmail(ctx, testEmail)
			
			if tc.expectedErr != nil {
				require.Error(t, err)
				if errors.Is(tc.expectedErr, ErrUserNotFound) {
					assert.ErrorIs(t, err, ErrUserNotFound)
				} else {
					assert.Equal(t, tc.expectedErr.Error(), err.Error())
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tc.expectedUser.ID, user.ID)
				assert.Equal(t, tc.expectedUser.Email, user.Email)
				assert.Equal(t, tc.expectedUser.Username, user.Username)
				assert.Equal(t, tc.expectedUser.IsEmailVerified, user.IsEmailVerified)
				assert.Equal(t, tc.expectedUser.FirstName, user.FirstName)
				assert.Equal(t, tc.expectedUser.LastName, user.LastName)
				assert.Equal(t, tc.expectedUser.IsAccepted, user.IsAccepted)
				assert.Equal(t, tc.expectedUser.SuperAdmin, user.SuperAdmin)
			}
		})
	}
}
