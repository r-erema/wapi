// Code generated by MockGen. DO NOT EDIT.
// Source: internal/repository/repository.go

// Package mock is a generated GoMock package.
package mock

import (
	gomock "github.com/golang/mock/gomock"
	model "github.com/r-erema/wapi/internal/model"
	reflect "reflect"
	time "time"
)

// MockMessageRepository is a mock of MessageRepository interface
type MockMessageRepository struct {
	ctrl     *gomock.Controller
	recorder *MockMessageRepositoryMockRecorder
}

// MockMessageRepositoryMockRecorder is the mock recorder for MockMessageRepository
type MockMessageRepositoryMockRecorder struct {
	mock *MockMessageRepository
}

// NewMockMessageRepository creates a new mock instance
func NewMockMessageRepository(ctrl *gomock.Controller) *MockMessageRepository {
	mock := &MockMessageRepository{ctrl: ctrl}
	mock.recorder = &MockMessageRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMessageRepository) EXPECT() *MockMessageRepositoryMockRecorder {
	return m.recorder
}

// SaveMessageTime mocks base method
func (m *MockMessageRepository) SaveMessageTime(msgID string, time time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveMessageTime", msgID, time)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveMessageTime indicates an expected call of SaveMessageTime
func (mr *MockMessageRepositoryMockRecorder) SaveMessageTime(msgID, time interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveMessageTime", reflect.TypeOf((*MockMessageRepository)(nil).SaveMessageTime), msgID, time)
}

// MessageTime mocks base method
func (m *MockMessageRepository) MessageTime(msgID string) (*time.Time, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MessageTime", msgID)
	ret0, _ := ret[0].(*time.Time)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MessageTime indicates an expected call of MessageTime
func (mr *MockMessageRepositoryMockRecorder) MessageTime(msgID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MessageTime", reflect.TypeOf((*MockMessageRepository)(nil).MessageTime), msgID)
}

// MockSessionRepository is a mock of SessionRepository interface
type MockSessionRepository struct {
	ctrl     *gomock.Controller
	recorder *MockSessionRepositoryMockRecorder
}

// MockSessionRepositoryMockRecorder is the mock recorder for MockSessionRepository
type MockSessionRepositoryMockRecorder struct {
	mock *MockSessionRepository
}

// NewMockSessionRepository creates a new mock instance
func NewMockSessionRepository(ctrl *gomock.Controller) *MockSessionRepository {
	mock := &MockSessionRepository{ctrl: ctrl}
	mock.recorder = &MockSessionRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSessionRepository) EXPECT() *MockSessionRepositoryMockRecorder {
	return m.recorder
}

// ReadSession mocks base method
func (m *MockSessionRepository) ReadSession(sessionID string) (*model.WapiSession, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadSession", sessionID)
	ret0, _ := ret[0].(*model.WapiSession)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadSession indicates an expected call of ReadSession
func (mr *MockSessionRepositoryMockRecorder) ReadSession(sessionID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadSession", reflect.TypeOf((*MockSessionRepository)(nil).ReadSession), sessionID)
}

// WriteSession mocks base method
func (m *MockSessionRepository) WriteSession(session *model.WapiSession) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteSession", session)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteSession indicates an expected call of WriteSession
func (mr *MockSessionRepositoryMockRecorder) WriteSession(session interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteSession", reflect.TypeOf((*MockSessionRepository)(nil).WriteSession), session)
}

// AllSavedSessionIds mocks base method
func (m *MockSessionRepository) AllSavedSessionIds() ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AllSavedSessionIds")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AllSavedSessionIds indicates an expected call of AllSavedSessionIds
func (mr *MockSessionRepositoryMockRecorder) AllSavedSessionIds() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllSavedSessionIds", reflect.TypeOf((*MockSessionRepository)(nil).AllSavedSessionIds))
}

// RemoveSession mocks base method
func (m *MockSessionRepository) RemoveSession(sessionID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveSession", sessionID)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveSession indicates an expected call of RemoveSession
func (mr *MockSessionRepositoryMockRecorder) RemoveSession(sessionID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveSession", reflect.TypeOf((*MockSessionRepository)(nil).RemoveSession), sessionID)
}
