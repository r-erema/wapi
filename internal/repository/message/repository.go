package message

import "time"

// Stores messages metadata.
type Repository interface {
	SaveMessageTime(msgID string, time time.Time) error
	GetMessageTime(msgID string) (*time.Time, error)
}
