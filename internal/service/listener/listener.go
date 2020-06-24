package listener

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/r-erema/wapi/internal/repository/message"
	sessionRepo "github.com/r-erema/wapi/internal/repository/session"
	"github.com/r-erema/wapi/internal/service/auth"
	"github.com/r-erema/wapi/internal/service/supervisor"

	_ "github.com/thoas/go-funk"
)

type Listener interface {
	ListenForSession(sessionId string, wg *sync.WaitGroup)
}

type listener struct {
	sessionWorks          sessionRepo.Repository
	connectionsSupervisor supervisor.ConnectionSupervisor
	auth                  auth.Authorizer
	webhookUrl            string
	msgRepo               message.Repository
}

func NewListener(
	sessionWorks sessionRepo.Repository,
	connectionsSupervisor supervisor.ConnectionSupervisor,
	auth auth.Authorizer,
	webhookUrl string,
	msgRepo message.Repository,
) *listener {
	return &listener{
		sessionWorks:          sessionWorks,
		connectionsSupervisor: connectionsSupervisor,
		auth:                  auth,
		webhookUrl:            webhookUrl,
		msgRepo:               msgRepo,
	}
}

func (l *listener) ListenForSession(sessionId string, wg *sync.WaitGroup) {

	_, err := l.connectionsSupervisor.GetAuthenticatedConnectionForSession(sessionId)
	if err == nil {
		log.Printf("Session `%s` is already listenning", sessionId)
		return
	}

	wac, session2, err := l.auth.Login(sessionId)
	if err != nil || wac == nil || session2 == nil {
		log.Printf("login failed in message listener: %v\n", err)
		wg.Done()
		return
	}

	log.Printf("start listening messages for sessionRepo `%s`, bound login: `%s`", session2.SessionId, session2.WhatsAppSession.Wid)

	wac.AddHandler(NewHandler(
		wac,
		session2,
		l.msgRepo,
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
	session2.WhatsAppSession = &waSession
	if err != nil {
		log.Printf("error disconnecting: %v\n", err)
		return
	}
	if err := l.sessionWorks.WriteSession(session2); err != nil {
		log.Printf("error saving sessionRepo: %v", err)
		return
	}
}
