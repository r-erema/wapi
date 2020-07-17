package http_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	internalHttp "github.com/r-erema/wapi/internal/http"
	"github.com/r-erema/wapi/internal/service"
	httpTest "github.com/r-erema/wapi/internal/testutil/http"
	"github.com/r-erema/wapi/internal/testutil/mock"

	"github.com/gavv/httpexpect"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQR(t *testing.T) {
	fs, fileResolver := qrMocks(t)
	handler := internalHttp.NewQR(fs, fileResolver)
	assert.NotNil(t, handler)
}

func TestImageHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name         string
		mocksFactory func(t *testing.T) (*mock.MockFileSystem, service.QRFileResolver)
		expectStatus int
	}{
		{
			name:         "OK",
			mocksFactory: qrMocks,
			expectStatus: http.StatusOK,
		},
		{
			name: "QR image file not fount",
			mocksFactory: func(t *testing.T) (*mock.MockFileSystem, service.QRFileResolver) {
				c := gomock.NewController(t)
				_, resolver := qrMocks(t)
				fs := mock.NewMockFileSystem(c)
				fs.EXPECT().Stat(gomock.Any()).Return(nil, nil)
				fs.EXPECT().IsNotExist(gomock.Any()).Return(true)
				return fs, resolver
			},
			expectStatus: http.StatusNotFound,
		},
		{
			name: "Couldn't open QR image file",
			mocksFactory: func(t *testing.T) (*mock.MockFileSystem, service.QRFileResolver) {
				c := gomock.NewController(t)
				_, resolver := qrMocks(t)
				fs := mock.NewMockFileSystem(c)
				fs.EXPECT().Stat(gomock.Any()).Return(nil, nil)
				fs.EXPECT().IsNotExist(gomock.Any()).Return(false)
				fs.EXPECT().Open(gomock.Any()).Return(nil, errors.New("can't open file"))
				return fs, resolver
			},
			expectStatus: http.StatusInternalServerError,
		},
		{
			name: "Couldn't copy QR image file to buffer",
			mocksFactory: func(t *testing.T) (*mock.MockFileSystem, service.QRFileResolver) {
				c := gomock.NewController(t)
				_, resolver := qrMocks(t)
				fs := mock.NewMockFileSystem(c)
				fs.EXPECT().Stat(gomock.Any()).Return(nil, nil)
				fs.EXPECT().IsNotExist(gomock.Any()).Return(false)

				file := mock.NewMockFile(c)
				file.EXPECT().Read(gomock.Any()).Return(0, io.ErrUnexpectedEOF)
				fs.EXPECT().Open(gomock.Any()).Return(file, nil)
				return fs, resolver
			},
			expectStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			handler := internalHttp.NewQR(tt.mocksFactory(t))
			server := httpTest.New(map[string]internalHttp.AppHTTPHandler{"/get-qr-code/{sessionID}/": handler})
			defer server.Close()

			expect := httpexpect.New(t, server.URL)
			expect.GET("/get-qr-code/{sessionID}/").
				Expect().
				Status(tt.expectStatus)
		})
	}

	fs, fileResolver := qrMocks(t)
	handler := internalHttp.NewQR(fs, fileResolver)
	server := httpTest.New(map[string]internalHttp.AppHTTPHandler{"/get-qr-code/{sessionID}/": handler})
	defer server.Close()

	expect := httpexpect.New(t, server.URL)
	expect.GET("/get-qr-code/{sessionID}/", "_sid_").
		Expect().
		Status(http.StatusOK)
}

func TestQRHandlerFailWriteResponse(t *testing.T) {
	handler := internalHttp.NewQR(qrMocks(t))
	w := mock.NewFailResponseRecorder(httptest.NewRecorder())
	r, err := http.NewRequest("GET", "/get-qr-code/_sid_/", nil)
	require.Nil(t, err)
	internalHttp.AppHandlerRunner{H: handler}.ServeHTTP(w, r)
	assert.Equal(t, w.Status(), http.StatusInternalServerError)
}

func qrMocks(t *testing.T) (*mock.MockFileSystem, service.QRFileResolver) {
	c := gomock.NewController(t)
	fs := mock.NewMockFileSystem(c)
	fs.EXPECT().Stat(gomock.Any()).MaxTimes(2).Return(nil, nil)
	fs.EXPECT().IsNotExist(gomock.Any()).MaxTimes(2).Return(false)

	file := mock.NewMockFile(c)
	file.EXPECT().Read(gomock.Any()).Return(0, io.EOF)
	fs.EXPECT().Open(gomock.Any()).Return(file, nil)

	fileResolver, err := service.NewQRImgResolver("/fake/path", fs)
	require.Nil(t, err)
	return fs, fileResolver
}
