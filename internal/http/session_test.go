package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/r-erema/wapi/internal/model"
	"github.com/r-erema/wapi/internal/testutil/mock"

	"github.com/gavv/httpexpect/v2"
	"github.com/golang/mock/gomock"
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSessInfoHandler(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	sessionRepo := mock.NewMockSession(mockCtrl)
	assert.Equal(t, NewSessInfoHandler(sessionRepo), &SessInfoHandler{sessionRepo: sessionRepo})
}

func TestSessInfoHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name         string
		mocksFactory func(t *testing.T) *mock.MockSession
		expectStatus int
	}{
		{
			"OK",
			func(t *testing.T) *mock.MockSession {
				mockCtrl := gomock.NewController(t)
				sessionRepo := mock.NewMockSession(mockCtrl)
				sessionRepo.EXPECT().
					ReadSession(gomock.Any()).
					DoAndReturn(func(sessionID string) (*model.WapiSession, error) {
						return &model.WapiSession{}, nil
					})
				return sessionRepo
			},
			http.StatusOK,
		},
		{
			"Session not found",
			func(t *testing.T) *mock.MockSession {
				mockCtrl := gomock.NewController(t)
				sessionRepo := mock.NewMockSession(mockCtrl)
				sessionRepo.EXPECT().
					ReadSession(gomock.Any()).
					DoAndReturn(func(sessionID string) (*model.WapiSession, error) {
						return nil, &os.PathError{}
					})
				return sessionRepo
			},
			http.StatusNotFound,
		},
		{
			"Internal server error",
			func(t *testing.T) *mock.MockSession {
				mockCtrl := gomock.NewController(t)
				sessionRepo := mock.NewMockSession(mockCtrl)
				sessionRepo.EXPECT().
					ReadSession(gomock.Any()).
					DoAndReturn(func(sessionID string) (*model.WapiSession, error) {
						return nil, fmt.Errorf("something went wrong... ")
					})
				return sessionRepo
			},
			http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			handler := NewSessInfoHandler(tt.mocksFactory(t))
			server := httptest.NewServer(handler)
			defer server.Close()
			expect := httpexpect.New(t, server.URL)

			expect.GET("/get-session-info/_sess_id_/").
				Expect().
				Status(tt.expectStatus)
		})
	}
}

func TestFailEncodeSession(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	sessionRepo := mock.NewMockSession(mockCtrl)
	sessionRepo.EXPECT().
		ReadSession(gomock.Any()).
		DoAndReturn(func(sessionID string) (*model.WapiSession, error) {
			return nil, nil
		})

	handler := NewSessInfoHandler(sessionRepo)
	w := mock.NewFailResponseRecorder(httptest.NewRecorder())
	r, err := http.NewRequest("GET", "/get-session-info/_sess_id_/", nil)
	require.Nil(t, err)

	handler.ServeHTTP(w, r)

	assert.Equal(t, w.Status(), http.StatusInternalServerError)
}
