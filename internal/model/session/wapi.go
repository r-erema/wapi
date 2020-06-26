package session

import "github.com/Rhymen/go-whatsapp"

type WapiSession struct {
	SessionID       string
	WhatsAppSession *whatsapp.Session
}
