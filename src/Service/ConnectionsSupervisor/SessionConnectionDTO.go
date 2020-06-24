package ConnectionsSupervisor

import (
	"github.com/r-erema/wapi/src/Session"

	"github.com/Rhymen/go-whatsapp"
)

type SessionConnectionDTO struct {
	wac      *whatsapp.Conn
	session  *Session.WapiSession
	pingQuit *chan string
}

func (s SessionConnectionDTO) GetSession() *Session.WapiSession {
	return s.session
}

func (s SessionConnectionDTO) GetWac() *whatsapp.Conn {
	return s.wac
}

func NewSessionConnectionDTO(wac *whatsapp.Conn, session *Session.WapiSession) *SessionConnectionDTO {
	quitCh := make(chan string)
	return &SessionConnectionDTO{wac: wac, session: session, pingQuit: &quitCh}
}
