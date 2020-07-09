package service_test

import (
	"errors"
	"testing"

	"github.com/r-erema/wapi/internal/config"
	infraWA "github.com/r-erema/wapi/internal/infrastructure/whatsapp"
	"github.com/r-erema/wapi/internal/model"
	"github.com/r-erema/wapi/internal/repository"
	"github.com/r-erema/wapi/internal/service"
	"github.com/r-erema/wapi/internal/testutil/mock"

	"github.com/Rhymen/go-whatsapp"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	sessRepo, connections, fileResolver, connector := authMocks(t)
	a := service.NewAuth(config.DefaultConnectionTimeout, sessRepo, connections, fileResolver, connector, make(chan string))
	assert.NotNil(t, a)
}

type authTestData struct {
	name         string
	mocksFactory authMocksFactory
	waitErr      bool
}
type authMocksFactory func(t *testing.T) (
	repository.Session,
	service.Connections,
	service.QRFileResolver,
	service.Connector,
)

func okLoginBySession() authTestData {
	return authTestData{
		name:         "Successful login by restoring session",
		mocksFactory: authMocks,
		waitErr:      false,
	}
}

func okLoginByQR() authTestData {
	return authTestData{
		name: "Successful login by qr code scanning",
		mocksFactory: func(t *testing.T) (repository.Session, service.Connections, service.QRFileResolver, service.Connector) {
			_, connections, fileResolver, _ := authMocks(t)

			c := gomock.NewController(t)

			connection := mock.NewMockConn(c)
			connection.EXPECT().Login(gomock.Any()).
				DoAndReturn(func(qrChan chan<- string) (whatsapp.Session, error) {
					qrChan <- "qr data"
					return whatsapp.Session{}, nil
				})

			connector := mock.NewMockConnector(c)
			connector.EXPECT().Connect(gomock.Any()).Return(connection, nil)

			sessRepo := mock.NewMockSession(c)
			sessRepo.EXPECT().ReadSession(gomock.Any()).Return(nil, errors.New("couldn't read session"))
			sessRepo.EXPECT().WriteSession(gomock.Any()).Return(nil)

			return sessRepo, connections, fileResolver, connector
		},
		waitErr: false,
	}
}

func connectionFailed() authTestData {
	return authTestData{
		name: "Connection failed",
		mocksFactory: func(t *testing.T) (repository.Session, service.Connections, service.QRFileResolver, service.Connector) {
			sessRepo, connections, fileResolver, _ := authMocks(t)

			c := gomock.NewController(t)
			connector := mock.NewMockConnector(c)
			connector.EXPECT().Connect(gomock.Any()).Return(nil, errors.New("connection failed"))
			return sessRepo, connections, fileResolver, connector
		},
		waitErr: true,
	}
}

func sessionRestoringFailed() authTestData {
	return authTestData{
		name: "Session restoring failed",
		mocksFactory: func(t *testing.T) (repository.Session, service.Connections, service.QRFileResolver, service.Connector) {
			_, connections, fileResolver, _ := authMocks(t)

			c := gomock.NewController(t)

			connection := mock.NewMockConn(c)
			connection.EXPECT().RestoreWithSession(gomock.Any()).Return(whatsapp.Session{}, errors.New(infraWA.ErrMsg401))

			connector := mock.NewMockConnector(c)
			connector.EXPECT().Connect(gomock.Any()).Return(connection, nil)

			sessRepo := mock.NewMockSession(c)
			sessRepo.EXPECT().ReadSession(gomock.Any()).Return(&model.WapiSession{}, nil)
			sessRepo.EXPECT().RemoveSession(gomock.Any()).Return(nil)

			return sessRepo, connections, fileResolver, connector
		},
		waitErr: true,
	}
}

