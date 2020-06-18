package Handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/Rhymen/go-whatsapp"
	"github.com/getsentry/sentry-go"
	"github.com/go-redis/redis"
	"github.com/r-erema/wapi/src/Service/ConnectionsSupervisor"
	"github.com/r-erema/wapi/src/Service/SessionWorks"
	"github.com/r-erema/wapi/src/Session"
)

type Handler struct {
	Connection            *whatsapp.Conn
	Session               *Session.WapiSession
	redisClient           *redis.Client
	connectionsSupervisor ConnectionsSupervisor.Interface
	sessionWorks          SessionWorks.Interface
	InitTimestamp         uint64
	WebhookUrl            string
}

func NewHandler(
	connection *whatsapp.Conn,
	session *Session.WapiSession,
	redisClient *redis.Client,
	connectionsSupervisor ConnectionsSupervisor.Interface,
	sessionWorks SessionWorks.Interface,
	initTimestamp uint64,
	webhookUrl string,
) *Handler {
	return &Handler{
		Connection:            connection,
		Session:               session,
		redisClient:           redisClient,
		InitTimestamp:         initTimestamp,
		WebhookUrl:            webhookUrl,
		connectionsSupervisor: connectionsSupervisor,
		sessionWorks:          sessionWorks,
	}
}

func (h *Handler) HandleError(err error) {
	reconnect := func(afterSeconds time.Duration) {

		pong, err := h.Connection.AdminTest()
		if !pong || (err != nil && err == whatsapp.ErrNotConnected) {
			h.connectionsSupervisor.RemoveConnectionForSession(h.Session.SessionId)
			_ = h.sessionWorks.RemoveSession(h.Session.SessionId)
			log.Printf("device isn't responding, session will be removed, reconnection cancelled: %v", err)
			sentry.CaptureException(fmt.Errorf(
				"device lost connection, need to connect manually (by QR-code) `%s`, login: `%s`: %v",
				h.Session.SessionId,
				h.Connection.Info.Wid,
				err,
			))
			sentry.Flush(time.Second * 5)
			return
		}

		log.Printf("waiting %d sec...\n", afterSeconds)
		<-time.After(afterSeconds * time.Second)
		log.Println("reconnecting...")
		_, err = h.Connection.RestoreWithSession(*h.Session.WhatsAppSession)
		if err != nil {
			log.Printf("restore failed, session `%v`: %v", h.Session.SessionId, err)
			sentry.CaptureException(fmt.Errorf(
				"couldn't restore connection for session `%s`, login: `%s`: %v",
				h.Session.SessionId,
				h.Connection.Info.Wid,
				err,
			))
			sentry.Flush(time.Second * 5)
		} else {
			log.Println("ok")
		}
	}

	if e, ok := err.(*whatsapp.ErrConnectionClosed); ok {
		log.Printf("connection closed for session `%s`, code: %v, text: %v", h.Session.SessionId, e.Code, e.Text)
		reconnect(1)
	} else if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		log.Printf("connection failed for session `%s`, underlying error: %v", h.Session.SessionId, e.Err)
		reconnect(30)
	} else {
		log.Printf("warning: %v\n", err)
	}
}

func (h *Handler) HandleTextMessage(message whatsapp.TextMessage) {

	if h.InitTimestamp == 0 {
		h.InitTimestamp = uint64(time.Now().Unix())
	}

	if !h.isMessageAllowedToHandle(message) {
		return
	}

	log.Printf("got message to handle from `%v`, destination `%v`", message.Info.RemoteJid, h.Session.WhatsAppSession.Wid)

	if message.Info.FromMe {
		return
	}

	webhookUrl := h.WebhookUrl + h.Session.SessionId

	requestBody, err := json.Marshal(&message)
	resp, err := http.Post(webhookUrl, "application/json", bytes.NewBuffer(requestBody))

	if nil != err {
		log.Println("error happened getting the response", err)
		return
	}

	err = h.redisClient.Set("wapi_sent_message:"+message.Info.Id, time.Now().String(), time.Hour*24*30).Err()
	if err != nil {
		log.Printf("can't store message id `%s` in redis: %v\n", message.Info.Id, err)
	}
	log.Printf("message sent to `%s`, by session `%s`, login `%s`", webhookUrl, h.Session.SessionId, h.Session.WhatsAppSession.Wid)

	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)

	if nil != err {
		fmt.Println("error happened reading the body", err)
		return
	}
}

func (h *Handler) isMessageAllowedToHandle(message whatsapp.TextMessage) bool {
	if h.messageAlreadySent(message.Info.Id) {
		return false
	}
	if message.Info.Timestamp <= h.InitTimestamp {
		return false
	}
	if message.Info.FromMe {
		return false
	}
	return true
}

func (h *Handler) messageAlreadySent(messageId string) bool {
	_, err := h.redisClient.Get("wapi_sent_message:" + messageId).Result()
	if err != nil {
		return false
	}
	return true
}
