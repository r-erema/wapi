package service_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	httpInfra "github.com/r-erema/wapi/internal/infrastructure/http"
	jsonInfra "github.com/r-erema/wapi/internal/infrastructure/json"
	infraWA "github.com/r-erema/wapi/internal/infrastructure/whatsapp"
	"github.com/r-erema/wapi/internal/model"
	"github.com/r-erema/wapi/internal/repository"
	"github.com/r-erema/wapi/internal/service"
	"github.com/r-erema/wapi/internal/testutil/mock"

	"github.com/Rhymen/go-whatsapp"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type msgTestData struct {
	name         string
	mocksFactory msgMocksFactory
	msg          *whatsapp.TextMessage
}

type msgHandleErrData struct {
	name         string
	mocksFactory msgMocksFactory
	err          error
}

type msgMocksFactory func(t *testing.T) (
	infraWA.Conn,
	*model.WapiSession,
	repository.Message,
	service.Connections,
	repository.Session,
	httpInfra.Client,
	*jsonInfra.MarshallCallback,
	uint64,
	string,
)

func TestNewHandler(t *testing.T) {
	msgHandler := service.NewMsgHandler(msgMocks(t))
	assert.IsType(t, msgHandler, &service.Handler{})
}

func TestHandleError(t *testing.T) {
	tests := []msgHandleErrData{
		{
			name:         "Connection closed",
			mocksFactory: msgMocks,
			err:          &whatsapp.ErrConnectionClosed{},
		},
		{
			name: "Restore closed connection failed",
			mocksFactory: func(t *testing.T) (
				infraWA.Conn,
				*model.WapiSession,
				repository.Message,
				service.Connections,
				repository.Session,
				httpInfra.Client,
				*jsonInfra.MarshallCallback,
				uint64,
				string,
			) {
				_, sess, msgRepo, connSV, sessRepo, client, marshal, time, wh := msgMocks(t)

				c := gomock.NewController(t)
				conn := mock.NewMockConn(c)
				conn.EXPECT().Info().Return(&whatsapp.Info{})
				conn.EXPECT().AdminTest().Return(true, nil)
				conn.EXPECT().RestoreWithSession(gomock.Any()).Return(whatsapp.Session{}, errors.New("something went wrong... "))

				return conn, sess, msgRepo, connSV, sessRepo, client, marshal, time, wh
			},
			err: &whatsapp.ErrConnectionClosed{},
		},
		{
			name: "Restore closed connection ok",
			mocksFactory: func(t *testing.T) (
				infraWA.Conn,
				*model.WapiSession,
				repository.Message,
				service.Connections,
				repository.Session,
				httpInfra.Client,
				*jsonInfra.MarshallCallback,
				uint64,
				string,
			) {
				_, sess, msgRepo, connSV, sessRepo, client, marshal, time, wh := msgMocks(t)

				c := gomock.NewController(t)
				conn := mock.NewMockConn(c)
				conn.EXPECT().Info().Return(&whatsapp.Info{})
				conn.EXPECT().AdminTest().Return(true, nil)
				conn.EXPECT().RestoreWithSession(gomock.Any()).Return(whatsapp.Session{}, nil)

				return conn, sess, msgRepo, connSV, sessRepo, client, marshal, time, wh
			},
			err: &whatsapp.ErrConnectionClosed{},
		},
		{
			name:         "Connection failed",
			mocksFactory: msgMocks,
			err:          &whatsapp.ErrConnectionFailed{},
		},
		{
			name:         "Unknown error",
			mocksFactory: msgMocks,
			err:          errors.New("some error"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			h := service.NewMsgHandler(tt.mocksFactory(t))
			h.HandleError(tt.err)
		})
	}
}

func sendMsgOk() msgTestData {
	return msgTestData{
		name:         "Send msg OK",
		mocksFactory: msgMocks,
		msg: &whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{Timestamp: 22, RemoteJid: "+000000000001"},
		},
	}
}

func msgHasWrongTimestamp() msgTestData {
	return msgTestData{
		name: "Message has wrong timestamp",
		mocksFactory: func(t *testing.T) (
			infraWA.Conn,
			*model.WapiSession,
			repository.Message,
			service.Connections,
			repository.Session,
			httpInfra.Client,
			*jsonInfra.MarshallCallback,
			uint64,
			string,
		) {
			conn, sess, msgRepo, connSV, sessRepo, client, marshal, _, wh := msgMocks(t)
			return conn, sess, msgRepo, connSV, sessRepo, client, marshal, 15, wh
		},
		msg: &whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{Timestamp: 111, RemoteJid: "+000000000000"},
		},
	}
}

func msgAlreadySent() msgTestData {
	return msgTestData{
		name: "Message already sent",
		mocksFactory: func(t *testing.T) (
			infraWA.Conn,
			*model.WapiSession,
			repository.Message,
			service.Connections,
			repository.Session,
			httpInfra.Client,
			*jsonInfra.MarshallCallback,
			uint64,
			string,
		) {
			conn, sess, _, connSV, sessRepo, client, marshal, _, wh := msgMocks(t)
			c := gomock.NewController(t)
			msgRepo := mock.NewMockMessage(c)
			msgRepo.EXPECT().MessageTime(gomock.Any()).Return(nil, nil)
			return conn, sess, msgRepo, connSV, sessRepo, client, marshal, 0, wh
		},
		msg: &whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{Timestamp: 112, RemoteJid: "+000000000000"},
		},
	}
}

