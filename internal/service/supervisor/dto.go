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

// Gets session.
func (s SessionConnectionDTO) Session() *session2.WapiSession {
	return s.session
}

// Gets connection.
func (s SessionConnectionDTO) Wac() *whatsapp.Conn {
	return s.wac
}

// Creates DTO.
func NewDTO(wac *whatsapp.Conn, session *session2.WapiSession) *SessionConnectionDTO {
	quitCh := make(chan string)
	return &SessionConnectionDTO{wac: wac, session: session, pingQuit: &quitCh}
}
