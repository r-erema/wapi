package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/r-erema/wapi/internal/model"
	"github.com/r-erema/wapi/internal/service"

	"github.com/Rhymen/go-whatsapp"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// ActiveConnectionInfoHandler provides info about connection by session ID.
type ActiveConnectionInfoHandler struct {
	connectionSupervisor service.Connections
}

// NewInfo creates ActiveConnectionInfoHandler.
func NewInfo(connectionSupervisor service.Connections) *ActiveConnectionInfoHandler {
	return &ActiveConnectionInfoHandler{connectionSupervisor: connectionSupervisor}
}

// Handle sends information of session and connection.
func (handler *ActiveConnectionInfoHandler) Handle(w http.ResponseWriter, r *http.Request) *AppError {
	params := mux.Vars(r)
	sessionID := params["sessionID"]
	result, err := handler.connectionSupervisor.AuthenticatedConnectionForSession(sessionID)
	if err != nil {
		return &AppError{
			Error:       errors.Wrap(err, "can't find active connection in info handler"),
			ResponseMsg: "can't find active connection",
			Code:        http.StatusNotFound,
		}
	}

	err = json.NewEncoder(w).Encode(&Resp{ConnectionInfo: result.Wac().Info(), SessionInfo: result.Session()})
	if err != nil {
		return &AppError{
			Error:       errors.Wrap(err, "can't encode result in info handler"),
			ResponseMsg: "can't encode result",
			Code:        http.StatusInternalServerError,
		}
	}
	log.Printf("connection info for session `%s` sent", sessionID)

	return nil
}

// Resp contains connection and session information.
type Resp struct {
	ConnectionInfo *whatsapp.Info
	SessionInfo    *model.WapiSession
}
