package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	internalHttp "github.com/r-erema/wapi/internal/http"
	"github.com/r-erema/wapi/internal/model"
	"github.com/r-erema/wapi/internal/service"
	httpTest "github.com/r-erema/wapi/internal/testutil/http"
	"github.com/r-erema/wapi/internal/testutil/mock"

	"github.com/Rhymen/go-whatsapp"
	"github.com/gavv/httpexpect"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	cs := mock.NewMockConnections(mockCtrl)
	assert.NotNil(t, internalHttp.NewInfo(cs))
}

func TestActiveConnectionInfoHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name         string
		mocksFactory func(t *testing.T) *mock.MockConnections
		expectStatus int
	}{
		{
			name: "OK",
			mocksFactory: func(t *testing.T) *mock.MockConnections {
				mockCtrl := gomock.NewController(t)
				cs := mock.NewMockConnections(mockCtrl)
				cs.EXPECT().
					AuthenticatedConnectionForSession(gomock.Any()).
					DoAndReturn(func(sessionID string) (*service.SessionConnectionDTO, error) {
						conn := mock.NewMockConn(mockCtrl)
						conn.EXPECT().Info().Return(&whatsapp.Info{Wid: "wid"})
						return service.NewDTO(conn, &model.WapiSession{}, make(chan string)), nil
					})
				return cs
			},
			expectStatus: http.StatusOK,
		},
		{
			name: "Session not found",
			mocksFactory: func(t *testing.T) *mock.MockConnections {
				mockCtrl := gomock.NewController(t)
				cs := mock.NewMockConnections(mockCtrl)
				cs.EXPECT().
					AuthenticatedConnectionForSession(gomock.Any()).
					DoAndReturn(func(sessionID string) (*service.SessionConnectionDTO, error) {
						return nil, &service.NotFoundError{SessionID: sessionID}
					})
				return cs
			},
			expectStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			handler := internalHttp.NewInfo(tt.mocksFactory(t))
			server := httpTest.New(map[string]internalHttp.AppHTTPHandler{"/get-active-connection-info/{sessionID}/": handler})
			defer server.Close()

			expect := httpexpect.New(t, server.URL)
			expect.GET("/get-active-connection-info/{sessionID}/", "_sess_id_").
				Expect().
				Status(tt.expectStatus)
		})
	}
}

func TestFailEncodeConnectionInfo(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	cs := mock.NewMockConnections(mockCtrl)
	cs.EXPECT().
		AuthenticatedConnectionForSession(gomock.Any()).
		DoAndReturn(func(sessionID string) (*service.SessionConnectionDTO, error) {
			conn := mock.NewMockConn(mockCtrl)
			conn.EXPECT().Info().Return(&whatsapp.Info{Wid: "wid"})
			return service.NewDTO(conn, &model.WapiSession{}, make(chan string)), nil
		})
	handler := internalHttp.NewInfo(cs)
	w := mock.NewFailResponseRecorder(httptest.NewRecorder())
	r, err := http.NewRequest("GET", "/get-active-connection-info/_sess_id_/", nil)
	require.Nil(t, err)

	internalHttp.AppHandlerRunner{H: handler}.ServeHTTP(w, r)

	assert.Equal(t, w.Status(), http.StatusInternalServerError)
}
