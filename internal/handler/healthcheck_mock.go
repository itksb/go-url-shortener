// Code generated by MockGen. DO NOT EDIT.
// Source: healthcheck.go

// Package mock_handler is a generated GoMock package.
package handler

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockIPingableDB is a mock of IPingableDB interface.
type MockIPingableDB struct {
	ctrl     *gomock.Controller
	recorder *MockIPingableDBMockRecorder
}

// MockIPingableDBMockRecorder is the mock recorder for MockIPingableDB.
type MockIPingableDBMockRecorder struct {
	mock *MockIPingableDB
}

// NewMockIPingableDB creates a new mock instance.
func NewMockIPingableDB(ctrl *gomock.Controller) *MockIPingableDB {
	mock := &MockIPingableDB{ctrl: ctrl}
	mock.recorder = &MockIPingableDBMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIPingableDB) EXPECT() *MockIPingableDBMockRecorder {
	return m.recorder
}

// Ping mocks base method.
func (m *MockIPingableDB) Ping(ctx context.Context) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping", ctx)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockIPingableDBMockRecorder) Ping(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockIPingableDB)(nil).Ping), ctx)
}