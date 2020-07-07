package service

import (
	"github.com/r-erema/wapi/internal/infrastructure/whatsapp"
	"github.com/r-erema/wapi/internal/model"
)

// SessionConnectionDTO containing service data for supervising.
type SessionConnectionDTO struct {
	wac      whatsapp.Conn
	session  *model.WapiSession
	pingQuit *chan string
}

// Session gets session.
func (s SessionConnectionDTO) Session() *model.WapiSession {
	return s.session
}

// Wac gets connection.
func (s SessionConnectionDTO) Wac() whatsapp.Conn {
	return s.wac
}

// NewDTO creates DTO.
func NewDTO(wac whatsapp.Conn, sess *model.WapiSession) *SessionConnectionDTO {
	quitCh := make(chan string)
	return &SessionConnectionDTO{wac: wac, session: sess, pingQuit: &quitCh}
}
