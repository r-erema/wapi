package supervisor

import (
	"github.com/r-erema/wapi/internal/model/session"

	"github.com/Rhymen/go-whatsapp"
)

type SessionConnectionDTO struct {
	wac      *whatsapp.Conn
	session  *session.WapiSession
	pingQuit *chan string
}

// Gets session.
func (s SessionConnectionDTO) Session() *session.WapiSession {
	return s.session
}

// Gets connection.
func (s SessionConnectionDTO) Wac() *whatsapp.Conn {
	return s.wac
}

// Creates DTO.
func NewDTO(wac *whatsapp.Conn, sess *session.WapiSession) *SessionConnectionDTO {
	quitCh := make(chan string)
	return &SessionConnectionDTO{wac: wac, session: sess, pingQuit: &quitCh}
}
