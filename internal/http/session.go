package http

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/r-erema/wapi/internal/repository"
)

// SessInfoHandler provides info about session.
type SessInfoHandler struct {
	sessionRepo repository.SessionRepository
}

// NewSessInfoHandler creates SessInfoHandler.
func NewSessInfoHandler(sessionWork repository.SessionRepository) *SessInfoHandler {
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
