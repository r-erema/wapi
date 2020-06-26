package supervisor

import (
	"fmt"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
)

type ConnectionSupervisor interface {
	AddAuthenticatedConnectionForSession(sessionID string, sessConnDTO *SessionConnectionDTO) error
	RemoveConnectionForSession(sessionID string)
	GetAuthenticatedConnectionForSession(sessionID string) (*SessionConnectionDTO, error)
}

type ConnectionsSupervisor struct {
	connectionSessionPool map[string]*SessionConnectionDTO
	pingDevicesDuration   time.Duration
}

func NewConnectionsSupervisor(pingDevicesDuration time.Duration) *ConnectionsSupervisor {
	return &ConnectionsSupervisor{
		connectionSessionPool: make(map[string]*SessionConnectionDTO),
		pingDevicesDuration:   pingDevicesDuration,
	}
}

func (supervisor *ConnectionsSupervisor) AddAuthenticatedConnectionForSession(sessionID string, sessConnDTO *SessionConnectionDTO) error {
	pong, err := sessConnDTO.GetWac().AdminTest()
	if !pong || err != nil {
		return fmt.Errorf("connection for session `%s`, not active, couldn't be added: %v", sessionID, err)
	}
	supervisor.RemoveConnectionForSession(sessionID)
	supervisor.connectionSessionPool[sessionID] = sessConnDTO
	supervisor.pingConnection(sessConnDTO)
	return nil
}

func (supervisor *ConnectionsSupervisor) RemoveConnectionForSession(sessionID string) {
	if target, ok := supervisor.connectionSessionPool[sessionID]; ok {
		_, _ = target.GetWac().Disconnect()
		*target.pingQuit <- ""
		delete(supervisor.connectionSessionPool, sessionID)
	}
}

func (supervisor *ConnectionsSupervisor) GetAuthenticatedConnectionForSession(sessionID string) (*SessionConnectionDTO, error) {
	if target, ok := supervisor.connectionSessionPool[sessionID]; ok {
		pong, err := target.GetWac().AdminTest()
		if !pong || err != nil {
			return nil, fmt.Errorf("connection for session `%s` existed, but device doesn't response at the moment: %v", sessionID, err)
		}
		return target, nil
	}
	return nil, fmt.Errorf("connection for session `%s` not found", sessionID)
}

func (supervisor *ConnectionsSupervisor) pingConnection(sessConn *SessionConnectionDTO) {
	ticker := time.NewTicker(supervisor.pingDevicesDuration * time.Second)
	notificationLimit := 3
	currentFailedAttempt := 0
	currentAttemptResult, previousAttemptResult := true, true
	go func() {
		for {
			select {
			case <-ticker.C:
				pong, err := sessConn.GetWac().AdminTest()
				if !pong || err != nil {
					currentAttemptResult = false
					if notificationLimit > currentFailedAttempt {
						msg := fmt.Sprintf(
							"device of session `%s` login `%s` is not responding: %v",
							sessConn.GetSession().SessionID,
							sessConn.GetWac().Info.Wid,
							err,
						)
						sentry.CaptureMessage(msg)
						log.Printf("warning: %s", msg)
					}
					currentFailedAttempt++
				} else {
					currentAttemptResult = true
				}

				if !previousAttemptResult && currentAttemptResult {
					msg := fmt.Sprintf(
						"device of session `%s` login `%s` is responding again",
						sessConn.GetSession().SessionID,
						sessConn.GetWac().Info.Wid,
					)
					sentry.CaptureMessage(msg)
					log.Printf("warning: %s", msg)
					currentFailedAttempt = 0
				}

				previousAttemptResult = currentAttemptResult
			case <-*sessConn.pingQuit:
				log.Printf("ping connection for session `%s` disabled", sessConn.GetSession().SessionID)
				ticker.Stop()
				return
			}
		}
	}()
}