func loginByQRFailed() authTestData {
	return authTestData{
		name: "Login by QR-code failed",
		mocksFactory: func(t *testing.T) (repository.Session, service.Connections, service.QRFileResolver, service.Connector) {
			_, connections, fileResolver, _ := authMocks(t)

			c := gomock.NewController(t)

			connection := mock.NewMockConn(c)
			connection.EXPECT().Login(gomock.Any()).DoAndReturn(func(qrChan chan<- string) (whatsapp.Session, error) {
				qrChan <- "qr data"
				return whatsapp.Session{}, errors.New("login failed")
			})

			connector := mock.NewMockConnector(c)
			connector.EXPECT().Connect(gomock.Any()).Return(connection, nil)

			sessRepo := mock.NewMockSession(c)
			sessRepo.EXPECT().ReadSession(gomock.Any()).Return(nil, errors.New("couldn't read session"))

			return sessRepo, connections, fileResolver, connector
		},
		waitErr: true,
	}
}

func supervisorFailed() authTestData {
	return authTestData{
		name: "Adding connection to supervisor failed",
		mocksFactory: func(t *testing.T) (repository.Session, service.Connections, service.QRFileResolver, service.Connector) {
			sessRepo, _, fileResolver, connector := authMocks(t)
			c := gomock.NewController(t)
			connections := mock.NewMockConnections(c)
			connections.EXPECT().AddAuthenticatedConnectionForSession(gomock.Any(), gomock.Any()).Return(errors.New("something went wrong"))
			return sessRepo, connections, fileResolver, connector
		},
		waitErr: true,
	}
}

func sessionStoringFail() authTestData {
	return authTestData{
		name: "Session storing fail",
		mocksFactory: func(t *testing.T) (repository.Session, service.Connections, service.QRFileResolver, service.Connector) {
			_, connections, fileResolver, connector := authMocks(t)
			c := gomock.NewController(t)
			sessRepo := mock.NewMockSession(c)
			sessRepo.EXPECT().ReadSession(gomock.Any()).Return(&model.WapiSession{}, nil)
			sessRepo.EXPECT().WriteSession(gomock.Any()).Return(errors.New("something went wrong"))
			return sessRepo, connections, fileResolver, connector
		},
		waitErr: true,
	}
}

func TestAuth_Login(t *testing.T) {
	tests := []authTestData{
		okLoginBySession(),
		okLoginByQR(),
		connectionFailed(),
		sessionRestoringFailed(),
		loginByQRFailed(),
		supervisorFailed(),
		sessionStoringFail(),
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			sessRepo, connections, fileResolver, connector := tt.mocksFactory(t)
			a := service.NewAuth(
				config.DefaultConnectionTimeout,
				sessRepo,
				connections,
				fileResolver,
				connector,
				make(chan string),
			)
			conn, sess, err := a.Login("_sid_")
			if tt.waitErr {
				assert.Nil(t, conn)
				assert.Nil(t, sess)
				assert.NotNil(t, err)
			} else {
				assert.NotNil(t, conn)
				assert.NotNil(t, sess)
				assert.Nil(t, err)
			}
		})
	}
}

func authMocks(t *testing.T) (repository.Session, service.Connections, service.QRFileResolver, service.Connector) {
	c := gomock.NewController(t)

	connection := mock.NewMockConn(c)
	connection.EXPECT().RestoreWithSession(gomock.Any()).Return(whatsapp.Session{}, nil)

	connector := mock.NewMockConnector(c)
	connector.EXPECT().Connect(gomock.Any()).Return(connection, nil)

	sessRepo := mock.NewMockSession(c)
	sessRepo.EXPECT().ReadSession(gomock.Any()).Return(&model.WapiSession{}, nil)
	sessRepo.EXPECT().WriteSession(gomock.Any()).Return(nil)

	connections := mock.NewMockConnections(c)
	connections.EXPECT().AddAuthenticatedConnectionForSession(gomock.Any(), gomock.Any()).Return(nil)

	fileResolver := mock.NewMockQRFileResolver(c)
	fileResolver.EXPECT().ResolveQrFilePath(gomock.Any()).AnyTimes().Return("")

	return sessRepo, connections, fileResolver, connector
}
