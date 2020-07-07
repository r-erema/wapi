package auth

import (
	"fmt"
	"log"
	"os"
	"time"

	qrCode "github.com/Baozisoftware/qrcode-terminal-go"
	whatsappRhymen "github.com/Rhymen/go-whatsapp"
	"github.com/r-erema/wapi/internal/infrastructure/whatsapp"
	sessionModel "github.com/r-erema/wapi/internal/model/session"
	sessionRepo "github.com/r-erema/wapi/internal/repository/session"
	"github.com/r-erema/wapi/internal/service/supervisor"
	"github.com/skip2/go-qrcode"
)

// Authorizer responsible for users authorization.
type Authorizer interface {
	// Authorizes user whether by stored session file or by qr-code.
	Login(sessionID string) (whatsapp.Conn, *sessionModel.WapiSession, error)
}

// Auth responsible for users authorization using qr-code or stored session.
type Auth struct {
	QrImagesFilesPath     string
	timeoutConnection     time.Duration
	SessionWorks          sessionRepo.Repository
	connectionsSupervisor supervisor.Connections
}

// New creates Auth service.
func New(
	qrImagesFilesPath string,
	timeoutConnection time.Duration,
	sessionWorks sessionRepo.Repository,
	connectionsSupervisor supervisor.Connections,
) (*Auth, error) {
	if _, err := os.Stat(qrImagesFilesPath); os.IsNotExist(err) {
		err := os.MkdirAll(qrImagesFilesPath, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	return &Auth{
		QrImagesFilesPath:     qrImagesFilesPath,
		timeoutConnection:     timeoutConnection,
		SessionWorks:          sessionWorks,
		connectionsSupervisor: connectionsSupervisor,
	}, nil
}

// Authorizes user whether by stored session file or by qr-code.
func (auth *Auth) Login(sessionID string) (whatsapp.Conn, *sessionModel.WapiSession, error) {
	wac, err := whatsapp.NewRhymenConn(auth.timeoutConnection)
	if err != nil {
		return nil, nil, fmt.Errorf("create connection failed for session `%s`: %v", sessionID, err)
	}

	wapiSession, err := auth.SessionWorks.ReadSession(sessionID)
	if err == nil {
		_, err = wac.RestoreWithSession(*wapiSession.WhatsAppSession)
		if err != nil {
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
			err = qrcode.WriteFile(qrData, qrcode.Medium, 256, auth.ResolveQrFilePath(sessionID))
			if err != nil {
				log.Printf("can't save QR-code as file: %v", err)
			}
		}()
		var session whatsappRhymen.Session
		session, err = wac.Login(qr)
		removeErr := os.Remove(auth.ResolveQrFilePath(sessionID))
		if removeErr != nil {
			log.Printf("can't remove qr image: %v\n", err)
		}

		if err != nil {
			return nil, nil, fmt.Errorf("error during login: %v", err)
		}
		wapiSession = &sessionModel.WapiSession{SessionID: sessionID, WhatsAppSession: &session}
	}

	if err = auth.connectionsSupervisor.AddAuthenticatedConnectionForSession(
		sessionID,
		supervisor.NewDTO(wac, wapiSession),
	); err != nil {
		return nil, nil, fmt.Errorf("error adding connection to supervisor: %v", err)
	}

	err = auth.SessionWorks.WriteSession(wapiSession)
	if err != nil {
		return wac, wapiSession, fmt.Errorf("error saving session: %v", err)
	}
	return wac, wapiSession, nil
}

// Returns path to image file of qr-code.
func (auth *Auth) ResolveQrFilePath(sessionID string) string {
	return fmt.Sprintf("%s/qr_%s.png", auth.QrImagesFilesPath, sessionID)
}
