package message

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/r-erema/wapi/internal/service/auth"
	"github.com/r-erema/wapi/internal/service/supervisor"

	"github.com/Rhymen/go-whatsapp"
)

type SendTextMessageHandler struct {
	auth                  auth.Authorizer
	connectionsSupervisor supervisor.Connections
}

// Creates SendTextMessageHandler.
func NewTextHandler(authorizer auth.Authorizer, connectionsSupervisor supervisor.Connections) *SendTextMessageHandler {
	return &SendTextMessageHandler{auth: authorizer, connectionsSupervisor: connectionsSupervisor}
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
			SenderJid: wac.Info.Wid,
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
	responseBody, err := json.Marshal(&message)
	if err != nil {
		log.Println("error message marshaling", err)
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

type SendMessageRequest struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	SessionID string `json:"session_name"`
}
