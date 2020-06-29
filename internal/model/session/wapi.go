package session

import "github.com/Rhymen/go-whatsapp"

// WapiSession is a model of wapi session.
type WapiSession struct {
	SessionID       string
	WhatsAppSession *whatsapp.Session
}
