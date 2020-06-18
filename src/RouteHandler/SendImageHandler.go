package RouteHandler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Rhymen/go-whatsapp"
	"github.com/r-erema/wapi/src/Service/Auth"
	"github.com/r-erema/wapi/src/Service/ConnectionsSupervisor"
)

type SendImageHandler struct {
	auth                  Auth.Interface
	connectionsSupervisor ConnectionsSupervisor.Interface
}

func NewSendImageHandler(auth Auth.Interface, connectionsSupervisor ConnectionsSupervisor.Interface) *SendImageHandler {
	return &SendImageHandler{auth: auth, connectionsSupervisor: connectionsSupervisor}
}

func (handler *SendImageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var msgReq SendImageRequest
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

	response, err := http.Get(msgReq.ImageUrl)
	if err != nil {
		errorPrefix := "image url error"
		http.Error(w, errorPrefix, http.StatusBadRequest)
		log.Printf("%s: %v\n", errorPrefix, err)
		return
	}
	defer func() {
		_ = response.Body.Close()
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
			RemoteJid: msgReq.ChatId,
			SenderJid: wac.Info.Wid,
		},
		Type:    http.DetectContentType(img),
		Content: response.Body,
		Caption: msgReq.Caption,
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

type SendImageRequest struct {
	SessionId string `json:"session_name"`
	ChatId    string `json:"chat_id"`
	ImageUrl  string `json:"image_url"`
	Caption   string `json:"caption"`
}
