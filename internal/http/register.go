package http

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/r-erema/wapi/internal/repository"
	"github.com/r-erema/wapi/internal/service"
)

// RegisterSessionHandler responsible for creation of new session.
type RegisterSessionHandler struct {
	auth         service.Authorizer
	listener     service.Listener
	sessionWorks repository.SessionRepository
}

// NewRegisterSessionHandler creates RegisterSessionHandler.
func NewRegisterSessionHandler(
	authorizer service.Authorizer,
	l service.Listener,
	sessRepo repository.SessionRepository,
) *RegisterSessionHandler {
	return &RegisterSessionHandler{auth: authorizer, listener: l, sessionWorks: sessRepo}
}

func (handler *RegisterSessionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var registerSession RegisterSessionRequest
	err := decoder.Decode(&registerSession)

	if err != nil {
		errPrefix := "request decoding error"
		http.Error(w, errPrefix, http.StatusBadRequest)
		log.Printf(`%v: %v`, errPrefix, err)
		return
	}
	if registerSession.SessionID == "" {
		errPrefix := "couldn't decode session_id param"
		http.Error(w, errPrefix, http.StatusBadRequest)
		log.Printf(`%v: %v`, errPrefix, err)
		return
	}

	if err = handler.startListenIncomingMessages(registerSession.SessionID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf(`%v`, err)
	}
}

func (handler *RegisterSessionHandler) startListenIncomingMessages(sessionID string) error {
	var wg sync.WaitGroup
	wg.Add(1)
	errChan := make(chan error)
	go func(sid string) {
		_, err := handler.listener.ListenForSession(sid, &wg)
		if err != nil {
			errChan <- err
		}
	}(sessionID)
	wg.Wait()

	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

// TryToAutoConnectAllSessions make attempt to connect sessions automatically.
func (handler *RegisterSessionHandler) TryToAutoConnectAllSessions() error {
	sessionIDs, err := handler.sessionWorks.AllSavedSessionIds()
	if err != nil {
		return err
	}
	for _, sessionID := range sessionIDs {
		if err := handler.startListenIncomingMessages(sessionID); err != nil {
			log.Printf("unable to auto connect session `%s`: %v", sessionID, err)
		}
	}
	return nil
}

// RegisterSessionRequest object for registering session.
type RegisterSessionRequest struct {
	SessionID string `json:"session_id"`
}
