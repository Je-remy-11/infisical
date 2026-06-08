package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/infisical/api/internal/repository"
	"github.com/infisical/api/internal/repository/mocks"
)

func TestUserRepository_GetUserByEmail(t *testing.T) {
	ctx := context.Background()
	email := "test@example.com"
	userID := uuid.New()
	query := `SELECT id, email, name FROM users WHERE email = $1`

	tests := []struct {
		name          string
		email         string
		mockSetup     func(mockQuerier *mocks.MockQuerier, mockRow *mocks.MockRow)
		expectedUser  *repository.User
		expectedError error
	}{
		{
			name:  "Success - User found",
			email: email,
			mockSetup: func(mockQuerier *mocks.MockQuerier, mockRow *mocks.MockRow) {
				// Expect QueryRow to be called with correct arguments and return our mocked row
				mockQuerier.On("QueryRow", ctx, query, email).Return(mockRow)

				// Expect Scan to be called, and simulate scanning values into the pointers
				mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					idPtr := args.Get(0).(*uuid.UUID)
					emailPtr := args.Get(1).(*string)
					namePtr := args.Get(2).(*string)

					*idPtr = userID
					*emailPtr = email
					*namePtr = "Test User"
				}).Return(nil)
			},
			expectedUser: &repository.User{
				ID:    userID,
				Email: email,
				Name:  "Test User",
			},
			expectedError: nil,
		},
		{
			name:  "Not Found - User does not exist",
			email: email,
			mockSetup: func(mockQuerier *mocks.MockQuerier, mockRow *mocks.MockRow) {
				mockQuerier.On("QueryRow", ctx, query, email).Return(mockRow)
				// Simulate returning pgx.ErrNoRows when no user is found
				mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything).Return(pgx.ErrNoRows)
			},
			expectedUser:  nil,
			expectedError: repository.ErrUserNotFound,
		},
		{
			name:  "Database Error - Internal error during scan",
			email: email,
			mockSetup: func(mockQuerier *mocks.MockQuerier, mockRow *mocks.MockRow) {
				mockQuerier.On("QueryRow", ctx, query, email).Return(mockRow)
				// Simulate a generic database error
				mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db connection lost"))
			},
			expectedUser:  nil,
			expectedError: errors.New("db connection lost"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockQuerier := new(mocks.MockQuerier)
			mockRow := new(mocks.MockRow)

			tt.mockSetup(mockQuerier, mockRow)

			repo := repository.NewUserRepository(mockQuerier)

			// Act
			user, err := repo.GetUserByEmail(ctx, tt.email)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedUser, user)

			// Verify that all expectations were met
			mockQuerier.AssertExpectations(t)
			mockRow.AssertExpectations(t)
		})
	}
}
