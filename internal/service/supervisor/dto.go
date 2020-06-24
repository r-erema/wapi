package supervisor

import (
	session2 "github.com/r-erema/wapi/internal/model/session"

	"github.com/Rhymen/go-whatsapp"
)

type SessionConnectionDTO struct {
	wac      *whatsapp.Conn
	session  *session2.WapiSession
	pingQuit *chan string
}

func (s SessionConnectionDTO) GetSession() *session2.WapiSession {
	return s.session
}

func (s SessionConnectionDTO) GetWac() *whatsapp.Conn {
	return s.wac
}

func NewSessionConnectionDTO(wac *whatsapp.Conn, session *session2.WapiSession) *SessionConnectionDTO {
	quitCh := make(chan string)
	return &SessionConnectionDTO{wac: wac, session: session, pingQuit: &quitCh}
}
