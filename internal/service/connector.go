package service

import (
	"time"

	"github.com/r-erema/wapi/internal/infrastructure/whatsapp"
)

// Connector is responsible for connection to WhatsApp server
type Connector interface {
	Connect(timeout time.Duration) (whatsapp.Conn, error)
}

// RhymenConnector is implementation Connector interface. It is responsible for connection to WhatsApp server.
type RhymenConnector struct{}

// Connect tries to connect to the WhatsApp server and returns an object with connection info if successful.
func (RhymenConnector) Connect(timeout time.Duration) (whatsapp.Conn, error) {
	return whatsapp.NewRhymenConn(timeout)
}
