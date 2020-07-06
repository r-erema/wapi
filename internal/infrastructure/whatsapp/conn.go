package whatsapp

import (
	"time"

	"github.com/Rhymen/go-whatsapp"
)

type Conn interface {
	Send(msg interface{}) (string, error)
	Info() *whatsapp.Info
	AdminTest() (bool, error)
	Disconnect() (whatsapp.Session, error)
	RestoreWithSession(session whatsapp.Session) (_ whatsapp.Session, err error)
	Login(qrChan chan<- string) (whatsapp.Session, error)
	AddHandler(handler whatsapp.Handler)
}

type RhymenConn struct {
	wac *whatsapp.Conn
}

func NewRhymenConn(timeout time.Duration) (*RhymenConn, error) {
	wac, err := whatsapp.NewConn(timeout)
	if err != nil {
		return nil, err
	}
	return &RhymenConn{wac: wac}, nil
}

func (r *RhymenConn) Send(msg interface{}) (string, error) {
	return r.wac.Send(msg)
}

func (r *RhymenConn) Info() *whatsapp.Info {
	return r.wac.Info
}

func (r *RhymenConn) AdminTest() (bool, error) {
	return r.wac.AdminTest()
}

func (r *RhymenConn) Disconnect() (whatsapp.Session, error) {
	return r.wac.Disconnect()
}

func (r *RhymenConn) RestoreWithSession(session whatsapp.Session) (_ whatsapp.Session, err error) {
	return r.wac.RestoreWithSession(session)
}

func (r *RhymenConn) Login(qrChan chan<- string) (whatsapp.Session, error) {
	return r.wac.Login(qrChan)
}

func (r *RhymenConn) AddHandler(handler whatsapp.Handler) {
	r.wac.AddHandler(handler)
}
