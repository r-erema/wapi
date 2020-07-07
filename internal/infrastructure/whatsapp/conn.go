package whatsapp

import (
	"time"

	"github.com/Rhymen/go-whatsapp"
)

// Connection object with WhatsApp server.
type Conn interface {
	Send(msg interface{}) (string, error)
	Info() *whatsapp.Info
	AdminTest() (bool, error)
	Disconnect() (whatsapp.Session, error)
	RestoreWithSession(session *whatsapp.Session) (_ whatsapp.Session, err error)
	Login(qrChan chan<- string) (whatsapp.Session, error)
	AddHandler(handler whatsapp.Handler)
}

// Connection object with WhatsApp server.
type RhymenConn struct {
	wac *whatsapp.Conn
}

// Creates connection object with WhatsApp server.
func NewRhymenConn(timeout time.Duration) (*RhymenConn, error) {
	wac, err := whatsapp.NewConn(timeout)
	if err != nil {
		return nil, err
	}
	return &RhymenConn{wac: wac}, nil
}

// Send sends messages to WhatsApp server.
func (r *RhymenConn) Send(msg interface{}) (string, error) {
	return r.wac.Send(msg)
}

// Info provides connection info.
func (r *RhymenConn) Info() *whatsapp.Info {
	return r.wac.Info
}

// AdminTest pings connection between WhatsApp server.
func (r *RhymenConn) AdminTest() (bool, error) {
	return r.wac.AdminTest()
}

// Disconnect destroy connrection.
func (r *RhymenConn) Disconnect() (whatsapp.Session, error) {
	return r.wac.Disconnect()
}

// RestoreWithSession restores connection suing session object.
func (r *RhymenConn) RestoreWithSession(session *whatsapp.Session) (_ whatsapp.Session, err error) {
	return r.wac.RestoreWithSession(*session)
}

// Login authenticates by qr code.
func (r *RhymenConn) Login(qrChan chan<- string) (whatsapp.Session, error) {
	return r.wac.Login(qrChan)
}

// AddHandler registers messages handler.
func (r *RhymenConn) AddHandler(handler whatsapp.Handler) {
	r.wac.AddHandler(handler)
}
