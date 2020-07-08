package http

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/r-erema/wapi/internal/repository"

	"github.com/gorilla/mux"
)

// SessInfoHandler provides info about session.
type SessInfoHandler struct {
	sessionRepo repository.Session
}

// NewSessInfoHandler creates SessInfoHandler.
func NewSessInfoHandler(sessionWork repository.Session) *SessInfoHandler {
	return &SessInfoHandler{sessionRepo: sessionWork}
}

// ServeHTTP sends session info.
func (handler *SessInfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sessionID := params["sessionID"]
	session, err := handler.sessionRepo.ReadSession(sessionID)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			http.Error(w, "session not found", http.StatusNotFound)
		} else {
			handleError(w, "can't read session", err, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(session)
	if err != nil {
		handleError(w, "can't encode session", err, http.StatusInternalServerError)
		return
	}
}
