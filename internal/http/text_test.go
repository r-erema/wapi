package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	internalHttp "github.com/r-erema/wapi/internal/http"
	jsonInfra "github.com/r-erema/wapi/internal/infrastructure/json"
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

func TestNewTextHandler(t *testing.T) {
	handler := internalHttp.NewTextHandler(mocksTextHandler(t))
	assert.NotNil(t, handler)
}

func TestSendTextMessageHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name         string
		mocksFactory func(t *testing.T) (*mock.MockAuthorizer, *mock.MockConnections, *jsonInfra.MarshallCallback)
		jsonRequest  func() interface{}
		expectStatus int
	}{
		{
			name:         "OK",
			mocksFactory: mocksTextHandler,
			jsonRequest:  messageRequest,
			expectStatus: http.StatusOK,
		},
		{
			name:         "Bad message request",
			mocksFactory: mocksTextHandler,
			jsonRequest: func() interface{} {
				return ""
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "Connection not found",
			mocksFactory: func(t *testing.T) (*mock.MockAuthorizer, *mock.MockConnections, *jsonInfra.MarshallCallback) {
				authorizer, _, marshal := mocksTextHandler(t)
				c := gomock.NewController(t)
				connections := mock.NewMockConnections(c)
				connections.EXPECT().
					AuthenticatedConnectionForSession(gomock.Any()).
					Return(nil, &service.NotFoundError{})
				return authorizer, connections, marshal
			},
			jsonRequest:  messageRequest,
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "Error message sending",
			mocksFactory: func(t *testing.T) (*mock.MockAuthorizer, *mock.MockConnections, *jsonInfra.MarshallCallback) {
				authorizer, _, marshal := mocksTextHandler(t)
				c := gomock.NewController(t)
				wac := mock.NewMockConn(c)
				wac.EXPECT().Info().Return(&whatsapp.Info{Wid: "wid"})
				wac.EXPECT().Send(gomock.Any()).Return("", errors.New("error image sending"))

				connections := mock.NewMockConnections(c)
				connections.EXPECT().
					AuthenticatedConnectionForSession(gomock.Any()).
					Return(service.NewDTO(wac, &model.WapiSession{}, make(chan string)), nil)

				return authorizer, connections, marshal
			},
			jsonRequest:  messageRequest,
			expectStatus: http.StatusInternalServerError,
		},
		{
			name: "Response marshaling error",
			mocksFactory: func(t *testing.T) (*mock.MockAuthorizer, *mock.MockConnections, *jsonInfra.MarshallCallback) {
				authorizer, connections, _ := mocksTextHandler(t)
				marshal := jsonInfra.MarshallCallback(func(i interface{}) ([]byte, error) {
					return nil, errors.New("marshaling error")
				})
				return authorizer, connections, &marshal
			},
			jsonRequest:  messageRequest,
			expectStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			server := httpTest.New(map[string]internalHttp.AppHTTPHandler{
				"/send-message/": internalHttp.NewTextHandler(tt.mocksFactory(t)),
			})
			defer server.Close()

			expect := httpexpect.New(t, server.URL)
			expect.POST("/send-message/").
				WithJSON(tt.jsonRequest()).
				Expect().
				Status(tt.expectStatus)
		})
	}
}

func mocksTextHandler(t *testing.T) (*mock.MockAuthorizer, *mock.MockConnections, *jsonInfra.MarshallCallback) {
	c := gomock.NewController(t)

	wac := mock.NewMockConn(c)
	wac.EXPECT().Info().Return(&whatsapp.Info{Wid: "wid"})
	wac.EXPECT().Send(gomock.Any()).Return("", nil)

	connections := mock.NewMockConnections(c)
	connections.EXPECT().
		AuthenticatedConnectionForSession(gomock.Any()).
		Return(service.NewDTO(wac, &model.WapiSession{}, make(chan string)), nil)

	marshal := jsonInfra.MarshallCallback(json.Marshal)
	return mock.NewMockAuthorizer(c), connections, &marshal
}

func TestTextHandlerFailWriteResponse(t *testing.T) {
	handler := internalHttp.NewTextHandler(mocksTextHandler(t))
	w := mock.NewFailResponseRecorder(httptest.NewRecorder())
	r, err := http.NewRequest("POST", "/send-message/", bytes.NewReader([]byte("{}")))
	require.Nil(t, err)
	internalHttp.AppHandlerRunner{H: handler}.ServeHTTP(w, r)
	assert.Equal(t, w.Status(), http.StatusInternalServerError)
}

func messageRequest() interface{} {
	return &internalHttp.SendMessageRequest{
		ChatID:    "+000000000000",
		Text:      "hello",
		SessionID: "_sid_",
	}
}
