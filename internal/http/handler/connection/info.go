package connection

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Rhymen/go-whatsapp"
	"github.com/gorilla/mux"
	"github.com/r-erema/wapi/internal/model/session"
	"github.com/r-erema/wapi/internal/service/supervisor"
)

type GetActiveConnectionInfoHandler struct {
	connectionSupervisor supervisor.ConnectionSupervisor
}

func NewGetActiveConnectionInfoHandler(connectionSupervisor supervisor.ConnectionSupervisor) *GetActiveConnectionInfoHandler {
	return &GetActiveConnectionInfoHandler{connectionSupervisor: connectionSupervisor}
}

func (handler *GetActiveConnectionInfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sessionId := params["sessionId"]
	result, err := handler.connectionSupervisor.GetAuthenticatedConnectionForSession(sessionId)
	if err != nil {
		errPrefix := "can't find active connection"
		http.Error(w, errPrefix, http.StatusNotFound)
		log.Printf("%s: %v", errPrefix, err)
		return
	}

	err = json.NewEncoder(w).Encode(&Resp{ConnectionInfo: result.GetWac().Info, SessionInfo: result.GetSession()})
	if err != nil {
		errPrefix := "can't encode result"
		http.Error(w, errPrefix, http.StatusInternalServerError)
		log.Printf("%s: %v", errPrefix, err)
		return
	}
	log.Printf("connection info for session `%s` sent", sessionId)
}

type Resp struct {
	ConnectionInfo *whatsapp.Info
	SessionInfo    *session.WapiSession
}
