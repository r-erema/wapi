package session

import (
	"github.com/r-erema/wapi/internal/model/session"
)

type Repository interface {
	ReadSession(sessionID string) (*session.WapiSession, error)
	WriteSession(session *session.WapiSession) error
	AllSavedSessionIds() ([]string, error)
	RemoveSession(sessionID string) error
}
