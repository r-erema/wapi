package whatsapp

import (
	"time"

	"github.com/Rhymen/go-whatsapp"
	"github.com/pkg/errors"
)

// ErrMsg401 should emerge if login failed because of 401 response.
const ErrMsg401 = "admin login responded with 401"

// Conn is an object of connection with Whatsapp server.
type Conn interface {
	// Send sends messages to WhatsApp server.
	Send(msg interface{}) (string, error)
	// Info provides connection info.
	Info() *whatsapp.Info
	// AdminTest pings connection between WhatsApp server.
	AdminTest() (bool, error)
	// Disconnect destroys connection.
	Disconnect() (whatsapp.Session, error)
	// RestoreWithSession restores connection suing session object.
	RestoreWithSession(session *whatsapp.Session) (_ whatsapp.Session, err error)
	// Login authenticates by qr code.
	Login(qrChan chan<- string) (whatsapp.Session, error)
	// AddHandler registers messages handler.
	AddHandler(handler whatsapp.Handler)
}

// RhymenConn is an object of connection with Whatsapp server
// implemented using github.com/Rhymen/go-whatsapp package.
type RhymenConn struct {
	wac *whatsapp.Conn
}

// NewRhymenConn creates connection object with WhatsApp server.
func NewRhymenConn(timeout time.Duration) (*RhymenConn, error) {
	wac, err := whatsapp.NewConn(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "connection failure")
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

// Disconnect destroys connection.
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