func dontHandleFromMeMsg() msgTestData {
	return msgTestData{
		name: "Don't handle `from me` message",
		mocksFactory: func(t *testing.T) (
			infraWA.Conn,
			*model.WapiSession,
			repository.Message,
			service.Connections,
			repository.Session,
			httpInfra.Client,
			*jsonInfra.MarshallCallback,
			uint64,
			string,
		) {
			conn, sess, msgRepo, connSV, sessRepo, client, marshal, _, wh := msgMocks(t)
			return conn, sess, msgRepo, connSV, sessRepo, client, marshal, 7, wh
		},
		msg: &whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{Timestamp: 8, RemoteJid: "+000000000000", FromMe: true},
		},
	}
}

func msgMarshalingError() msgTestData {
	return msgTestData{
		name: "Message marshaling error",
		mocksFactory: func(t *testing.T) (
			infraWA.Conn,
			*model.WapiSession,
			repository.Message,
			service.Connections,
			repository.Session,
			httpInfra.Client,
			*jsonInfra.MarshallCallback,
			uint64,
			string,
		) {
			conn, sess, msgRepo, connSV, sessRepo, client, _, _, wh := msgMocks(t)
			marshal := jsonInfra.MarshallCallback(func(i interface{}) ([]byte, error) {
				return nil, errors.New("marshaling error")
			})
			return conn, sess, msgRepo, connSV, sessRepo, client, &marshal, 1, wh
		},
		msg: &whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{Timestamp: 2, RemoteJid: "+000000000000"},
		},
	}
}

func msgSendToBebHookError() msgTestData {
	return msgTestData{
		name: "Message sending to webhook error",
		mocksFactory: func(t *testing.T) (
			infraWA.Conn,
			*model.WapiSession,
			repository.Message,
			service.Connections,
			repository.Session,
			httpInfra.Client,
			*jsonInfra.MarshallCallback,
			uint64,
			string,
		) {
			conn, sess, msgRepo, connSV, sessRepo, _, marshal, _, wh := msgMocks(t)
			c := gomock.NewController(t)
			client := mock.NewMockClient(c)
			client.EXPECT().
				Post(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("something went wrong... "))
			return conn, sess, msgRepo, connSV, sessRepo, client, marshal, 10, wh
		},
		msg: &whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{Timestamp: 22, RemoteJid: "+000000000000"},
		},
	}
}

func msgSavingError() msgTestData {
	return msgTestData{
		name: "Message saving error",
		mocksFactory: func(t *testing.T) (
			infraWA.Conn,
			*model.WapiSession,
			repository.Message,
			service.Connections,
			repository.Session,
			httpInfra.Client,
			*jsonInfra.MarshallCallback,
			uint64,
			string,
		) {
			conn, sess, _, connSV, sessRepo, client, marshal, _, wh := msgMocks(t)
			c := gomock.NewController(t)
			msgRepo := mock.NewMockMessage(c)
			msgRepo.EXPECT().SaveMessageTime(gomock.Any(), gomock.Any()).Return(errors.New("saving error"))
			msgRepo.EXPECT().MessageTime(gomock.Any()).Return(nil, errors.New("message not found"))
			return conn, sess, msgRepo, connSV, sessRepo, client, marshal, 100, wh
		},
		msg: &whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{Timestamp: 200, RemoteJid: "+000000000000"},
		},
	}
}

func TestHandleTextMessage(t *testing.T) {
	tests := []msgTestData{
		sendMsgOk(),
		msgHasWrongTimestamp(),
		msgAlreadySent(),
		dontHandleFromMeMsg(),
		msgMarshalingError(),
		msgSendToBebHookError(),
		msgSavingError(),
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			h := service.NewMsgHandler(tt.mocksFactory(t))
			h.HandleTextMessage(tt.msg)
		})
	}
}

func msgMocks(t *testing.T) (
	conn infraWA.Conn,
	sess *model.WapiSession,
	msgRepo repository.Message,
	connSupervisor service.Connections,
	sessRepo repository.Session,
	client httpInfra.Client,
	marshal *jsonInfra.MarshallCallback,
	_ uint64,
	_ string,
) {
	c := gomock.NewController(t)

	connMock := mock.NewMockConn(c)
	connMock.EXPECT().AdminTest().Return(false, nil)
	connMock.EXPECT().Info().Return(&whatsapp.Info{})
	conn = connMock

	sess = &model.WapiSession{WhatsAppSession: &whatsapp.Session{}}

	msgRepoMock := mock.NewMockMessage(c)
	msgRepoMock.EXPECT().MessageTime(gomock.Any()).Return(nil, errors.New("message not found"))
	msgRepoMock.EXPECT().SaveMessageTime(gomock.Any(), gomock.Any()).Return(nil)
	msgRepo = msgRepoMock

	connSupervisorMock := mock.NewMockConnections(c)
	connSupervisorMock.EXPECT().RemoveConnectionForSession(gomock.Any())
	connSupervisor = connSupervisorMock

	sessRepoMock := mock.NewMockSession(c)
	sessRepoMock.EXPECT().RemoveSession(gomock.Any())
	sessRepo = sessRepoMock

	clientMock := mock.NewMockClient(c)
	clientMock.EXPECT().
		Post(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(""))}, nil)
	client = clientMock

	m := jsonInfra.MarshallCallback(json.Marshal)
	marshal = &m
	return conn, sess, msgRepo, connSupervisor, sessRepo, client, marshal, 0, "webhook/url"
}
