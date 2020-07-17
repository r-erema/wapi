package http_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	internalHttp "github.com/r-erema/wapi/internal/http"
	"github.com/r-erema/wapi/internal/model"
	testHttp "github.com/r-erema/wapi/internal/testutil/http"
	"github.com/r-erema/wapi/internal/testutil/mock"

	"github.com/gavv/httpexpect/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSessInfoHandler(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	sessionRepo := mock.NewMockSession(mockCtrl)
	assert.NotNil(t, internalHttp.NewSessInfoHandler(sessionRepo))
}

func TestSessInfoHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name         string
		mocksFactory func(t *testing.T) *mock.MockSession
		expectStatus int
	}{
		{
			name: "OK",
			mocksFactory: func(t *testing.T) *mock.MockSession {
				mockCtrl := gomock.NewController(t)
				sessionRepo := mock.NewMockSession(mockCtrl)
				sessionRepo.EXPECT().
					ReadSession(gomock.Any()).
					DoAndReturn(func(sessionID string) (*model.WapiSession, error) {
						return &model.WapiSession{}, nil
					})
				return sessionRepo
			},
			expectStatus: http.StatusOK,
		},
		{
			name: "Session not found",
			mocksFactory: func(t *testing.T) *mock.MockSession {
				mockCtrl := gomock.NewController(t)
				sessionRepo := mock.NewMockSession(mockCtrl)
				sessionRepo.EXPECT().
					ReadSession(gomock.Any()).
					DoAndReturn(func(sessionID string) (*model.WapiSession, error) {
						return nil, &os.PathError{}
					})
				return sessionRepo
			},
			expectStatus: http.StatusNotFound,
		},
		{
			name: "Internal server error",
			mocksFactory: func(t *testing.T) *mock.MockSession {
				mockCtrl := gomock.NewController(t)
				sessionRepo := mock.NewMockSession(mockCtrl)
				sessionRepo.EXPECT().
					ReadSession(gomock.Any()).
					DoAndReturn(func(sessionID string) (*model.WapiSession, error) {
						return nil, fmt.Errorf("something went wrong... ")
					})
				return sessionRepo
			},
			expectStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			server := testHttp.New(map[string]internalHttp.AppHTTPHandler{
				"/get-session-info/_sess_id_/": internalHttp.NewSessInfoHandler(tt.mocksFactory(t)),
			})
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

	handler := internalHttp.NewSessInfoHandler(sessionRepo)
	w := mock.NewFailResponseRecorder(httptest.NewRecorder())
	r, err := http.NewRequest("GET", "/get-session-info/_sess_id_/", nil)
	require.Nil(t, err)

	internalHttp.AppHandlerRunner{H: handler}.ServeHTTP(w, r)

	assert.Equal(t, w.Status(), http.StatusInternalServerError)
}
