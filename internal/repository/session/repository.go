package session

import (
	"github.com/r-erema/wapi/internal/model/session"
)

type Repository interface {
	ReadSession(sessionId string) (*session.WapiSession, error)
	WriteSession(session *session.WapiSession) error
	GetAllSavedSessionIds() ([]string, error)
	RemoveSession(sessionId string) error
}
