package auth

import (
	"fmt"
	"log"
	"os"
	"time"

	sessionModel "github.com/r-erema/wapi/internal/model/session"
	sessionRepo "github.com/r-erema/wapi/internal/repository/session"
	"github.com/r-erema/wapi/internal/service/supervisor"

	"github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Rhymen/go-whatsapp"
	"github.com/skip2/go-qrcode"
)

type Authorizer interface {
	Login(sessionId string) (*whatsapp.Conn, *sessionModel.WapiSession, error)
}

type auth struct {
	QrImagesFilesPath        string
	timeoutConnectionSeconds time.Duration
	SessionWorks             sessionRepo.Repository
	connectionsSupervisor    supervisor.ConnectionSupervisor
}

func NewAuth(
	qrImagesFilesPath string,
	timeoutConnectionSeconds time.Duration,
	sessionWorks sessionRepo.Repository,
	connectionsSupervisor supervisor.ConnectionSupervisor,
) (*auth, error) {

	if _, err := os.Stat(qrImagesFilesPath); os.IsNotExist(err) {
		err := os.MkdirAll(qrImagesFilesPath, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	return &auth{
		QrImagesFilesPath:        qrImagesFilesPath,
		timeoutConnectionSeconds: timeoutConnectionSeconds,
		SessionWorks:             sessionWorks,
		connectionsSupervisor:    connectionsSupervisor,
	}, nil
}

func (auth *auth) Login(sessionId string) (*whatsapp.Conn, *sessionModel.WapiSession, error) {
	wac, err := whatsapp.NewConn(auth.timeoutConnectionSeconds)
	if err != nil {
		return nil, nil, fmt.Errorf("create connection failed for session `%s`: %v\n", sessionId, err)
	}

	wapiSession, err := auth.SessionWorks.ReadSession(sessionId)
	if err == nil {
		_, err := wac.RestoreWithSession(*wapiSession.WhatsAppSession)
		if err != nil {
			removeSessionFileTxt := ""
			if err.Error() == "admin login responded with 401" {
				_ = auth.SessionWorks.RemoveSession(wapiSession.SessionId)
				removeSessionFileTxt = ", probably logout happened on the phone, session file will be removed"
			}
			return nil, nil, fmt.Errorf("restoring failed: %v%v\n", err, removeSessionFileTxt)
		}
	} else {
		qr := make(chan string)
		go func() {
			qrData := <-qr
			terminal := qrcodeTerminal.New()
			terminal.Get(qrData).Print()
			err := qrcode.WriteFile(qrData, qrcode.Medium, 256, auth.ResolveQrFilePath(sessionId))
			if err != nil {
				log.Printf("can't save QR-code as file: %v", err)
			}
		}()
		session, err := wac.Login(qr)
		removeErr := os.Remove(auth.ResolveQrFilePath(sessionId))
		if removeErr != nil {
			log.Printf("can't remove qr image: %v\n", err)
		}

		if err != nil {
			return nil, nil, fmt.Errorf("error during login: %v\n", err)
		}
		wapiSession = &sessionModel.WapiSession{SessionId: sessionId, WhatsAppSession: &session}
	}

	if err := auth.connectionsSupervisor.AddAuthenticatedConnectionForSession(
		sessionId,
		supervisor.NewSessionConnectionDTO(wac, wapiSession),
	); err != nil {
		return nil, nil, fmt.Errorf("error adding connection to supervisor: %v\n", err)
	}

	err = auth.SessionWorks.WriteSession(wapiSession)
	if err != nil {
		return wac, wapiSession, fmt.Errorf("error saving session: %v\n", err)
	}
	return wac, wapiSession, nil
}

func (auth *auth) ResolveQrFilePath(sessionId string) string {
	return auth.QrImagesFilesPath + "/qr_" + sessionId + ".png"
}
