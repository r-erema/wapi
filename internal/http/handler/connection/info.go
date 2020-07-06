package connection

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/r-erema/wapi/internal/model/session"
	"github.com/r-erema/wapi/internal/service/supervisor"

	"github.com/Rhymen/go-whatsapp"
	"github.com/gorilla/mux"
)

// ActiveConnectionInfoHandler provides info about connection by session ID.
type ActiveConnectionInfoHandler struct {
	connectionSupervisor supervisor.Connections
}

// New creates ActiveConnectionInfoHandler.
func New(connectionSupervisor supervisor.Connections) *ActiveConnectionInfoHandler {
	return &ActiveConnectionInfoHandler{connectionSupervisor: connectionSupervisor}
}

func (handler *ActiveConnectionInfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sessionID := params["sessionID"]
	result, err := handler.connectionSupervisor.AuthenticatedConnectionForSession(sessionID)
	if err != nil {
		errPrefix := "can't find active connection"
		http.Error(w, errPrefix, http.StatusNotFound)
		log.Printf("%s: %v", errPrefix, err)
		return
	}

	err = json.NewEncoder(w).Encode(&Resp{ConnectionInfo: result.Wac().Info, SessionInfo: result.Session()})
	if err != nil {
		errPrefix := "can't encode result"
		http.Error(w, errPrefix, http.StatusInternalServerError)
		log.Printf("%s: %v", errPrefix, err)
		return
	}
	log.Printf("connection info for session `%s` sent", sessionID)
}

type Resp struct {
	ConnectionInfo *whatsapp.Info
	SessionInfo    *session.WapiSession
}
