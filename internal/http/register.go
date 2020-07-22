package http

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/r-erema/wapi/internal/repository"
	"github.com/r-erema/wapi/internal/service"

	"github.com/pkg/errors"
)

// RegisterSessionHandler is responsible for creation of new session.
type RegisterSessionHandler struct {
	auth        service.Authorizer
	listener    service.Listener
	sessionRepo repository.Session
}

// NewRegisterSessionHandler creates RegisterSessionHandler.
func NewRegisterSessionHandler(
	authorizer service.Authorizer,
	l service.Listener,
	sessRepo repository.Session,
) *RegisterSessionHandler {
	return &RegisterSessionHandler{auth: authorizer, listener: l, sessionRepo: sessRepo}
}

// Hendle registers session and starts listening incoming messages.
func (handler *RegisterSessionHandler) Handle(w http.ResponseWriter, r *http.Request) *AppError {
	decoder := json.NewDecoder(r.Body)
	var registerSession RegisterSessionRequest
	err := decoder.Decode(&registerSession)

	if err != nil {
		return &AppError{
			Error:       errors.Wrap(err, "request decoding error in register handler"),
			ResponseMsg: "request decoding error",
			Code:        http.StatusBadRequest,
		}
	}
	if registerSession.SessionID == "" {
		return &AppError{
			Error:       errors.Wrap(err, "couldn't decode session_id param in register handler"),
			ResponseMsg: "couldn't decode session_id param",
			Code:        http.StatusBadRequest,
		}
	}

	if err = handler.startListenIncomingMessages(registerSession.SessionID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf(`%v`, err)
		return &AppError{
			Error:       errors.Wrap(err, "start listening error in register handler"),
			ResponseMsg: "start listening error",
			Code:        http.StatusInternalServerError,
		}
	}

	return nil
}

func (handler *RegisterSessionHandler) startListenIncomingMessages(sessionID string) error {
	var wg sync.WaitGroup
	wg.Add(1)
	errChan := make(chan error)
	go func(sid string) {
		if _, err := handler.listener.ListenForSession(sid, &wg); err != nil {
			errChan <- err
		}
	}(sessionID)
	wg.Wait()

	select {
	case err := <-errChan:
		return errors.Wrap(err, "error occurred while listening message for session")
	default:
		return nil
	}
}

// TryToAutoConnectAllSessions make attempt to connect sessions automatically.
func (handler *RegisterSessionHandler) TryToAutoConnectAllSessions() error {
	sessionIDs, err := handler.sessionRepo.AllSavedSessionIds()
	if err != nil {
		return errors.Wrap(err, "couldn't get all session ids")
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
