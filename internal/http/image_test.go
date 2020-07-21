package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	internalHttp "github.com/r-erema/wapi/internal/http"
	httpInfra "github.com/r-erema/wapi/internal/infrastructure/http"
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

type imagesMocksFactory func(t *testing.T) (service.Authorizer, service.Connections, httpInfra.Client, *jsonInfra.MarshallCallback)

func TestNewImageHandler(t *testing.T) {
	imgHandler := internalHttp.NewImageHandler(mocks(t))
	assert.NotNil(t, imgHandler)
}

type testData struct {
	name string
	imagesMocksFactory
	jsonRequest  func() interface{}
	expectStatus int
}

func ok() testData {
	return testData{
		"OK",
		func(t *testing.T) (service.Authorizer, service.Connections, httpInfra.Client, *jsonInfra.MarshallCallback) {
			return mocks(t)
		},
		imageRequest,
		http.StatusOK,
	}
}

func badImageRequest() testData {
	return testData{
		name: "Bad image request",
		imagesMocksFactory: func(t *testing.T) (service.Authorizer, service.Connections, httpInfra.Client, *jsonInfra.MarshallCallback) {
			return mocks(t)
		},
		jsonRequest: func() interface{} {
			return ""
		},
		expectStatus: http.StatusBadRequest,
	}
}

func connectionNotFound() testData {
	return testData{
		name: "Connection not found",
		imagesMocksFactory: func(t *testing.T) (service.Authorizer, service.Connections, httpInfra.Client, *jsonInfra.MarshallCallback) {
			authorizer, _, client, marshal := mocks(t)
			c := gomock.NewController(t)
			connections := mock.NewMockConnections(c)
			connections.EXPECT().
				AuthenticatedConnectionForSession(gomock.Any()).
				Return(nil, &service.NotFoundError{})
			return authorizer, connections, client, marshal
		},
		jsonRequest:  imageRequest,
		expectStatus: http.StatusBadRequest,
	}
}

func badImageURL() testData {
	return testData{
		name: "Bad image url",
		imagesMocksFactory: func(t *testing.T) (service.Authorizer, service.Connections, httpInfra.Client, *jsonInfra.MarshallCallback) {
			authorizer, connections, _, marshal := mocks(t)
			c := gomock.NewController(t)
			httpClient := mock.NewMockClient(c)
			httpClient.EXPECT().
				Get(gomock.Any()).
				Return(nil, fmt.Errorf("bad image url"))
			return authorizer, connections, httpClient, marshal
		},
		jsonRequest:  imageRequest,
		expectStatus: http.StatusInternalServerError,
	}
}

func cantReadImageBody() testData {
	return testData{
		name: "Couldn't read image body by url",
		imagesMocksFactory: func(t *testing.T) (service.Authorizer, service.Connections, httpInfra.Client, *jsonInfra.MarshallCallback) {
			authorizer, connections, _, marshal := mocks(t)
			c := gomock.NewController(t)
			httpClient := mock.NewMockClient(c)
			httpClient.EXPECT().
				Get(gomock.Any()).
				Return(&http.Response{Body: ioutil.NopCloser(&mock.FailReader{})}, nil)
			return authorizer, connections, httpClient, marshal
		},
		jsonRequest:  imageRequest,
		expectStatus: http.StatusInternalServerError,
	}
}

func errorImageSending() testData {
	return testData{
		name: "Error image sending",
		imagesMocksFactory: func(t *testing.T) (service.Authorizer, service.Connections, httpInfra.Client, *jsonInfra.MarshallCallback) {
			authorizer, _, httpClient, marshal := mocks(t)
			c := gomock.NewController(t)
			wac := mock.NewMockConn(c)
			wac.EXPECT().Info().Return(&whatsapp.Info{Wid: "wid"})
			wac.EXPECT().Send(gomock.Any()).Return("", errors.New("error image sending"))

			connections := mock.NewMockConnections(c)
			connections.EXPECT().
				AuthenticatedConnectionForSession(gomock.Any()).
				Return(service.NewDTO(wac, &model.WapiSession{}, make(chan string)), nil)

			return authorizer, connections, httpClient, marshal
		},
		jsonRequest:  imageRequest,
		expectStatus: http.StatusInternalServerError,
	}
}

func marshalingError() testData {
	return testData{
		name: "Response marshaling error",
		imagesMocksFactory: func(t *testing.T) (
			service.Authorizer,
			service.Connections,
			httpInfra.Client,
			*jsonInfra.MarshallCallback,
		) {
			authorizer, connections, httpClient, _ := mocks(t)
			marshal := jsonInfra.MarshallCallback(func(i interface{}) ([]byte, error) {
				return nil, errors.New("marshaling error")
			})
			return authorizer, connections, httpClient, &marshal
		},
		jsonRequest:  imageRequest,
		expectStatus: http.StatusInternalServerError,
	}
}

func TestSendImageHandler(t *testing.T) {
	tests := []testData{
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
			handler := internalHttp.NewImageHandler(tt.imagesMocksFactory(t))
			server := httpTest.New(map[string]internalHttp.AppHTTPHandler{"/send-image/": handler})
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
	handler := internalHttp.NewImageHandler(mocks(t))
	w := mock.NewFailResponseRecorder(httptest.NewRecorder())
	r, err := http.NewRequest("POST", "/send-image/", bytes.NewReader([]byte("{}")))
	require.Nil(t, err)
	internalHttp.AppHandlerRunner{H: handler}.ServeHTTP(w, r)
	assert.Equal(t, w.Status(), http.StatusInternalServerError)
}

func mocks(t *testing.T) (
	*mock.MockAuthorizer,
	*mock.MockConnections,
	*mock.MockClient,
	*jsonInfra.MarshallCallback,
) {
	c := gomock.NewController(t)

	wac := mock.NewMockConn(c)
	wac.EXPECT().Info().Return(&whatsapp.Info{Wid: "wid"})
	wac.EXPECT().Send(gomock.Any()).Return("", nil)

	connections := mock.NewMockConnections(c)
	connections.EXPECT().
		AuthenticatedConnectionForSession(gomock.Any()).
		Return(service.NewDTO(wac, &model.WapiSession{}, make(chan string)), nil)

	httpClient := mock.NewMockClient(c)
	httpClient.EXPECT().
		Get(gomock.Any()).
		Return(&http.Response{Body: ioutil.NopCloser(bytes.NewBufferString("{}"))}, nil)

	marshal := jsonInfra.MarshallCallback(json.Marshal)
	return mock.NewMockAuthorizer(c), connections, httpClient, &marshal
}

func imageRequest() interface{} {
	return &internalHttp.SendImageRequest{
		SessionID: "_sid_",
		ChatID:    "+000000000000",
		ImageURL:  "https://img.jpg",
		Caption:   "test image",
	}
}
