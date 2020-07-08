package repository

import (
	"time"

	"github.com/r-erema/wapi/internal/model"
)

// Message stores messages metadata.
type Message interface {
	// SaveMessageTime stores message time in repository.
	SaveMessageTime(msgID string, time time.Time) error
	// MessageTime retrieves message time from repository.
	MessageTime(msgID string) (*time.Time, error)
}

// Session stores sessions metadata.
type Session interface {
	// ReadSession retrieves session from repository.
	ReadSession(sessionID string) (*model.WapiSession, error)
	// WriteSession retrieves session from repository.
	WriteSession(session *model.WapiSession) error
	// AllSavedSessionIds retrieves all sessions ids from repository.
	AllSavedSessionIds() ([]string, error)
	// RemoveSession removes session from repository.
	RemoveSession(sessionID string) error
}
