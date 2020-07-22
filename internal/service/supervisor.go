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

const defaultNotificationsLimit = 3

type notificationState struct {
	notificationLimit, currentFailedAttempt     int
	currentAttemptResult, previousAttemptResult bool
}

// ConnectionsPool stores and checks connections state.
type ConnectionsPool struct {
	connectionSessionPool map[string]*SessionConnectionDTO
	pingDevicesDuration   time.Duration
	notifications         notificationState
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
		notifications: notificationState{
			notificationLimit:     defaultNotificationsLimit,
			currentFailedAttempt:  0,
			currentAttemptResult:  true,
			previousAttemptResult: true,
		},
	}
}

// Binds session and connection together.
func (supervisor *ConnectionsPool) AddAuthenticatedConnectionForSession(sessionID string, sessConnDTO *SessionConnectionDTO) error {
	pong, err := sessConnDTO.Wac().AdminTest()
	if !pong || err != nil {
		return fmt.Errorf("connection for session `%s`, not active, couldn't be added: %w", sessionID, err)
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
			return nil, fmt.Errorf("connection for session `%s` existed, but device doesn't response at the moment: %w", sessionID, err)
		}
		return target, nil
	}
	return nil, &NotFoundError{SessionID: sessionID}
}

func (supervisor *ConnectionsPool) pingConnection(sessConn *SessionConnectionDTO) {
	ticker := time.NewTicker(supervisor.pingDevicesDuration * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				supervisor.notificationLogic(sessConn)
			case <-*sessConn.pingQuit:
				log.Printf("ping connection for session `%s` disabled", sessConn.Session().SessionID)
				ticker.Stop()
				return
			}
		}
	}()
}

func (supervisor *ConnectionsPool) notificationLogic(sessConn *SessionConnectionDTO) {
	pong, err := sessConn.Wac().AdminTest()
	if !pong || err != nil {
		supervisor.notifications.currentAttemptResult = false
		if supervisor.notifications.notificationLimit > supervisor.notifications.currentFailedAttempt {
			notifyDeviceUnavailable(sessConn, err)
		}
		supervisor.notifications.currentFailedAttempt++
	} else {
		supervisor.notifications.currentAttemptResult = true
	}

	if !supervisor.notifications.previousAttemptResult && supervisor.notifications.currentAttemptResult {
		notifyDeviceActiveAgain(sessConn)
		supervisor.notifications.currentFailedAttempt = 0
	}

	supervisor.notifications.previousAttemptResult = supervisor.notifications.currentAttemptResult
}

func notifyDeviceActiveAgain(sessConn *SessionConnectionDTO) {
	msg := fmt.Sprintf(
		"device of session `%s` login `%s` is responding again",
		sessConn.Session().SessionID,
		sessConn.Wac().Info().Wid,
	)
	sentry.CaptureMessage(msg)
	log.Printf("warning: %s", msg)
}

func notifyDeviceUnavailable(sessConn *SessionConnectionDTO, reason error) {
	msg := fmt.Sprintf(
		"device of session `%s` login `%s` is not responding: %v",
		sessConn.Session().SessionID,
		sessConn.Wac().Info().Wid,
		reason,
	)
	sentry.CaptureMessage(msg)
	log.Printf("warning: %s", msg)
}
