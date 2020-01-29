package MessageListener

import "sync"

type Interface interface {
	ListenForSession(sessionId string, wg *sync.WaitGroup)
}
