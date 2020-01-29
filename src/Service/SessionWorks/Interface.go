package SessionWorks

import (
	"Session"
)

type Interface interface {
	ReadSession(sessionId string) (*Session.WapiSession, error)
	WriteSession(session *Session.WapiSession) error
	GetAllSavedSessionIds() ([]string, error)
	RemoveSession(sessionId string) error
}
