package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/r-erema/wapi/internal/http/handler/session"
	sessionModel "github.com/r-erema/wapi/internal/model/session"
	mockAuth "github.com/r-erema/wapi/internal/testutil/mock/auth"
	mockListener "github.com/r-erema/wapi/internal/testutil/mock/listener"
	mockSession "github.com/r-erema/wapi/internal/testutil/mock/session"

	"github.com/Rhymen/go-whatsapp"
	"github.com/golang/mock/gomock"
)

func TestRegisterSessionSuccess(t *testing.T) {
	handler := session.NewRegisterSessionHandler(prepareMocks(t))
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
