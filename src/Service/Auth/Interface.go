package Auth

import (
	"github.com/r-erema/wapi/src/Session"

	"github.com/Rhymen/go-whatsapp"
)

type Interface interface {
	Login(sessionId string) (*whatsapp.Conn, *Session.WapiSession, error)
}
