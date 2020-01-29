package ConnectionsSupervisor

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"log"
	"time"
)

type connectionsSupervisor struct {
	connectionSessionPool   map[string]*SessionConnectionDTO
	pingDevicesDurationSecs time.Duration
}

func NewConnectionsSupervisor(pingDevicesDurationSecs time.Duration) *connectionsSupervisor {
	return &connectionsSupervisor{connectionSessionPool: make(map[string]*SessionConnectionDTO), pingDevicesDurationSecs: pingDevicesDurationSecs}
}

func (supervisor *connectionsSupervisor) AddAuthenticatedConnectionForSession(sessionId string, sessConnDTO *SessionConnectionDTO) error {
	pong, err := sessConnDTO.GetWac().AdminTest()
	if !pong || err != nil {
		return fmt.Errorf("connection for session `%s`, not active, couldn't be added: %v\n", sessionId, err)
	}
	supervisor.RemoveConnectionForSession(sessionId)
	supervisor.connectionSessionPool[sessionId] = sessConnDTO
	supervisor.pingConnection(sessConnDTO)
	return nil
}

func (supervisor *connectionsSupervisor) RemoveConnectionForSession(sessionId string) {
	if target, ok := supervisor.connectionSessionPool[sessionId]; ok {
		_, _ = target.GetWac().Disconnect()
		*target.pingQuit <- ""
		delete(supervisor.connectionSessionPool, sessionId)
	}
}

func (supervisor *connectionsSupervisor) GetAuthenticatedConnectionForSession(sessionId string) (*SessionConnectionDTO, error) {
	if target, ok := supervisor.connectionSessionPool[sessionId]; ok {
		pong, err := target.GetWac().AdminTest()
		if !pong || err != nil {
			return nil, fmt.Errorf("connection for session `%s` existed, but device doesn't response at the moment: %v\n", sessionId, err)
		}
		return target, nil
	}
	return nil, fmt.Errorf("connection for session `%s` not found\n", sessionId)
}

func (supervisor *connectionsSupervisor) pingConnection(sessConn *SessionConnectionDTO) {
	ticker := time.NewTicker(supervisor.pingDevicesDurationSecs * time.Second)
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
							sessConn.GetSession().SessionId,
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
						sessConn.GetSession().SessionId,
						sessConn.GetWac().Info.Wid,
					)
					sentry.CaptureMessage(msg)
					log.Printf("warning: %s", msg)
					currentFailedAttempt = 0
				}

				previousAttemptResult = currentAttemptResult
			case <-*sessConn.pingQuit:
				log.Printf("ping connection for session `%s` disabled", sessConn.GetSession().SessionId)
				ticker.Stop()
				return
			}
		}
	}()
}
