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
)

type Listener interface {
	ListenForSession(sessionID string, wg *sync.WaitGroup) (gracefulDone bool, err error)
}

type WebHook struct {
	sessionWorks          sessionRepo.Repository
	connectionsSupervisor supervisor.ConnectionSupervisor
	auth                  auth.Authorizer
	webhookURL            string
	msgRepo               message.Repository
}

func NewListener(
	sessionWorks sessionRepo.Repository,
	connectionsSupervisor supervisor.ConnectionSupervisor,
	authorizer auth.Authorizer,
	webhookURL string,
	msgRepo message.Repository,
) *WebHook {
	return &WebHook{
		sessionWorks:          sessionWorks,
		connectionsSupervisor: connectionsSupervisor,
		auth:                  authorizer,
		webhookURL:            webhookURL,
		msgRepo:               msgRepo,
	}
}

func (l *WebHook) ListenForSession(sessionID string, wg *sync.WaitGroup) (gracefulDone bool, err error) {
	_, err = l.connectionsSupervisor.GetAuthenticatedConnectionForSession(sessionID)
	if err == nil {
		log.Printf("Session `%s` is already listenning", sessionID)
		return false, err
	}

	wac, session2, err := l.auth.Login(sessionID)
	if err != nil || wac == nil || session2 == nil {
		log.Printf("login failed in message ListenerWebHook: %v\n", err)
		wg.Done()
		return false, err
	}

	log.Printf("start listening messages for sessionRepo `%s`, bound login: `%s`", session2.SessionID, session2.WhatsAppSession.Wid)

	wac.AddHandler(NewHandler(
		wac,
		session2,
		l.msgRepo,
		l.connectionsSupervisor,
		l.sessionWorks,
		uint64(time.Now().Unix()),
		l.webhookURL,
	))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	wg.Done()
	<-c

	waSession, err := wac.Disconnect()
	session2.WhatsAppSession = &waSession
	if err != nil {
		log.Printf("error disconnecting: %v\n", err)
		return false, err
	}
	if err := l.sessionWorks.WriteSession(session2); err != nil {
		log.Printf("error saving sessionRepo: %v", err)
		return false, err
	}
	return true, nil
}
