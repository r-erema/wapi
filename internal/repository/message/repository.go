package message

import "time"

// Stores messages metadata.
type Repository interface {
	// SaveMessageTime stores message time in repository.
	SaveMessageTime(msgID string, time time.Time) error
	// MessageTime retrieves message time from repository.
	MessageTime(msgID string) (*time.Time, error)
}
