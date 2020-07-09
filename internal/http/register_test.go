package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/r-erema/wapi/internal/model"
	"github.com/r-erema/wapi/internal/testutil/mock"

	"github.com/gavv/httpexpect/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandlerHTTPRequests(t *testing.T) {
	tests := []struct {
		name         string
		data         interface{}
		mocksFactory func(t *testing.T) (*mock.MockAuthorizer, *mock.MockListener, *mock.MockSession)
		expectStatus int
	}{
		{
			"OK",
			map[string]string{"session_id": "session_id_token_81E25FCF8393C916D131A81C60AFFEB11"},
			prepareMocks,
			http.StatusOK,
		},
		{
			"Invalid JSON",
			"invalid__json",
			prepareMocks,
			http.StatusBadRequest,
		},
		{
			"Empty Session ID",
			map[string]string{"session_id": ""},
			prepareMocks,
			http.StatusBadRequest,
		},
		{
			"Listener error",
			map[string]string{"session_id": "session_id_token_81E25FCF8393C916D131A81C60AFFEB11"},
			func(t *testing.T) (*mock.MockAuthorizer, *mock.MockListener, *mock.MockSession) {
				mockCtrl := gomock.NewController(t)
				listener := mock.NewMockListener(mockCtrl)
				listener.EXPECT().
					ListenForSession(gomock.Any(), gomock.Any()).
					DoAndReturn(func(sessionID string, wg *sync.WaitGroup) (bool, error) {
						return false, fmt.Errorf("something went wrong... ")
					}).
					Do(func(sessionID string, wg *sync.WaitGroup) {
						wg.Done()
					})
				auth, _, sessionWorks := prepareMocks(t)
				return auth, listener, sessionWorks
			},
			http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			handler := NewRegisterSessionHandler(tt.mocksFactory(t))
			server := httptest.NewServer(handler)
			defer server.Close()
			expect := httpexpect.New(t, server.URL)

			expect.POST("/register-session/").
				WithJSON(tt.data).
				Expect().
				Status(tt.expectStatus)
		})
	}
}

func TestFailRestoreSessions(t *testing.T) {
	auth, listener, _ := prepareMocks(t)
	mockCtrl := gomock.NewController(t)
	sessionRepo := mock.NewMockSession(mockCtrl)
	sessionRepo.EXPECT().AllSavedSessionIds().DoAndReturn(func() ([]string, error) {
		return nil, fmt.Errorf("something went wrong... ")
	})
	handler := NewRegisterSessionHandler(auth, listener, sessionRepo)
	err := handler.TryToAutoConnectAllSessions()
	assert.NotNil(t, err)
}

func TestSuccessRestoreSessions(t *testing.T) {
	auth, _, _ := prepareMocks(t)
	mockCtrl := gomock.NewController(t)
	sessionRepo := mock.NewMockSession(mockCtrl)
	sessionRepo.EXPECT().AllSavedSessionIds().DoAndReturn(func() ([]string, error) {
		return []string{
			"sess_id_1",
			"sess_id_2",
		}, nil
	})

	listener := mock.NewMockListener(mockCtrl)
	listener.EXPECT().
		ListenForSession(gomock.Any(), gomock.Any()).
		MinTimes(2).
		DoAndReturn(func(sessionID string, wg *sync.WaitGroup) (bool, error) {
			return true, nil
		}).
		Do(func(sessionID string, wg *sync.WaitGroup) {
			wg.Done()
		})

	handler := NewRegisterSessionHandler(auth, listener, sessionRepo)
	err := handler.TryToAutoConnectAllSessions()
	assert.Nil(t, err)
}

func TestSkipFailedListenerOnRestoringSessions(t *testing.T) {
	auth, _, _ := prepareMocks(t)
	mockCtrl := gomock.NewController(t)
	sessionRepo := mock.NewMockSession(mockCtrl)
	sessionRepo.EXPECT().AllSavedSessionIds().DoAndReturn(func() ([]string, error) {
		return []string{"sess_id_1"}, nil
	})

	listener := mock.NewMockListener(mockCtrl)
	listener.EXPECT().
		ListenForSession(gomock.Any(), gomock.Any()).
		DoAndReturn(func(sessionID string, wg *sync.WaitGroup) (bool, error) {
			return false, fmt.Errorf("something went wrong... ")
		}).
		Do(func(sessionID string, wg *sync.WaitGroup) {
			wg.Done()
		})

	handler := NewRegisterSessionHandler(auth, listener, sessionRepo)
	err := handler.TryToAutoConnectAllSessions()
	assert.Nil(t, err)
}

func prepareMocks(t *testing.T) (
	auth *mock.MockAuthorizer,
	listener *mock.MockListener,
	sessionRepo *mock.MockSession,
) {
	sessionID := "session_id_token_81E25FCF8393C916D131A81C60AFFEB11"
	mockCtrl := gomock.NewController(t)
	auth = mock.NewMockAuthorizer(mockCtrl)
	conn := mock.NewMockConn(mockCtrl)
	auth.EXPECT().Login(sessionID).Return(conn, &model.WapiSession{}, nil)

	listener = mock.NewMockListener(mockCtrl)
	listener.EXPECT().
		ListenForSession(gomock.Any(), gomock.Any()).
		Do(func(sessionID string, wg *sync.WaitGroup) { wg.Done() })
	sessionRepo = mock.NewMockSession(mockCtrl)
	return
}
