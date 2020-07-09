package http

import (
	"errors"
	"testing"

	"github.com/r-erema/wapi/internal/config"
	"github.com/r-erema/wapi/internal/infrastructure/os"
	"github.com/r-erema/wapi/internal/repository"
	"github.com/r-erema/wapi/internal/service"
	"github.com/r-erema/wapi/internal/testutil/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type routerMocksFactory func(t *testing.T) (
	*config.Config,
	repository.Session,
	service.Connections,
	service.Authorizer,
	service.QRFileResolver,
	service.Listener,
	os.FileSystem,
)

func TestRouter(t *testing.T) {
	tests := []struct {
		name         string
		mocksFactory routerMocksFactory
		expectError  bool
	}{
		{
			name:         "OK",
			mocksFactory: routerMocks,
			expectError:  false,
		},
		{
			name: "Creation error",
			mocksFactory: func(t *testing.T) (
				*config.Config,
				repository.Session,
				service.Connections,
				service.Authorizer,
				service.QRFileResolver,
				service.Listener,
				os.FileSystem,
			) {
				conf, _, connSupervisor, authorizer, fileResolver, listener, fs := routerMocks(t)
				c := gomock.NewController(t)
				sessRepo := mock.NewMockSession(c)
				sessRepo.EXPECT().AllSavedSessionIds().Return(nil, errors.New("something went wrong... "))
				return conf, sessRepo, connSupervisor, authorizer, fileResolver, listener, fs
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r, err := Router(tt.mocksFactory(t))
			if tt.expectError {
				assert.Nil(t, r)
				assert.NotNil(t, err)
			} else {
				assert.NotNil(t, r)
				assert.Nil(t, err)
			}
		})
	}
}

func routerMocks(t *testing.T) (
	*config.Config,
	repository.Session,
	service.Connections,
	service.Authorizer,
	service.QRFileResolver,
	service.Listener,
	os.FileSystem,
) {
	conf := &config.Config{
		ListenHTTPHost:              "localhost",
		ConnectionTimeout:           10,
		FileSystemRootPath:          "/tmp",
		WebHookURL:                  "/wh/",
		RedisHost:                   "localhost:6379",
		Env:                         config.DevMode,
		CertFilePath:                "/tmp/cert.crt",
		CertKeyPath:                 "/tmp/cert.key",
		SentryDSN:                   "dsn@sentry.io/test",
		ConnectionsCheckoutDuration: 60,
	}
	c := gomock.NewController(t)
	sessRepo := mock.NewMockSession(c)
	sessRepo.EXPECT().AllSavedSessionIds().Return(nil, nil)
	return conf,
		sessRepo,
		mock.NewMockConnections(c),
		mock.NewMockAuthorizer(c),
		mock.NewMockQRFileResolver(c),
		mock.NewMockListener(c),
		mock.NewMockFileSystem(c)
}
