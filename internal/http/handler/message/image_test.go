package message

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	httpInfra "github.com/r-erema/wapi/internal/infrastructure/http"
	jsonInfra "github.com/r-erema/wapi/internal/infrastructure/json"
	"github.com/r-erema/wapi/internal/model/session"
	"github.com/r-erema/wapi/internal/service/auth"
	"github.com/r-erema/wapi/internal/service/supervisor"
	httpTest "github.com/r-erema/wapi/internal/testutil/http"
	mockAuth "github.com/r-erema/wapi/internal/testutil/mock/auth"
	mockHttp "github.com/r-erema/wapi/internal/testutil/mock/http"
	"github.com/r-erema/wapi/internal/testutil/mock/io"
	mockSupervisor "github.com/r-erema/wapi/internal/testutil/mock/supervisor"
	mockWhatsapp "github.com/r-erema/wapi/internal/testutil/mock/whatsapp"
	"github.com/stretchr/testify/require"

	"github.com/Rhymen/go-whatsapp"
	"github.com/gavv/httpexpect"
	"github.com/golang/mock/gomock"
	"github.com/magiconair/properties/assert"
)

type mocksFactory func(t *testing.T) (auth.Authorizer, supervisor.Connections, httpInfra.Client, *jsonInfra.MarshallCallback)

func TestNewImageHandler(t *testing.T) {
	a, s, c, m := mocks(t)
	imgHandler := NewImageHandler(a, s, c, m)
	assert.Equal(t, imgHandler, &SendImageHandler{
		auth:                  a,
		connectionsSupervisor: s,
		httpClient:            c,
		marshal:               m,
	})
}

type testData struct {
	name string
	mocksFactory
	jsonRequest  func() interface{}
	expectStatus int
}

func ok() testData {
	return testData{
		"OK",
		func(t *testing.T) (auth.Authorizer, supervisor.Connections, httpInfra.Client, *jsonInfra.MarshallCallback) {
			return mocks(t)
		},
		imageRequest,
		http.StatusOK,
	}
}

func badImageRequest() testData {
	return testData{
		"Bad image request",
		func(t *testing.T) (auth.Authorizer, supervisor.Connections, httpInfra.Client, *jsonInfra.MarshallCallback) {
			return mocks(t)
		},
		func() interface{} {
			return ""
		},
		http.StatusBadRequest,
	}
}

func connectionNotFound() testData {
	return testData{
		"Connection not found",
		func(t *testing.T) (auth.Authorizer, supervisor.Connections, httpInfra.Client, *jsonInfra.MarshallCallback) {
			authorizer, _, client, marshal := mocks(t)
			c := gomock.NewController(t)
			connections := mockSupervisor.NewMockConnections(c)
			connections.EXPECT().
				AuthenticatedConnectionForSession(gomock.Any()).
				Return(nil, &supervisor.NotFoundError{})
			return authorizer, connections, client, marshal
		},
		imageRequest,
		http.StatusBadRequest,
	}
}

func badImageURL() testData {
	return testData{
		"Bad image url",
		func(t *testing.T) (auth.Authorizer, supervisor.Connections, httpInfra.Client, *jsonInfra.MarshallCallback) {
			authorizer, connections, _, marshal := mocks(t)
			c := gomock.NewController(t)
			httpClient := mockHttp.NewMockClient(c)
			httpClient.EXPECT().
				Get(gomock.Any()).
				Return(nil, fmt.Errorf("bad image url"))
			return authorizer, connections, httpClient, marshal
		},
		imageRequest,
		http.StatusBadRequest,
	}
}

func cantReadImageBody() testData {
	return testData{
		"Couldn't read image body by url",
		func(t *testing.T) (auth.Authorizer, supervisor.Connections, httpInfra.Client, *jsonInfra.MarshallCallback) {
			authorizer, connections, _, marshal := mocks(t)
			c := gomock.NewController(t)
			httpClient := mockHttp.NewMockClient(c)
			httpClient.EXPECT().
				Get(gomock.Any()).
				Return(&http.Response{Body: ioutil.NopCloser(&io.FailReader{})}, nil)
			return authorizer, connections, httpClient, marshal
		},
		imageRequest,
		http.StatusBadRequest,
	}
}

