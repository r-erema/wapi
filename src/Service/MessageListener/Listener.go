package MessageListener

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/r-erema/wapi/src/Service/Auth"
	"github.com/r-erema/wapi/src/Service/ConnectionsSupervisor"
	"github.com/r-erema/wapi/src/Service/MessageListener/Handler"
	"github.com/r-erema/wapi/src/Service/SessionWorks"

	"github.com/go-redis/redis"
	_ "github.com/thoas/go-funk"
)

type listener struct {
	sessionWorks          SessionWorks.Interface
	connectionsSupervisor ConnectionsSupervisor.Interface
	auth                  Auth.Interface
	webhookUrl            string
	redisClient           *redis.Client
}

func NewListener(
	sessionWorks SessionWorks.Interface,
	connectionsSupervisor ConnectionsSupervisor.Interface,
	auth Auth.Interface,
	webhookUrl string,
	redisClient *redis.Client,
) *listener {
	return &listener{
		sessionWorks:          sessionWorks,
		connectionsSupervisor: connectionsSupervisor,
		auth:                  auth,
		webhookUrl:            webhookUrl,
		redisClient:           redisClient,
	}
}

func (l *listener) ListenForSession(sessionId string, wg *sync.WaitGroup) {

	_, err := l.connectionsSupervisor.GetAuthenticatedConnectionForSession(sessionId)
	if err == nil {
		log.Printf("Session `%s` is already listenning", sessionId)
		return
	}

	wac, session, err := l.auth.Login(sessionId)
	if err != nil || wac == nil || session == nil {
		log.Printf("login failed in message listener: %v\n", err)
		wg.Done()
		return
	}

	log.Printf("start listening messages for session `%s`, bound login: `%s`", session.SessionId, session.WhatsAppSession.Wid)

	wac.AddHandler(Handler.NewHandler(
		wac,
		session,
		l.redisClient,
		l.connectionsSupervisor,
		l.sessionWorks,
		uint64(time.Now().Unix()),
		l.webhookUrl,
	))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	wg.Done()
	<-c

	waSession, err := wac.Disconnect()
	session.WhatsAppSession = &waSession
	if err != nil {
		log.Printf("error disconnecting: %v\n", err)
		return
	}
	if err := l.sessionWorks.WriteSession(session); err != nil {
		log.Printf("error saving session: %v", err)
		return
	}
}
