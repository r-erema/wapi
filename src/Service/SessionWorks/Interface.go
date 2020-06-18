package SessionWorks

import (
	"github.com/r-erema/wapi/src/Session"
)

type Interface interface {
	ReadSession(sessionId string) (*Session.WapiSession, error)
	WriteSession(session *Session.WapiSession) error
	GetAllSavedSessionIds() ([]string, error)
	RemoveSession(sessionId string) error
}
