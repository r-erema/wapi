package service

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/r-erema/wapi/internal/infrastructure/whatsapp"
	"github.com/r-erema/wapi/internal/model"
	"github.com/r-erema/wapi/internal/repository"

	terminal "github.com/Baozisoftware/qrcode-terminal-go"
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
	SessionRepo           repository.Session
	connectionsSupervisor Connections
	fileResolver          QRFileResolver
	connector             Connector
	qrDataChan            chan string
}

// NewAuth creates Auth service.
func NewAuth(
	timeoutConnection time.Duration,
	sessionRepo repository.Session,
	connectionsSupervisor Connections,
	fileResolver QRFileResolver,
	connector Connector,
	qrDataChan chan string,
) *Auth {
	return &Auth{
		timeoutConnection:     timeoutConnection,
		SessionRepo:           sessionRepo,
		connectionsSupervisor: connectionsSupervisor,
		fileResolver:          fileResolver,
		connector:             connector,
		qrDataChan:            qrDataChan,
	}
}

// Login authorizes user whether by stored session file or by qr-code.
func (auth *Auth) Login(sessionID string) (whatsapp.Conn, *model.WapiSession, error) {
	wac, err := auth.connector.Connect(auth.timeoutConnection)
	if err != nil {
		return nil, nil, fmt.Errorf("create connection failed for session `%s`: %v", sessionID, err)
	}

	wapiSession, err := auth.SessionRepo.ReadSession(sessionID)
	if err == nil {
		if _, err = wac.RestoreWithSession(wapiSession.WhatsAppSession); err != nil {
			removeSessionFileTxt := ""
			if err.Error() == whatsapp.ErrMsg401 {
				_ = auth.SessionRepo.RemoveSession(wapiSession.SessionID)
				removeSessionFileTxt = ", probably logout happened on the phone, session file will be removed"
			}
			return nil, nil, fmt.Errorf("restoring failed: %v%v", err, removeSessionFileTxt)
		}
	} else {
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			qrData := <-auth.qrDataChan
			t := terminal.New()
			t.Get(qrData).Print()
			errWrite := qrcode.WriteFile(qrData, qrcode.Medium, 256, auth.fileResolver.ResolveQrFilePath(sessionID))
			if errWrite != nil {
				log.Printf("can't save QR-code as file: %v", err)
			}
			wg.Done()
		}()
		var session whatsappRhymen.Session
		session, err = wac.Login(auth.qrDataChan)
		wg.Wait()
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

	err = auth.SessionRepo.WriteSession(wapiSession)
	if err != nil {
		return nil, nil, fmt.Errorf("error saving session: %v", err)
	}
	return wac, wapiSession, nil
}
