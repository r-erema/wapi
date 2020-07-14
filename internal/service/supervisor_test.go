package service_test

import (
	"testing"
	"time"

	"github.com/r-erema/wapi/internal/model"
	"github.com/r-erema/wapi/internal/service"
	"github.com/r-erema/wapi/internal/testutil/mock"

	"github.com/Rhymen/go-whatsapp"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSV(t *testing.T) {
	sv := service.NewSV(0)
	assert.NotNil(t, sv)
}

func TestAddAuthenticatedConnectionForSession(t *testing.T) {
	tests := []struct {
		name         string
		mocksFactory func(t *testing.T) (*mock.MockConn, *model.WapiSession, chan string)
		expectError,
		sendQuitSignal bool
	}{
		{
			name:           "OK",
			mocksFactory:   SVMocks,
			expectError:    false,
			sendQuitSignal: true,
		},
		{
			name: "Not active connection error",
			mocksFactory: func(t *testing.T) (*mock.MockConn, *model.WapiSession, chan string) {
				_, sess, q := SVMocks(t)
				c := gomock.NewController(t)
				conn := mock.NewMockConn(c)
				conn.EXPECT().AdminTest().AnyTimes().Return(false, nil)
				return conn, sess, q
			},
			expectError:    true,
			sendQuitSignal: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			sv := service.NewSV(1)
			conn, sess, q := tt.mocksFactory(t)
			err := sv.AddAuthenticatedConnectionForSession("_sid_", service.NewDTO(conn, sess, q))

			time.Sleep(time.Millisecond * 10)

			if tt.sendQuitSignal {
				q <- ""
			}

			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestRemoveConnectionForSession(t *testing.T) {
	sv := service.NewSV(1)
	conn, sess, q := SVMocks(t)
	err := sv.AddAuthenticatedConnectionForSession("_sid_", service.NewDTO(conn, sess, q))
	require.Nil(t, err)
	sv.RemoveConnectionForSession("_sid_")
}

func TestAuthenticatedConnectionForSession(t *testing.T) {
	tests := []struct {
		name         string
		mocksFactory func(t *testing.T) (*mock.MockConn, *model.WapiSession, chan string)
		expectError,
		addSession bool
	}{
		{
			name:         "OK",
			mocksFactory: SVMocks,
			addSession:   true,
			expectError:  false,
		},
		{
			name: "Device doesn't response",
			mocksFactory: func(t *testing.T) (*mock.MockConn, *model.WapiSession, chan string) {
				_, sess, q := SVMocks(t)
				c := gomock.NewController(t)
				conn := mock.NewMockConn(c)
				conn.EXPECT().AdminTest().MaxTimes(1).Return(true, nil)
				conn.EXPECT().AdminTest().MaxTimes(1).Return(false, nil)
				return conn, sess, q
			},
			addSession:  true,
			expectError: true,
		},
		{
			name:         "Session not found",
			mocksFactory: SVMocks,
			addSession:   false,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			sv := service.NewSV(1)

			if tt.addSession {
				conn, sess, q := tt.mocksFactory(t)
				err := sv.AddAuthenticatedConnectionForSession("_wid_", service.NewDTO(conn, sess, q))
				q <- ""
				require.Nil(t, err)
			}

			connSess, err := sv.AuthenticatedConnectionForSession("_wid_")

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Nil(t, connSess)
			} else {
				assert.NotNil(t, connSess)
				assert.Nil(t, err)
			}
		})
	}
}

func TestNotFoundError(t *testing.T) {
	err := service.NotFoundError{SessionID: "_wid_"}
	assert.Equal(t, "connection for session `_wid_` not found", err.Error())
}

func SVMocks(t *testing.T) (*mock.MockConn, *model.WapiSession, chan string) {
	c := gomock.NewController(t)
	conn := mock.NewMockConn(c)
	conn.EXPECT().AdminTest().MaxTimes(2).Return(true, nil)
	conn.EXPECT().AdminTest().MaxTimes(3).Return(false, nil)
	conn.EXPECT().AdminTest().AnyTimes().Return(true, nil)
	conn.EXPECT().Disconnect().Return(whatsapp.Session{}, nil)
	conn.EXPECT().Info().AnyTimes().Return(&whatsapp.Info{Wid: "_wid_"})
	return conn, &model.WapiSession{}, make(chan string)
}
