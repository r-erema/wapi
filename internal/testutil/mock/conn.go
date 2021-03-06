// Code generated by MockGen. DO NOT EDIT.
// Source: internal/infrastructure/whatsapp/conn.go

// Package mock is a generated GoMock package.
package mock

import (
	whatsapp "github.com/Rhymen/go-whatsapp"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockConn is a mock of Conn interface
type MockConn struct {
	ctrl     *gomock.Controller
	recorder *MockConnMockRecorder
}

// MockConnMockRecorder is the mock recorder for MockConn
type MockConnMockRecorder struct {
	mock *MockConn
}

// NewMockConn creates a new mock instance
func NewMockConn(ctrl *gomock.Controller) *MockConn {
	mock := &MockConn{ctrl: ctrl}
	mock.recorder = &MockConnMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockConn) EXPECT() *MockConnMockRecorder {
	return m.recorder
}

// Send mocks base method
func (m *MockConn) Send(msg interface{}) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", msg)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Send indicates an expected call of Send
func (mr *MockConnMockRecorder) Send(msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockConn)(nil).Send), msg)
}

// Info mocks base method
func (m *MockConn) Info() *whatsapp.Info {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Info")
	ret0, _ := ret[0].(*whatsapp.Info)
	return ret0
}

// Info indicates an expected call of Info
func (mr *MockConnMockRecorder) Info() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockConn)(nil).Info))
}

// AdminTest mocks base method
func (m *MockConn) AdminTest() (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AdminTest")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AdminTest indicates an expected call of AdminTest
func (mr *MockConnMockRecorder) AdminTest() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AdminTest", reflect.TypeOf((*MockConn)(nil).AdminTest))
}

// Disconnect mocks base method
func (m *MockConn) Disconnect() (whatsapp.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Disconnect")
	ret0, _ := ret[0].(whatsapp.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Disconnect indicates an expected call of Disconnect
func (mr *MockConnMockRecorder) Disconnect() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Disconnect", reflect.TypeOf((*MockConn)(nil).Disconnect))
}

// RestoreWithSession mocks base method
func (m *MockConn) RestoreWithSession(session *whatsapp.Session) (whatsapp.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RestoreWithSession", session)
	ret0, _ := ret[0].(whatsapp.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RestoreWithSession indicates an expected call of RestoreWithSession
func (mr *MockConnMockRecorder) RestoreWithSession(session interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RestoreWithSession", reflect.TypeOf((*MockConn)(nil).RestoreWithSession), session)
}

// Login mocks base method
func (m *MockConn) Login(qrChan chan<- string) (whatsapp.Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Login", qrChan)
	ret0, _ := ret[0].(whatsapp.Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Login indicates an expected call of Login
func (mr *MockConnMockRecorder) Login(qrChan interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Login", reflect.TypeOf((*MockConn)(nil).Login), qrChan)
}

// AddHandler mocks base method
func (m *MockConn) AddHandler(handler whatsapp.Handler) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddHandler", handler)
}

// AddHandler indicates an expected call of AddHandler
func (mr *MockConnMockRecorder) AddHandler(handler interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddHandler", reflect.TypeOf((*MockConn)(nil).AddHandler), handler)
}
