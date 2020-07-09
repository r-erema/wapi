package http

import (
	"encoding/json"
	"log"
	"net/http"

	jsonInfra "github.com/r-erema/wapi/internal/infrastructure/json"
	"github.com/r-erema/wapi/internal/service"

	"github.com/Rhymen/go-whatsapp"
)

// SendTextMessageHandler is responsible for sending text messages.
type SendTextMessageHandler struct {
	auth                  service.Authorizer
	connectionsSupervisor service.Connections
	marshal               *jsonInfra.MarshallCallback
}

// NewTextHandler creates SendTextMessageHandler.
func NewTextHandler(
	authorizer service.Authorizer,
	connectionsSupervisor service.Connections,
	marshal *jsonInfra.MarshallCallback,
) *SendTextMessageHandler {
	return &SendTextMessageHandler{auth: authorizer, connectionsSupervisor: connectionsSupervisor, marshal: marshal}
}

// ServeHTTP sends text message to WhatsApp server.
func (handler *SendTextMessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var msgReq SendMessageRequest
	err := decoder.Decode(&msgReq)
	if err != nil {
		handleError(w, "can't decode request", err, http.StatusBadRequest)
		return
	}

	sessConnDTO, err := handler.connectionsSupervisor.AuthenticatedConnectionForSession(msgReq.SessionID)
	if err != nil {
		handleError(w, "session not registered", err, http.StatusBadRequest)
		return
	}
	wac := sessConnDTO.Wac()

	message := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: msgReq.ChatID,
			SenderJid: wac.Info().Wid,
		},
		Text: msgReq.Text,
	}

	if _, err = wac.Send(message); err != nil {
		handleError(w, "sending message error", err, http.StatusInternalServerError)
	}
	log.Printf("message sent to %s by session %s \n", msgReq.ChatID, msgReq.SessionID)
	marshal := *handler.marshal
	responseBody, err := marshal(&message)
	if err != nil {
		handleError(w, "error message marshaling", err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(responseBody); err != nil {
		handleError(w, "can't write body to response", err, http.StatusInternalServerError)
		return
	}
}

// SendMessageRequest is the request for sending text message to WhatsApp.
type SendMessageRequest struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	SessionID string `json:"session_name"`
}
