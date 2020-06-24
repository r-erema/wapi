package RouteHandler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/r-erema/wapi/src/Service/Auth"
	"github.com/r-erema/wapi/src/Service/ConnectionsSupervisor"

	"github.com/Rhymen/go-whatsapp"
)

type SendTextMessageHandler struct {
	auth                  Auth.Interface
	connectionsSupervisor ConnectionsSupervisor.Interface
}

func NewSendMessageHandler(auth Auth.Interface, connectionsSupervisor ConnectionsSupervisor.Interface) *SendTextMessageHandler {
	return &SendTextMessageHandler{auth: auth, connectionsSupervisor: connectionsSupervisor}
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

	sessConnDTO, err := handler.connectionsSupervisor.GetAuthenticatedConnectionForSession(msgReq.SessionId)
	if err != nil {
		errorPrefix := "session not registered"
		http.Error(w, errorPrefix, http.StatusBadRequest)
		log.Printf("%s: %v\n", errorPrefix, err)
		return
	}
	wac := sessConnDTO.GetWac()

	message := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: msgReq.ChatId,
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
	log.Printf("message sent to %s by session %s \n", msgReq.ChatId, msgReq.SessionId)
	responseBody, err := json.Marshal(&message)
	if err != nil {
		log.Println("error message marshalling", err)
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
	ChatId    string `json:"chat_id"`
	Text      string `json:"text"`
	SessionId string `json:"session_name"`
}
