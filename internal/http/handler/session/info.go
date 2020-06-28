package session

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	sessionRepo "github.com/r-erema/wapi/internal/repository/session"
)

type SessInfoHandler struct {
	sessionWork sessionRepo.Repository
}

// Creates SessInfoHandler.
func NewSessInfoHandler(sessionWork sessionRepo.Repository) *SessInfoHandler {
	return &SessInfoHandler{sessionWork: sessionWork}
}

func (handler *SessInfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sessionID := params["sessionID"]
	session, err := handler.sessionWork.ReadSession(sessionID)
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
	err = json.NewEncoder(w).Encode(session)
	if err != nil {
		errPrefix := "can't encode session"
		http.Error(w, errPrefix, http.StatusInternalServerError)
		log.Printf("%s: %v", errPrefix, err)
		return
	}
}
