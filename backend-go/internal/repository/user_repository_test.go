package repository

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/infisical/api/internal/repository/mocks"
)

type stubRow struct {
	scanErr error
	email   string
	now     time.Time
}

func (r stubRow) Scan(dest ...any) error {
	if r.scanErr != nil {
		return r.scanErr
	}

	stringIndex := 0
	nextString := func() string {
		values := []string{
			uuid.MustParse("11111111-1111-1111-1111-111111111111").String(),
			r.email,
			"test-user",
			"test-value",
		}
		if stringIndex >= len(values) {
			value := fmt.Sprintf("value-%d", stringIndex)
			stringIndex++
			return value
		}

		value := values[stringIndex]
		stringIndex++
		return value
	}

	for _, d := range dest {
		if d == nil {
			continue
		}

		switch v := d.(type) {
		case *string:
			*v = nextString()
			continue
		case *uuid.UUID:
			*v = uuid.MustParse("11111111-1111-1111-1111-111111111111")
			continue
		case *time.Time:
			*v = r.now
			continue
		}

		rv := reflect.ValueOf(d)
		if rv.Kind() != reflect.Pointer || rv.IsNil() {
			return fmt.Errorf("invalid scan destination %T", d)
		}

		elem := rv.Elem()

		switch elem.Kind() {
		case reflect.String:
			elem.SetString(nextString())
		case reflect.Bool:
			elem.SetBool(true)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			elem.SetInt(1)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			elem.SetUint(1)
		case reflect.Slice:
			if elem.Type().Elem().Kind() == reflect.Uint8 {
				elem.SetBytes([]byte("value"))
			}
		case reflect.Struct:
			if elem.Type().PkgPath() == "time" && elem.Type().Name() == "Time" {
				elem.Set(reflect.ValueOf(r.now))
				continue
			}

			if elem.Type().PkgPath() == "github.com/google/uuid" && elem.Type().Name() == "UUID" {
				elem.Set(reflect.ValueOf(uuid.MustParse("11111111-1111-1111-1111-111111111111")))
				continue
			}

			if field := elem.FieldByName("Valid"); field.IsValid() && field.CanSet() && field.Kind() == reflect.Bool {
				field.SetBool(true)
			}
			if field := elem.FieldByName("String"); field.IsValid() && field.CanSet() && field.Kind() == reflect.String {
				field.SetString(nextString())
			}
			if field := elem.FieldByName("Bool"); field.IsValid() && field.CanSet() && field.Kind() == reflect.Bool {
				field.SetBool(true)
			}
			if field := elem.FieldByName("Int64"); field.IsValid() && field.CanSet() && field.Kind() == reflect.Int64 {
				field.SetInt(1)
			}
			if field := elem.FieldByName("Time"); field.IsValid() && field.CanSet() && field.Type().PkgPath() == "time" && field.Type().Name() == "Time" {
				field.Set(reflect.ValueOf(r.now))
			}
			if field := elem.FieldByName("V"); field.IsValid() && field.CanSet() {
				switch field.Kind() {
				case reflect.String:
					field.SetString(nextString())
				case reflect.Bool:
					field.SetBool(true)
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					field.SetInt(1)
				default:
					if field.Type().PkgPath() == "time" && field.Type().Name() == "Time" {
						field.Set(reflect.ValueOf(r.now))
					}
					if field.Type().PkgPath() == "github.com/google/uuid" && field.Type().Name() == "UUID" {
						field.Set(reflect.ValueOf(uuid.MustParse("11111111-1111-1111-1111-111111111111")))
					}
				}
			}
		}
	}

	return nil
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	email := "alice@example.com"
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	dbErr := errors.New("database connection failed")

	tests := []struct {
		name        string
		row         pgx.Row
		wantNilUser bool
		wantErr     error
	}{
		{
			name: "正常查询",
			row: stubRow{
				email: email,
				now:   now,
			},
			wantNilUser: false,
			wantErr:     nil,
		},
		{
			name: "未找到",
			row: stubRow{
				scanErr: pgx.ErrNoRows,
			},
			wantNilUser: true,
			wantErr:     nil,
		},
		{
			name: "数据库错误",
			row: stubRow{
				scanErr: dbErr,
			},
			wantNilUser: true,
			wantErr:     dbErr,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewDB(t)
			mockDB.
				On("QueryRow", mock.Anything, mock.Anything, mock.Anything).
				Return(tt.row).
				Once()

			repo := newUserRepositoryForTest(t, mockDB)

			got, err := repo.GetUserByEmail(ctx, email)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr.Error())
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)

			if tt.wantNilUser {
				assert.Nil(t, got)
				return
			}

			require.NotNil(t, got)
			assertStringFieldEquals(t, got, []string{"Email", "email"}, email)
		})
	}
}

func newUserRepositoryForTest(t *testing.T, db any) *UserRepository {
	t.Helper()

	repo := &UserRepository{}
	repoValue := reflect.ValueOf(repo).Elem()
	dbValue := reflect.ValueOf(db)

	for i := 0; i < repoValue.NumField(); i++ {
		field := repoValue.Field(i)
		if !dbValue.Type().AssignableTo(field.Type()) {
			continue
		}

		if field.CanSet() {
			field.Set(dbValue)
		} else {
			reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(dbValue)
		}
		return repo
	}

	t.Fatalf("failed to inject DB mock into UserRepository")
	return nil
}

func assertStringFieldEquals(t *testing.T, value any, fieldNames []string, want string) {
	t.Helper()

	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}

	if !rv.IsValid() || rv.Kind() != reflect.Struct {
		return
	}

	for _, fieldName := range fieldNames {
		field := rv.FieldByName(fieldName)
		if !field.IsValid() {
			continue
		}

		if field.Kind() == reflect.String {
			assert.Equal(t, want, field.String())
			return
		}

		if field.Kind() == reflect.Struct {
			stringField := field.FieldByName("String")
			validField := field.FieldByName("Valid")
			if stringField.IsValid() && stringField.Kind() == reflect.String {
				if validField.IsValid() && validField.Kind() == reflect.Bool {
					assert.True(t, validField.Bool())
				}
				assert.Equal(t, want, stringField.String())
				return
			}
			valueField := field.FieldByName("V")
			if valueField.IsValid() && valueField.Kind() == reflect.String {
				assert.Equal(t, want, valueField.String())
				return
			}
		}
	}
}
