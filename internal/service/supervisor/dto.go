package supervisor

import (
	"github.com/r-erema/wapi/internal/model/session"

	"github.com/Rhymen/go-whatsapp"
)

// SessionConnectionDTO containing service data for supervising.
type SessionConnectionDTO struct {
	wac      *whatsapp.Conn
	session  *session.WapiSession
	pingQuit *chan string
}

// Session gets session.
func (s SessionConnectionDTO) Session() *session.WapiSession {
	return s.session
}

// Wac gets connection.
func (s SessionConnectionDTO) Wac() *whatsapp.Conn {
	return s.wac
}

// NewDTO creates DTO.
func NewDTO(wac *whatsapp.Conn, sess *session.WapiSession) *SessionConnectionDTO {
	quitCh := make(chan string)
	return &SessionConnectionDTO{wac: wac, session: sess, pingQuit: &quitCh}
}
