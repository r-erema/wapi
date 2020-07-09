package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/r-erema/wapi/internal/model"
	"github.com/r-erema/wapi/internal/service"

	"github.com/Rhymen/go-whatsapp"
	"github.com/gorilla/mux"
)

// ActiveConnectionInfoHandler provides info about connection by session ID.
type ActiveConnectionInfoHandler struct {
	connectionSupervisor service.Connections
}

// NewInfo creates ActiveConnectionInfoHandler.
func NewInfo(connectionSupervisor service.Connections) *ActiveConnectionInfoHandler {
	return &ActiveConnectionInfoHandler{connectionSupervisor: connectionSupervisor}
}

// ServeHTTP sends information of session and connection.
func (handler *ActiveConnectionInfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sessionID := params["sessionID"]
	result, err := handler.connectionSupervisor.AuthenticatedConnectionForSession(sessionID)
	if err != nil {
		handleError(w, "can't find active connection", err, http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(&Resp{ConnectionInfo: result.Wac().Info(), SessionInfo: result.Session()})
	if err != nil {
		handleError(w, "can't encode result", err, http.StatusInternalServerError)
		return
	}
	log.Printf("connection info for session `%s` sent", sessionID)
}

// Resp contains connection and session information.
type Resp struct {
	ConnectionInfo *whatsapp.Info
	SessionInfo    *model.WapiSession
}
