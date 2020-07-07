package http

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	sessionRepo "github.com/r-erema/wapi/internal/repository/session"
)

// SessInfoHandler provides info about session.
type SessInfoHandler struct {
	sessionRepo sessionRepo.Repository
}

// NewSessInfoHandler creates SessInfoHandler.
func NewSessInfoHandler(sessionWork sessionRepo.Repository) *SessInfoHandler {
	return &SessInfoHandler{sessionRepo: sessionWork}
}

func (handler *SessInfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sessionID := params["sessionID"]
	session, err := handler.sessionRepo.ReadSession(sessionID)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			http.Error(w, "session not found", http.StatusNotFound)
		} else {
			errPrefix := "can't read session"
			http.Error(w, errPrefix, http.StatusInternalServerError)
			log.Printf("%s: %v", errPrefix, err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(session)
	if err != nil {
		errPrefix := "can't encode session"
		http.Error(w, errPrefix, http.StatusInternalServerError)
		log.Printf("%s: %v", errPrefix, err)
		return
	}
}
