package session

import (
	"github.com/r-erema/wapi/internal/model/session"
)

// Stores sessions metadata.
type Repository interface {
	// ReadSession retrieves session from repository.
	ReadSession(sessionID string) (*session.WapiSession, error)
	// WriteSession retrieves session from repository.
	WriteSession(session *session.WapiSession) error
	// AllSavedSessionIds retrieves all sessions ids from repository.
	AllSavedSessionIds() ([]string, error)
	// RemoveSession removes session from repository.
	RemoveSession(sessionID string) error
}
