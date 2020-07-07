package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Rhymen/go-whatsapp"
	"github.com/gorilla/mux"
	"github.com/r-erema/wapi/internal/model"
	"github.com/r-erema/wapi/internal/service"
)

// ActiveConnectionInfoHandler provides info about connection by session ID.
type ActiveConnectionInfoHandler struct {
	connectionSupervisor service.Connections
}

// NewInfo creates ActiveConnectionInfoHandler.
func NewInfo(connectionSupervisor service.Connections) *ActiveConnectionInfoHandler {
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

	err = json.NewEncoder(w).Encode(&Resp{ConnectionInfo: result.Wac().Info(), SessionInfo: result.Session()})
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
	SessionInfo    *model.WapiSession
}
