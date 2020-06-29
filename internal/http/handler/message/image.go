package message

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Rhymen/go-whatsapp"
	"github.com/r-erema/wapi/internal/service/auth"
	"github.com/r-erema/wapi/internal/service/supervisor"
)

// Responsible for sending images.
type SendImageHandler struct {
	auth                  auth.Authorizer
	connectionsSupervisor supervisor.Connections
}

// Creates SendImageHandler.
func NewImageHandler(authorizer auth.Authorizer, connectionsSupervisor supervisor.Connections) *SendImageHandler {
	return &SendImageHandler{auth: authorizer, connectionsSupervisor: connectionsSupervisor}
}

func (handler *SendImageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var msgReq SendImageRequest
	err := decoder.Decode(&msgReq)
	if err != nil {
		const errorPrefix = "can't decode request"
		http.Error(w, errorPrefix, http.StatusBadRequest)
		log.Printf("%s: %v\n", errorPrefix, err)
		return
	}

	sessConnDTO, err := handler.connectionsSupervisor.AuthenticatedConnectionForSession(msgReq.SessionID)
	if err != nil {
		const errorPrefix = "session not registered"
		http.Error(w, errorPrefix, http.StatusBadRequest)
		log.Printf("%s: %v\n", errorPrefix, err)
		return
	}
	wac := sessConnDTO.Wac()

	response, err := http.Get(msgReq.ImageURL)
	if err != nil {
		errorPrefix := "image url error"
		http.Error(w, errorPrefix, http.StatusBadRequest)
		log.Printf("%s: %v\n", errorPrefix, err)
		return
	}
	defer func() {
		err = response.Body.Close()
		log.Printf("%s: %v\n", "response body closing error", err)
	}()

	img, err := ioutil.ReadAll(response.Body)
	if err != nil {
		errorPrefix := "reading image error"
		http.Error(w, errorPrefix, http.StatusBadRequest)
		log.Printf("%s: %v\n", errorPrefix, err)
		return
	}
	response.Body = ioutil.NopCloser(bytes.NewBuffer(img))

	message := whatsapp.ImageMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: msgReq.ChatID,
			SenderJid: wac.Info.Wid,
		},
		Type:    http.DetectContentType(img),
		Content: response.Body,
		Caption: msgReq.Caption,
	}

	_, err = wac.Send(message)
	if err != nil {
		const errorPrefix = "sending message error"
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
		const errorPrefix = "can't write body to response"
		http.Error(w, errorPrefix, http.StatusInternalServerError)
		log.Printf("%s: %v\n", errorPrefix, err)
		return
	}
}

type SendImageRequest struct {
	SessionID string `json:"session_name"`
	ChatID    string `json:"chat_id"`
	ImageURL  string `json:"image_url"`
	Caption   string `json:"caption"`
}
