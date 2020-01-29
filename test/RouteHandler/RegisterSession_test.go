package RouteHandler

import (
	"../../test/Mock/Auth"
	"../../test/Mock/MessageListener"
	"../../test/Mock/SessionWorks"
	"RouteHandler"
	"Session"
	"github.com/Rhymen/go-whatsapp"
	"github.com/golang/mock/gomock"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestRegisterSessionSuccess(t *testing.T) {

	handler := RouteHandler.NewRegisterSessionHandler(prepareMocks(t))
	r := httptest.NewRequest(
		"POST",
		"/register-session/",
		strings.NewReader(`{"session_id":"session_id_token_81E25FCF8393C916D131A81C60AFFEB11"}`),
	)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Wrong status %v, expected %v", w.Code, http.StatusOK)
	}
}

func prepareMocks(t *testing.T) (
	auth *mock_Auth.MockInterface,
	listener *mock_MessageListener.MockInterface,
	sessionWorks *mock_SessionWorks.MockInterface,
) {
	sessionId := "session_id_token_81E25FCF8393C916D131A81C60AFFEB11"
	mockCtrl := gomock.NewController(t)
	auth = mock_Auth.NewMockInterface(mockCtrl)
	auth.EXPECT().Login(sessionId).Return(&whatsapp.Conn{}, &Session.WapiSession{}, nil)

	listener = mock_MessageListener.NewMockInterface(mockCtrl)
	listener.EXPECT().
		ListenForSession(gomock.Any(), gomock.Any()).
		Do(func(sessionId string, wg *sync.WaitGroup) { wg.Done() })
	sessionWorks = mock_SessionWorks.NewMockInterface(mockCtrl)
	return
}
