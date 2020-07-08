package service

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/r-erema/wapi/internal/infrastructure/whatsapp"
	"github.com/r-erema/wapi/internal/model"
	"github.com/r-erema/wapi/internal/repository"

	qrCode "github.com/Baozisoftware/qrcode-terminal-go"
	whatsappRhymen "github.com/Rhymen/go-whatsapp"
	"github.com/skip2/go-qrcode"
)

// Authorizer is responsible for users authorization.
type Authorizer interface {
	// Login authorizes user whether by stored session file or by qr-code.
	Login(sessionID string) (whatsapp.Conn, *model.WapiSession, error)
}

// Auth is responsible for users authorization using qr-code or stored session.
type Auth struct {
	timeoutConnection     time.Duration
	SessionWorks          repository.Session
	connectionsSupervisor Connections
	fileResolver          QRFileResolver
}

// New creates Auth service.
func New(
	timeoutConnection time.Duration,
	sessionWorks repository.Session,
	connectionsSupervisor Connections,
	fileResolver QRFileResolver,
) *Auth {
	return &Auth{
		timeoutConnection:     timeoutConnection,
		SessionWorks:          sessionWorks,
		connectionsSupervisor: connectionsSupervisor,
		fileResolver:          fileResolver,
	}
}

// Login authorizes user whether by stored session file or by qr-code.
func (auth *Auth) Login(sessionID string) (whatsapp.Conn, *model.WapiSession, error) {
	wac, err := whatsapp.NewRhymenConn(auth.timeoutConnection)
	if err != nil {
		return nil, nil, fmt.Errorf("create connection failed for session `%s`: %v", sessionID, err)
	}

	wapiSession, err := auth.SessionWorks.ReadSession(sessionID)
	if err == nil {
		if _, err = wac.RestoreWithSession(wapiSession.WhatsAppSession); err != nil {
			removeSessionFileTxt := ""
			if err.Error() == "admin login responded with 401" {
				_ = auth.SessionWorks.RemoveSession(wapiSession.SessionID)
				removeSessionFileTxt = ", probably logout happened on the phone, session file will be removed"
			}
			return nil, nil, fmt.Errorf("restoring failed: %v%v", err, removeSessionFileTxt)
		}
	} else {
		qr := make(chan string)
		go func() {
			qrData := <-qr
			terminal := qrCode.New()
			terminal.Get(qrData).Print()
			err = qrcode.WriteFile(qrData, qrcode.Medium, 256, auth.fileResolver.ResolveQrFilePath(sessionID))
			if err != nil {
				log.Printf("can't save QR-code as file: %v", err)
			}
		}()
		var session whatsappRhymen.Session
		session, err = wac.Login(qr)
		removeErr := os.Remove(auth.fileResolver.ResolveQrFilePath(sessionID))
		if removeErr != nil {
			log.Printf("can't remove qr image: %v\n", err)
		}

		if err != nil {
			return nil, nil, fmt.Errorf("error during login: %v", err)
		}
		wapiSession = &model.WapiSession{SessionID: sessionID, WhatsAppSession: &session}
	}

	if err = auth.connectionsSupervisor.AddAuthenticatedConnectionForSession(
		sessionID,
		NewDTO(wac, wapiSession),
	); err != nil {
		return nil, nil, fmt.Errorf("error adding connection to supervisor: %v", err)
	}

	err = auth.SessionWorks.WriteSession(wapiSession)
	if err != nil {
		return wac, wapiSession, fmt.Errorf("error saving session: %v", err)
	}
	return wac, wapiSession, nil
}