func errorImageSending() testData {
	return testData{
		"Error image sending",
		func(t *testing.T) (auth.Authorizer, supervisor.Connections, httpInfra.Client, *jsonInfra.MarshallCallback) {
			authorizer, _, httpClient, marshal := mocks(t)
			c := gomock.NewController(t)
			wac := mockWhatsapp.NewMockConn(c)
			wac.EXPECT().Info().Return(&whatsapp.Info{Wid: "wid"})
			wac.EXPECT().Send(gomock.Any()).Return("", errors.New("error image sending"))

			connections := mockSupervisor.NewMockConnections(c)
			connections.EXPECT().
				AuthenticatedConnectionForSession(gomock.Any()).
				Return(supervisor.NewDTO(wac, &session.WapiSession{}), nil)

			return authorizer, connections, httpClient, marshal
		},
		imageRequest,
		http.StatusInternalServerError,
	}
}

func marshalingError() testData {
	return testData{
		"Response marshaling error",
		func(t *testing.T) (
			auth.Authorizer,
			supervisor.Connections,
			httpInfra.Client,
			*jsonInfra.MarshallCallback,
		) {
			authorizer, connections, httpClient, _ := mocks(t)
			marshal := jsonInfra.MarshallCallback(func(i interface{}) ([]byte, error) {
				return nil, errors.New("marshaling error")
			})
			return authorizer, connections, httpClient, &marshal
		},
		imageRequest,
		http.StatusInternalServerError,
	}
}

func TestSendImageHandler(t *testing.T) {
	tests := []struct {
		name string
		mocksFactory
		jsonRequest  func() interface{}
		expectStatus int
	}{
		ok(),
		badImageRequest(),
		connectionNotFound(),
		badImageURL(),
		cantReadImageBody(),
		errorImageSending(),
		marshalingError(),
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			handler := NewImageHandler(tt.mocksFactory(t))
			server := httpTest.New(map[string]http.Handler{"/send-image/": handler})
			defer server.Close()

			expect := httpexpect.New(t, server.URL)
			expect.POST("/send-image/").
				WithJSON(tt.jsonRequest()).
				Expect().
				Status(tt.expectStatus)
		})
	}
}

func TestFailWriteResponse(t *testing.T) {
	handler := NewImageHandler(mocks(t))
	w := mockHttp.NewFailResponseRecorder(httptest.NewRecorder())
	r, err := http.NewRequest("POST", "/send-image/", bytes.NewReader([]byte("{}")))
	require.Nil(t, err)
	handler.ServeHTTP(w, r)
	assert.Equal(t, w.Status(), http.StatusInternalServerError)
}

func mocks(t *testing.T) (
	*mockAuth.MockAuthorizer,
	*mockSupervisor.MockConnections,
	*mockHttp.MockClient,
	*jsonInfra.MarshallCallback,
) {
	c := gomock.NewController(t)

	wac := mockWhatsapp.NewMockConn(c)
	wac.EXPECT().Info().Return(&whatsapp.Info{Wid: "wid"})
	wac.EXPECT().Send(gomock.Any()).Return("", nil)

	connections := mockSupervisor.NewMockConnections(c)
	connections.EXPECT().
		AuthenticatedConnectionForSession(gomock.Any()).
		Return(supervisor.NewDTO(wac, &session.WapiSession{}), nil)

	httpClient := mockHttp.NewMockClient(c)
	httpClient.EXPECT().
		Get(gomock.Any()).
		Return(&http.Response{Body: ioutil.NopCloser(bytes.NewBufferString("{}"))}, nil)

	marshal := jsonInfra.MarshallCallback(json.Marshal)
	return mockAuth.NewMockAuthorizer(c), connections, httpClient, &marshal
}

func imageRequest() interface{} {
	return &SendImageRequest{
		SessionID: "_sid_",
		ChatID:    "+000000000000",
		ImageURL:  "https://img.jpg",
		Caption:   "test image",
	}
}
