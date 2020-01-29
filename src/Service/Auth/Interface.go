package Auth

import (
	"Session"
	"github.com/Rhymen/go-whatsapp"
)

type Interface interface {
	Login(sessionId string) (*whatsapp.Conn, *Session.WapiSession, error)
}
