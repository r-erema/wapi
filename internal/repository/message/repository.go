package message

import "time"

type Repository interface {
	SaveMessageTime(msgId string, time time.Time) error
	GetMessageTime(msgId string) (time.Time, error)
}
