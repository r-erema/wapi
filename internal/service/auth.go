package service

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/r-erema/wapi/internal/infrastructure/whatsapp"
	"github.com/r-erema/wapi/internal/model"
	"github.com/r-erema/wapi/internal/repository"

	terminal "github.com/Baozisoftware/qrcode-terminal-go"
	whatsappRhymen "github.com/Rhymen/go-whatsapp"
	"github.com/pkg/errors"
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
		return nil, nil, errors.Wrapf(err, "create connection failed for session `%s`", sessionID)
	}

	wapiSession, err := auth.SessionRepo.ReadSession(sessionID)
	if err == nil {
		if loginErr := auth.tryLoginBySession(wac, wapiSession); loginErr != nil {
			return nil, nil, errors.Wrap(loginErr, "couldn't login by session")
		}
	} else {
		wapiSession, err = auth.loginByQR(sessionID, wac)
		if err != nil {
			return nil, nil, errors.Wrap(err, "couldn't login by qr-code")
		}
	}

	if err = auth.connectionsSupervisor.AddAuthenticatedConnectionForSession(
		sessionID,
		NewDTO(wac, wapiSession, make(chan string)),
	); err != nil {
		return nil, nil, errors.Wrap(err, "error adding connection to supervisor")
	}

	err = auth.SessionRepo.WriteSession(wapiSession)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error saving session")
	}
	return wac, wapiSession, nil
}

func (auth *Auth) loginByQR(sessionID string, wac whatsapp.Conn) (*model.WapiSession, error) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		qrData := <-auth.qrDataChan
		t := terminal.New()
		t.Get(qrData).Print()
		errWrite := qrcode.WriteFile(qrData, qrcode.Medium, 256, auth.fileResolver.ResolveQrFilePath(sessionID))
		if errWrite != nil {
			log.Printf("can't save QR-code as file: %v", errWrite)
		}
		wg.Done()
	}()
	var session whatsappRhymen.Session
	session, err := wac.Login(auth.qrDataChan)
	wg.Wait()
	removeErr := os.Remove(auth.fileResolver.ResolveQrFilePath(sessionID))
	if removeErr != nil {
		log.Printf("can't remove qr image: %v\n", err)
	}

	if err != nil {
		return nil, errors.Wrap(err, "couldn't login by qr-code")
	}
	return &model.WapiSession{SessionID: sessionID, WhatsAppSession: &session}, nil
}

func (auth *Auth) tryLoginBySession(wac whatsapp.Conn, wapiSession *model.WapiSession) error {
	if _, err := wac.RestoreWithSession(wapiSession.WhatsAppSession); err != nil {
		removeSessionFileTxt := ""
		if err.Error() == whatsapp.ErrMsg401 {
			_ = auth.SessionRepo.RemoveSession(wapiSession.SessionID)
			removeSessionFileTxt = ", probably logout happened on the phone, session file will be removed"
		}
		return errors.Wrapf(err, "restoring failed%s", removeSessionFileTxt)
	}
	return nil
}
