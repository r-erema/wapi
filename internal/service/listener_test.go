package service_test

import (
	"errors"
	"os"
	"sync"
	"testing"

	httpInfra "github.com/r-erema/wapi/internal/infrastructure/http"
	"github.com/r-erema/wapi/internal/model"
	"github.com/r-erema/wapi/internal/repository"
	"github.com/r-erema/wapi/internal/service"
	"github.com/r-erema/wapi/internal/testutil/mock"

	"github.com/Rhymen/go-whatsapp"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type listenerTestData struct {
	name            string
	mocksFactory    listenerMocksFactory
	ignoreInterrupt bool
	waitErr         bool
}
type listenerMocksFactory func(t *testing.T) (
	repository.Session,
	service.Connections,
	service.Authorizer,
	string,
	repository.Message,
	httpInfra.Client,
	chan os.Signal,
)

func TestNewWebHook(t *testing.T) {
	listener := service.NewWebHook(listenerMocks(t))
	assert.IsType(t, listener, &service.WebHook{})
}

func TestWebHook_ListenForSession(t *testing.T) {
	tests := []listenerTestData{
		{
			name:            "OK",
			mocksFactory:    listenerMocks,
			ignoreInterrupt: false,
			waitErr:         false,
		},
		{
			name: "Session is already listening",
			mocksFactory: func(t *testing.T) (
				repository.Session,
				service.Connections,
				service.Authorizer,
				string, repository.Message,
				httpInfra.Client,
				chan os.Signal,
			) {
				sessRepo, _, auth, wh, msgRepo, client, interruptCh := listenerMocks(t)
				c := gomock.NewController(t)
				connSV := mock.NewMockConnections(c)
				connSV.EXPECT().AuthenticatedConnectionForSession(gomock.Any()).Return(nil, nil)
				return sessRepo, connSV, auth, wh, msgRepo, client, interruptCh
			},
			ignoreInterrupt: true,
			waitErr:         true,
		},
		{
			name: "Login failed",
			mocksFactory: func(t *testing.T) (
				repository.Session,
				service.Connections,
				service.Authorizer,
				string, repository.Message,
				httpInfra.Client,
				chan os.Signal,
			) {
				sessRepo, connSV, _, wh, msgRepo, client, interruptCh := listenerMocks(t)
				c := gomock.NewController(t)
				auth := mock.NewMockAuthorizer(c)
				auth.EXPECT().Login(gomock.Any()).Return(nil, nil, errors.New("login failed"))
				return sessRepo, connSV, auth, wh, msgRepo, client, interruptCh
			},
			ignoreInterrupt: true,
			waitErr:         true,
		},
		{
			name: "Disconnect failed",
			mocksFactory: func(t *testing.T) (
				repository.Session,
				service.Connections,
				service.Authorizer,
				string, repository.Message,
				httpInfra.Client,
				chan os.Signal,
			) {
				sessRepo, connSV, _, wh, msgRepo, client, interruptCh := listenerMocks(t)

				c := gomock.NewController(t)

				conn := mock.NewMockConn(c)
				conn.EXPECT().AddHandler(gomock.Any())
				conn.EXPECT().Disconnect().Return(whatsapp.Session{}, errors.New("disconnect error"))

				sess := &model.WapiSession{SessionID: "_sid_", WhatsAppSession: &whatsapp.Session{Wid: "_wid_"}}

				auth := mock.NewMockAuthorizer(c)
				auth.EXPECT().Login(gomock.Any()).Return(conn, sess, nil)

				return sessRepo, connSV, auth, wh, msgRepo, client, interruptCh
			},
			ignoreInterrupt: false,
			waitErr:         true,
		},
		{
			name: "Session writing failed",
			mocksFactory: func(t *testing.T) (
				repository.Session,
				service.Connections,
				service.Authorizer,
				string, repository.Message,
				httpInfra.Client,
				chan os.Signal,
			) {
				_, connSV, auth, wh, msgRepo, client, interruptCh := listenerMocks(t)
				c := gomock.NewController(t)
				sessRepo := mock.NewMockSession(c)
				sessRepo.EXPECT().WriteSession(gomock.Any()).Return(errors.New("writing error"))
				return sessRepo, connSV, auth, wh, msgRepo, client, interruptCh
			},
			ignoreInterrupt: false,
			waitErr:         true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			sessRepo, connSV, auth, wh, msgRepo, client, interruptCh := tt.mocksFactory(t)
			listener := service.NewWebHook(sessRepo, connSV, auth, wh, msgRepo, client, interruptCh)
			var err error
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {

				if tt.ignoreInterrupt {
					<-interruptCh
				}

				testWg := &sync.WaitGroup{}
				testWg.Add(1)
				_, err = listener.ListenForSession("_sid_", testWg)

				wg.Done()
			}()
			interruptCh <- os.Interrupt
			wg.Wait()

			if tt.waitErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func listenerMocks(t *testing.T) (
	repository.Session,
	service.Connections,
	service.Authorizer,
	string,
	repository.Message,
	httpInfra.Client,
	chan os.Signal,
) {
	c := gomock.NewController(t)

	sessRepo := mock.NewMockSession(c)
	sessRepo.EXPECT().WriteSession(gomock.Any()).Return(nil)

	connSupervisor := mock.NewMockConnections(c)
	connSupervisor.EXPECT().AuthenticatedConnectionForSession(gomock.Any()).Return(nil, &service.NotFoundError{})

	conn := mock.NewMockConn(c)
	conn.EXPECT().AddHandler(gomock.Any())
	conn.EXPECT().Disconnect().Return(whatsapp.Session{}, nil)

	sess := &model.WapiSession{SessionID: "_sid_", WhatsAppSession: &whatsapp.Session{Wid: "_wid_"}}

	auth := mock.NewMockAuthorizer(c)
	auth.EXPECT().Login(gomock.Any()).Return(conn, sess, nil)

	return sessRepo,
		connSupervisor,
		auth,
		"/webhook_url/",
		mock.NewMockMessage(c),
		mock.NewMockClient(c),
		make(chan os.Signal)
}
