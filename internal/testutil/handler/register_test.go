package handler

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/r-erema/wapi/internal/http/handler/session"
	sessionModel "github.com/r-erema/wapi/internal/model/session"
	mockAuth "github.com/r-erema/wapi/internal/testutil/mock/auth"
	mockListener "github.com/r-erema/wapi/internal/testutil/mock/listener"
	mockSession "github.com/r-erema/wapi/internal/testutil/mock/session"

	"github.com/Rhymen/go-whatsapp"
	"github.com/gavv/httpexpect/v2"
	"github.com/golang/mock/gomock"
)

func TestRegisterSessionSuccess(t *testing.T) {
	handler := session.NewRegisterSessionHandler(prepareMocks(t))

	server := httptest.NewServer(handler)
	defer server.Close()

	e := httpexpect.New(t, server.URL)
	e.POST("/register-session/").
		WithJSON(map[string]string{"session_id": "session_id_token_81E25FCF8393C916D131A81C60AFFEB11"}).
		Expect().
		Status(http.StatusOK)
}

func prepareMocks(t *testing.T) (
	auth *mockAuth.MockAuthorizer,
	listener *mockListener.MockListener,
	sessionWorks *mockSession.MockRepository,
) {
	sessionID := "session_id_token_81E25FCF8393C916D131A81C60AFFEB11"
	mockCtrl := gomock.NewController(t)
	auth = mockAuth.NewMockAuthorizer(mockCtrl)
	auth.EXPECT().Login(sessionID).Return(&whatsapp.Conn{}, &sessionModel.WapiSession{}, nil)

	listener = mockListener.NewMockListener(mockCtrl)
	listener.EXPECT().
		ListenForSession(gomock.Any(), gomock.Any()).
		Do(func(sessionId string, wg *sync.WaitGroup) { wg.Done() })
	sessionWorks = mockSession.NewMockRepository(mockCtrl)
	return
}
