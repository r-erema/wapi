package http

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/r-erema/wapi/internal/repository"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// SessInfoHandler provides info about session.
type SessInfoHandler struct {
	sessionRepo repository.Session
}

// NewSessInfoHandler creates SessInfoHandler.
func NewSessInfoHandler(sessionWork repository.Session) *SessInfoHandler {
	return &SessInfoHandler{sessionRepo: sessionWork}
}

// Handle sends session info.
func (handler *SessInfoHandler) Handle(w http.ResponseWriter, r *http.Request) *AppError {
	params := mux.Vars(r)
	sessionID := params["sessionID"]
	session, err := handler.sessionRepo.ReadSession(sessionID)
	if err != nil {
		if err, ok := err.(*os.PathError); ok {
			return &AppError{
				Error:       errors.Wrap(err, "session not found in session handler"),
				ResponseMsg: "session not found",
				Code:        http.StatusNotFound,
			}
		}

		return &AppError{
			Error:       errors.Wrap(err, "session reading error in session handler"),
			ResponseMsg: "session reading error",
			Code:        http.StatusInternalServerError,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(session)
	if err != nil {
		return &AppError{
			Error:       errors.Wrap(err, "session encoding error in session handler"),
			ResponseMsg: "can't encode session",
			Code:        http.StatusInternalServerError,
		}
	}

	return nil
}
