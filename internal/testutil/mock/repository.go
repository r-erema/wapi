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

// MockMessage is a mock of Message interface
type MockMessage struct {
	ctrl     *gomock.Controller
	recorder *MockMessageMockRecorder
}

// MockMessageMockRecorder is the mock recorder for MockMessage
type MockMessageMockRecorder struct {
	mock *MockMessage
}

// NewMockMessage creates a new mock instance
func NewMockMessage(ctrl *gomock.Controller) *MockMessage {
	mock := &MockMessage{ctrl: ctrl}
	mock.recorder = &MockMessageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMessage) EXPECT() *MockMessageMockRecorder {
	return m.recorder
}

// SaveMessageTime mocks base method
func (m *MockMessage) SaveMessageTime(msgID string, time time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveMessageTime", msgID, time)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveMessageTime indicates an expected call of SaveMessageTime
func (mr *MockMessageMockRecorder) SaveMessageTime(msgID, time interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveMessageTime", reflect.TypeOf((*MockMessage)(nil).SaveMessageTime), msgID, time)
}

// MessageTime mocks base method
func (m *MockMessage) MessageTime(msgID string) (*time.Time, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MessageTime", msgID)
	ret0, _ := ret[0].(*time.Time)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MessageTime indicates an expected call of MessageTime
func (mr *MockMessageMockRecorder) MessageTime(msgID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MessageTime", reflect.TypeOf((*MockMessage)(nil).MessageTime), msgID)
}

// MockSession is a mock of Session interface
type MockSession struct {
	ctrl     *gomock.Controller
	recorder *MockSessionMockRecorder
}

// MockSessionMockRecorder is the mock recorder for MockSession
type MockSessionMockRecorder struct {
	mock *MockSession
}

// NewMockSession creates a new mock instance
func NewMockSession(ctrl *gomock.Controller) *MockSession {
	mock := &MockSession{ctrl: ctrl}
	mock.recorder = &MockSessionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSession) EXPECT() *MockSessionMockRecorder {
	return m.recorder
}

// ReadSession mocks base method
func (m *MockSession) ReadSession(sessionID string) (*model.WapiSession, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadSession", sessionID)
	ret0, _ := ret[0].(*model.WapiSession)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadSession indicates an expected call of ReadSession
func (mr *MockSessionMockRecorder) ReadSession(sessionID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadSession", reflect.TypeOf((*MockSession)(nil).ReadSession), sessionID)
}

// WriteSession mocks base method
func (m *MockSession) WriteSession(session *model.WapiSession) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteSession", session)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteSession indicates an expected call of WriteSession
func (mr *MockSessionMockRecorder) WriteSession(session interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteSession", reflect.TypeOf((*MockSession)(nil).WriteSession), session)
}

// AllSavedSessionIds mocks base method
func (m *MockSession) AllSavedSessionIds() ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AllSavedSessionIds")
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AllSavedSessionIds indicates an expected call of AllSavedSessionIds
func (mr *MockSessionMockRecorder) AllSavedSessionIds() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllSavedSessionIds", reflect.TypeOf((*MockSession)(nil).AllSavedSessionIds))
}

// RemoveSession mocks base method
func (m *MockSession) RemoveSession(sessionID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveSession", sessionID)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveSession indicates an expected call of RemoveSession
func (mr *MockSessionMockRecorder) RemoveSession(sessionID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveSession", reflect.TypeOf((*MockSession)(nil).RemoveSession), sessionID)
}
