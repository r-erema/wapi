package RouteHandler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/r-erema/wapi/src/Service/SessionWorks"
)

type GetSessionInfoHandler struct {
	sessionWork SessionWorks.Interface
}

func NewGetSessionInfoHandler(sessionWork SessionWorks.Interface) *GetSessionInfoHandler {
	return &GetSessionInfoHandler{sessionWork: sessionWork}
}

func (handler *GetSessionInfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sessionId := params["sessionId"]
	session, err := handler.sessionWork.ReadSession(sessionId)
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
