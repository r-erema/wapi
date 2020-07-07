package http

import (
	"encoding/json"
	"log"
	"net/http"

	jsonInfra "github.com/r-erema/wapi/internal/infrastructure/json"
	"github.com/r-erema/wapi/internal/service"

	"github.com/Rhymen/go-whatsapp"
)

// SendTextMessageHandler responsible for sending text messages.
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

func (handler *SendTextMessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var msgReq SendMessageRequest
	err := decoder.Decode(&msgReq)
	if err != nil {
		errorPrefix := "can't decode request"
		http.Error(w, errorPrefix, http.StatusBadRequest)
		log.Printf("%s: %v\n", errorPrefix, err)
		return
	}

	sessConnDTO, err := handler.connectionsSupervisor.AuthenticatedConnectionForSession(msgReq.SessionID)
	if err != nil {
		errorPrefix := "session not registered"
		http.Error(w, errorPrefix, http.StatusBadRequest)
		log.Printf("%s: %v\n", errorPrefix, err)
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

	_, err = wac.Send(message)
	if err != nil {
		errorPrefix := "sending message error"
		http.Error(w, errorPrefix, http.StatusInternalServerError)
		log.Printf("%s: %v\n", errorPrefix, err)
	}
	log.Printf("message sent to %s by session %s \n", msgReq.ChatID, msgReq.SessionID)
	marshal := *handler.marshal
	responseBody, err := marshal(&message)
	if err != nil {
		errorPrefix := "error message marshaling"
		http.Error(w, errorPrefix, http.StatusInternalServerError)
		log.Printf("%s: %v\n", errorPrefix, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(responseBody)
	if err != nil {
		errorPrefix := "can't write body to response"
		http.Error(w, errorPrefix, http.StatusInternalServerError)
		log.Printf("%s: %v\n", errorPrefix, err)
		return
	}
}

// SendMessageRequest is the request for sending text message to WhatsApp.
type SendMessageRequest struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	SessionID string `json:"session_name"`
}
