package Auth

import (
	"github.com/Rhymen/go-whatsapp"
	"github.com/r-erema/wapi/src/Session"
)

type Interface interface {
	Login(sessionId string) (*whatsapp.Conn, *Session.WapiSession, error)
}
