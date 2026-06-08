package mocks

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"
)

type DB struct {
	mock.Mock
}

func NewDB(t interface {
	mock.TestingT
	Cleanup(func())
}) *DB {
	m := &DB{}
	m.Mock.Test(t)

	t.Cleanup(func() {
		m.AssertExpectations(t)
	})

	return m
}

func (m *DB) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	callArgs := make([]any, 0, 2+len(args))
	callArgs = append(callArgs, ctx, query)
	callArgs = append(callArgs, args...)

	ret := m.Called(callArgs...)
	if len(ret) == 0 {
		panic("no return value specified for QueryRow")
	}

	if fn, ok := ret.Get(0).(func(context.Context, string, ...any) pgx.Row); ok {
		return fn(ctx, query, args...)
	}

	if row, ok := ret.Get(0).(pgx.Row); ok {
		return row
	}

	return nil
}
