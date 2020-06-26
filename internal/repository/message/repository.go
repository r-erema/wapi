package message

import "time"

type Repository interface {
	SaveMessageTime(msgID string, time time.Time) error
	GetMessageTime(msgID string) (*time.Time, error)
}
