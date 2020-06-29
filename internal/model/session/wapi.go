package session

import "github.com/Rhymen/go-whatsapp"

// Model of wapi session.
type WapiSession struct {
	SessionID       string
	WhatsAppSession *whatsapp.Session
}
