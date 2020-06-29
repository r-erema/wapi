package session

import (
	"github.com/r-erema/wapi/internal/model/session"
)

// Stores sessions metadata.
type Repository interface {
	ReadSession(sessionID string) (*session.WapiSession, error)
	WriteSession(session *session.WapiSession) error
	AllSavedSessionIds() ([]string, error)
	RemoveSession(sessionID string) error
}
