package repository

import (
	"time"

	"github.com/r-erema/wapi/internal/model/session"
)

// Stores messages metadata.
type MessageRepository interface {
	// SaveMessageTime stores message time in repository.
	SaveMessageTime(msgID string, time time.Time) error
	// MessageTime retrieves message time from repository.
	MessageTime(msgID string) (*time.Time, error)
}

// Stores sessions metadata.
type SessionRepository interface {
	// ReadSession retrieves session from repository.
	ReadSession(sessionID string) (*session.WapiSession, error)
	// WriteSession retrieves session from repository.
	WriteSession(session *session.WapiSession) error
	// AllSavedSessionIds retrieves all sessions ids from repository.
	AllSavedSessionIds() ([]string, error)
	// RemoveSession removes session from repository.
	RemoveSession(sessionID string) error
}
