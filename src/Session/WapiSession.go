package Session

import "github.com/Rhymen/go-whatsapp"

type WapiSession struct {
	SessionId       string
	WhatsAppSession *whatsapp.Session
}
