package service

import (
	"fmt"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
)

// Stores and monitors the states of connections.
type Connections interface {
	// Binds session and connection together.
	AddAuthenticatedConnectionForSession(sessionID string, sessConnDTO *SessionConnectionDTO) error
	// Unbinds session and connection.
	RemoveConnectionForSession(sessionID string)
	// Gets connection of specific session.
	AuthenticatedConnectionForSession(sessionID string) (*SessionConnectionDTO, error)
}

// ConnectionsPool stores and checks connections state.
type ConnectionsPool struct {
	connectionSessionPool map[string]*SessionConnectionDTO
	pingDevicesDuration   time.Duration
}

// Error for case of not found connection.
type NotFoundError struct {
	SessionID string
}

func (m *NotFoundError) Error() string {
	return fmt.Sprintf("connection for session `%s` not found", m.SessionID)
}

// New creates connection supervisor.
func NewSV(pingDevicesDuration time.Duration) *ConnectionsPool {
	return &ConnectionsPool{
		connectionSessionPool: make(map[string]*SessionConnectionDTO),
		pingDevicesDuration:   pingDevicesDuration,
	}
}

// Binds session and connection together.
func (supervisor *ConnectionsPool) AddAuthenticatedConnectionForSession(sessionID string, sessConnDTO *SessionConnectionDTO) error {
	pong, err := sessConnDTO.Wac().AdminTest()
	if !pong || err != nil {
		return fmt.Errorf("connection for session `%s`, not active, couldn't be added: %v", sessionID, err)
	}
	supervisor.RemoveConnectionForSession(sessionID)
	supervisor.connectionSessionPool[sessionID] = sessConnDTO
	supervisor.pingConnection(sessConnDTO)
	return nil
}

// Unbinds session and connection.
func (supervisor *ConnectionsPool) RemoveConnectionForSession(sessionID string) {
	if target, ok := supervisor.connectionSessionPool[sessionID]; ok {
		_, _ = target.Wac().Disconnect()
		*target.pingQuit <- ""
		delete(supervisor.connectionSessionPool, sessionID)
	}
}

// Gets connection of specific session.
func (supervisor *ConnectionsPool) AuthenticatedConnectionForSession(sessionID string) (*SessionConnectionDTO, error) {
	if target, ok := supervisor.connectionSessionPool[sessionID]; ok {
		pong, err := target.Wac().AdminTest()
		if !pong || err != nil {
			return nil, fmt.Errorf("connection for session `%s` existed, but device doesn't response at the moment: %v", sessionID, err)
		}
		return target, nil
	}
	return nil, &NotFoundError{SessionID: sessionID}
}

func (supervisor *ConnectionsPool) pingConnection(sessConn *SessionConnectionDTO) {
	ticker := time.NewTicker(supervisor.pingDevicesDuration * time.Millisecond)
	notificationLimit := 3
	currentFailedAttempt := 0
	currentAttemptResult, previousAttemptResult := true, true
	go func() {
		for {
			select {
			case <-ticker.C:
				pong, err := sessConn.Wac().AdminTest()
				if !pong || err != nil {
					currentAttemptResult = false
					if notificationLimit > currentFailedAttempt {
						msg := fmt.Sprintf(
							"device of session `%s` login `%s` is not responding: %v",
							sessConn.Session().SessionID,
							sessConn.Wac().Info().Wid,
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
						sessConn.Session().SessionID,
						sessConn.Wac().Info().Wid,
					)
					sentry.CaptureMessage(msg)
					log.Printf("warning: %s", msg)
					currentFailedAttempt = 0
				}

				previousAttemptResult = currentAttemptResult
			case <-*sessConn.pingQuit:
				log.Printf("ping connection for session `%s` disabled", sessConn.Session().SessionID)
				ticker.Stop()
				return
			}
		}
	}()
}
