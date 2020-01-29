package MessageListener

import (
	"Service/Auth"
	"Service/ConnectionsSupervisor"
	"Service/MessageListener/Handler"
	"Service/SessionWorks"
	"github.com/go-redis/redis"
	_ "github.com/thoas/go-funk"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
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
