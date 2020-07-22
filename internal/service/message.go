package service

import (
	"bytes"
	"fmt"
	"log"
	"time"

	httpInfra "github.com/r-erema/wapi/internal/infrastructure/http"
	jsonInfra "github.com/r-erema/wapi/internal/infrastructure/json"
	infrastructureWhatsapp "github.com/r-erema/wapi/internal/infrastructure/whatsapp"
	"github.com/r-erema/wapi/internal/model"
	"github.com/r-erema/wapi/internal/repository"

	"github.com/Rhymen/go-whatsapp"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

// Sentry flush timout
const SentryFlushTimeoutSeconds = 5

// Handler is responsible handle incoming messages and errors.
type Handler struct {
	Connection            infrastructureWhatsapp.Conn
	Session               *model.WapiSession
	messageRepo           repository.Message
	connectionsSupervisor Connections
	storedSession         repository.Session
	client                httpInfra.Client
	marshal               *jsonInfra.MarshallCallback
	InitTimestamp         uint64
	WebhookURL            string
}

// NewMsgHandler creates errors and messages handler.
func NewMsgHandler(
	connection infrastructureWhatsapp.Conn,
	wapiSession *model.WapiSession,
	messageRepo repository.Message,
	connectionsSupervisor Connections,
	sessionRepo repository.Session,
	client httpInfra.Client,
	marshal *jsonInfra.MarshallCallback,
	initTimestamp uint64,
	webhookURL string,
) *Handler {
	return &Handler{
		Connection:            connection,
		Session:               wapiSession,
		messageRepo:           messageRepo,
		InitTimestamp:         initTimestamp,
		WebhookURL:            webhookURL,
		connectionsSupervisor: connectionsSupervisor,
		storedSession:         sessionRepo,
		client:                client,
		marshal:               marshal,
	}
}

// HandleError handles connection errors.
func (h *Handler) HandleError(err error) {
	reconnect := func(interval time.Duration) {
		var pong bool
		pong, err = h.Connection.AdminTest()
		if !pong || (err != nil && err == whatsapp.ErrNotConnected) {
			h.connectionsSupervisor.RemoveConnectionForSession(h.Session.SessionID)
			_ = h.storedSession.RemoveSession(h.Session.SessionID)
			log.Printf("device isn't responding, session will be removed, reconnection canceled: %v", err)
			sentry.CaptureException(fmt.Errorf(
				"device lost connection, need to connect manually (by QR-code) `%s`, login: `%s`: %v",
				h.Session.SessionID,
				h.Connection.Info().Wid,
				err,
			))
			sentry.Flush(time.Second * SentryFlushTimeoutSeconds)
			return
		}

		log.Printf("waiting %d sec...\n", interval)
		<-time.After(interval * time.Second)
		log.Println("reconnecting...")
		if _, err = h.Connection.RestoreWithSession(h.Session.WhatsAppSession); err != nil {
			log.Printf("restore failed, session `%v`: %v", h.Session.SessionID, err)
			sentry.CaptureException(errors.Wrapf(
				err,
				"couldn't restore connection for session `%s`, login: `%s`",
				h.Session.SessionID,
				h.Connection.Info().Wid,
			))
			sentry.Flush(time.Second * SentryFlushTimeoutSeconds)
		} else {
			log.Println("ok")
		}
	}

	if e, ok := err.(*whatsapp.ErrConnectionClosed); ok {
		log.Printf("connection closed for session `%s`, code: %v, text: %v", h.Session.SessionID, e.Code, e.Text)
		reconnect(1)
		return
	}

	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		log.Printf("connection failed for session `%s`, underlying error: %v", h.Session.SessionID, e.Err)

		timeOut := time.Second * 30
		reconnect(timeOut)
		return
	}

	log.Printf("warning: %v\n", err)
}

// HandleTextMessage sends message to webhook and stores it in repository.
func (h *Handler) HandleTextMessage(msg *whatsapp.TextMessage) {
	if h.InitTimestamp == 0 {
		h.InitTimestamp = uint64(time.Now().Unix())
	}

	if !h.isMessageAllowedToHandle(msg) {
		return
	}

	log.Printf("got msg to handle from `%v`, destination `%v`", msg.Info.RemoteJid, h.Session.WhatsAppSession.Wid)

	marshal := *h.marshal
	requestBody, err := marshal(&msg)
	if err != nil {
		log.Println("error msg marshaling", err)
		return
	}

	_, err = h.client.Post(h.SessionWebhookURL(), "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("error happened getting the response", err)
		return
	}

	log.Printf("msg sent to `%s`, by session `%s`, login `%s`", h.SessionWebhookURL(), h.Session.SessionID, h.Session.WhatsAppSession.Wid)

	err = h.messageRepo.SaveMessageTime("wapi_sent_message:"+msg.Info.Id, time.Now())
	if err != nil {
		log.Printf("can't store msg id `%s` in redis: %v\n", msg.Info.Id, err)
		return
	}
}

func (h *Handler) isMessageAllowedToHandle(msg *whatsapp.TextMessage) bool {
	if h.messageAlreadySent(msg.Info.Id) {
		return false
	}
	if msg.Info.Timestamp <= h.InitTimestamp {
		return false
	}
	if msg.Info.FromMe {
		return false
	}
	return true
}

func (h *Handler) messageAlreadySent(messageID string) bool {
	_, err := h.messageRepo.MessageTime("wapi_sent_message:" + messageID)
	return err == nil
}

// Builds webhook URL accordingly session id.
func (h *Handler) SessionWebhookURL() string {
	return h.WebhookURL + h.Session.SessionID
}
