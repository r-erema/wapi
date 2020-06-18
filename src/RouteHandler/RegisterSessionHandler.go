package RouteHandler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/r-erema/wapi/src/Service/Auth"
	"github.com/r-erema/wapi/src/Service/MessageListener"
	"github.com/r-erema/wapi/src/Service/SessionWorks"
)

type RegisterSessionHandler struct {
	auth         Auth.Interface
	listener     MessageListener.Interface
	sessionWorks SessionWorks.Interface
}

func NewRegisterSessionHandler(
	auth Auth.Interface,
	listener MessageListener.Interface,
	sessionWorks SessionWorks.Interface,
) *RegisterSessionHandler {
	return &RegisterSessionHandler{auth: auth, listener: listener, sessionWorks: sessionWorks}
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
	if registerSession.SessionId == "" {
		errPrefix := "couldn't decode session_id param"
		http.Error(w, errPrefix, http.StatusBadRequest)
		log.Printf(`%v: %v`, errPrefix, err)
		return
	}

	if err = handler.startListenIncomingMessages(registerSession.SessionId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf(`%v`, err)
	}

}

func (handler *RegisterSessionHandler) startListenIncomingMessages(sessionId string) error {
	var wg sync.WaitGroup
	wg.Add(1)
	go func(sid string) {
		handler.listener.ListenForSession(sid, &wg)
	}(sessionId)
	wg.Wait()

	return nil
}

func (handler *RegisterSessionHandler) TryToAutoConnectAllSessions() error {
	sessionIds, err := handler.sessionWorks.GetAllSavedSessionIds()
	if err != nil {
		return err
	}
	for _, sessionId := range sessionIds {
		if err := handler.startListenIncomingMessages(sessionId); err != nil {
			log.Printf("unable to auto connect session `%s`: %v", sessionId, err)
		}
	}
	return nil
}

type RegisterSessionRequest struct {
	SessionId string `json:"session_id"`
}
