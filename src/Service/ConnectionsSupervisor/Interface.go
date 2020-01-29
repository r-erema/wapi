package ConnectionsSupervisor

type Interface interface {
	AddAuthenticatedConnectionForSession(sessionId string, sessConnDTO *SessionConnectionDTO) error
	RemoveConnectionForSession(sessionId string)
	GetAuthenticatedConnectionForSession(sessionId string) (*SessionConnectionDTO, error)
}
