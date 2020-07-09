package service

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	httpInfra "github.com/r-erema/wapi/internal/infrastructure/http"
	"github.com/r-erema/wapi/internal/repository"
)

// Listener listens for incoming messages from WhatsApp server.
type Listener interface {
	// Receives messages from WhatsApp server and propagates them to handlers.
	ListenForSession(sessionID string, wg *sync.WaitGroup) (gracefulDone bool, err error)
}

// WebHook listens for incoming messages and propagate them to webhook handler.
type WebHook struct {
	sessionWorks          repository.Session
	connectionsSupervisor Connections
	auth                  Authorizer
	webhookURL            string
	msgRepo               repository.Message
	client                httpInfra.Client
}

// NewWebHook creates listener for sending messages to webhook.
func NewWebHook(
	sessionWorks repository.Session,
	connectionsSupervisor Connections,
	authorizer Authorizer,
	webhookURL string,
	msgRepo repository.Message,
	client httpInfra.Client,
) *WebHook {
	return &WebHook{
		sessionWorks:          sessionWorks,
		connectionsSupervisor: connectionsSupervisor,
		auth:                  authorizer,
		webhookURL:            webhookURL,
		msgRepo:               msgRepo,
		client:                client,
	}
}

// Receives messages from WhatsApp server and propagates them to handlers.
func (l *WebHook) ListenForSession(sessionID string, wg *sync.WaitGroup) (gracefulDone bool, err error) {
	if _, err = l.connectionsSupervisor.AuthenticatedConnectionForSession(sessionID); err == nil {
		log.Printf("Session `%s` is already listenning", sessionID)
		return false, err
	}

	wac, session, err := l.auth.Login(sessionID)
	if err != nil || wac == nil || session == nil {
		log.Printf("login failed in message ListenerWebHook: %v\n", err)
		wg.Done()
		return false, err
	}

	log.Printf("start listening messages for sessionRepo `%s`, bound login: `%s`", session.SessionID, session.WhatsAppSession.Wid)

	wac.AddHandler(NewHandler(
		wac,
		session,
		l.msgRepo,
		l.connectionsSupervisor,
		l.sessionWorks,
		l.client,
		uint64(time.Now().Unix()),
		l.webhookURL,
	))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	wg.Done()
	<-c

	waSession, err := wac.Disconnect()
	session.WhatsAppSession = &waSession
	if err != nil {
		log.Printf("error disconnecting: %v\n", err)
		return false, err
	}
	if err := l.sessionWorks.WriteSession(session); err != nil {
		log.Printf("error saving sessionRepo: %v", err)
		return false, err
	}
	return true, nil
}
