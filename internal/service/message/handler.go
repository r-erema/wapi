package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/r-erema/wapi/internal/model/session"
	"github.com/r-erema/wapi/internal/repository/message"
	storedSession "github.com/r-erema/wapi/internal/repository/session"
	"github.com/r-erema/wapi/internal/service/supervisor"

	"github.com/Rhymen/go-whatsapp"
	"github.com/getsentry/sentry-go"
)

// Sentry flush timout
const SentryFlushTimeoutSeconds = 5

// Handler responsible handle incoming messages and errors.
type Handler struct {
	Connection            *whatsapp.Conn
	Session               *session.WapiSession
	messageRepo           message.Repository
	connectionsSupervisor supervisor.Connections
	storedSession         storedSession.Repository
	InitTimestamp         uint64
	WebhookURL            string
}

// NewHandler creates errors and messages handler.
func NewHandler(
	connection *whatsapp.Conn,
	wapiSession *session.WapiSession,
	messageRepo message.Repository,
	connectionsSupervisor supervisor.Connections,
	sessionWorks storedSession.Repository,
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
		storedSession:         sessionWorks,
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
				h.Connection.Info.Wid,
				err,
			))
			sentry.Flush(time.Second * SentryFlushTimeoutSeconds)
			return
		}

		log.Printf("waiting %d sec...\n", interval)
		<-time.After(interval * time.Second)
		log.Println("reconnecting...")
		_, err = h.Connection.RestoreWithSession(*h.Session.WhatsAppSession)
		if err != nil {
			log.Printf("restore failed, session `%v`: %v", h.Session.SessionID, err)
			sentry.CaptureException(fmt.Errorf(
				"couldn't restore connection for session `%s`, login: `%s`: %v",
				h.Session.SessionID,
				h.Connection.Info.Wid,
				err,
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

	if msg.Info.FromMe {
		return
	}

	requestBody, err := json.Marshal(&msg)
	if err != nil {
		log.Println("error msg marshaling", err)
	}
	resp, err := http.Post(h.SessionWebhookURL(), "application/json", bytes.NewBuffer(requestBody))

	if err != nil {
		log.Println("error happened getting the response", err)
		return
	}

	err = h.messageRepo.SaveMessageTime("wapi_sent_message:"+msg.Info.Id, time.Now())
	if err != nil {
		log.Printf("can't store msg id `%s` in redis: %v\n", msg.Info.Id, err)
	}
	log.Printf("msg sent to `%s`, by session `%s`, login `%s`", h.SessionWebhookURL(), h.Session.SessionID, h.Session.WhatsAppSession.Wid)

	defer func() {
		_ = resp.Body.Close()
	}()
	_, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("error happened reading the body", err)
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
