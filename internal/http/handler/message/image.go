package message

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Rhymen/go-whatsapp"
	httpInfra "github.com/r-erema/wapi/internal/infrastructure/http"
	jsonInfra "github.com/r-erema/wapi/internal/infrastructure/json"
	"github.com/r-erema/wapi/internal/service/auth"
	"github.com/r-erema/wapi/internal/service/supervisor"
)

// SendImageHandler responsible for sending images.
type SendImageHandler struct {
	auth                  auth.Authorizer
	connectionsSupervisor supervisor.Connections
	httpClient            httpInfra.Client
	marshal               *jsonInfra.MarshallCallback
}

// NewImageHandler creates SendImageHandler.
func NewImageHandler(
	authorizer auth.Authorizer,
	connectionsSupervisor supervisor.Connections,
	client httpInfra.Client,
	marshal *jsonInfra.MarshallCallback,
) *SendImageHandler {
	return &SendImageHandler{
		auth:                  authorizer,
		connectionsSupervisor: connectionsSupervisor,
		httpClient:            client,
		marshal:               marshal,
	}
}

func (h *SendImageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var msgReq SendImageRequest
	err := decoder.Decode(&msgReq)
	if err != nil {
		var errorPrefix = "can't decode request"
		http.Error(w, errorPrefix, http.StatusBadRequest)
		log.Printf("%s: %v\n", errorPrefix, err)
		return
	}

	sessConnDTO, err := h.connectionsSupervisor.AuthenticatedConnectionForSession(msgReq.SessionID)
	if err != nil {
		var errorPrefix = "session not registered"
		http.Error(w, errorPrefix, http.StatusBadRequest)
		log.Printf("%s: %v\n", errorPrefix, err)
		return
	}
	wac := sessConnDTO.Wac()

	response, err := h.httpClient.Get(msgReq.ImageURL)
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
			SenderJid: wac.Info().Wid,
		},
		Type:    http.DetectContentType(img),
		Content: response.Body,
		Caption: msgReq.Caption,
	}

	_, err = wac.Send(message)
	if err != nil {
		var errorPrefix = "sending message error"
		http.Error(w, errorPrefix, http.StatusInternalServerError)
		log.Printf("%s: %v\n", errorPrefix, err)
	}
	log.Printf("message sent to %s by session %s \n", msgReq.ChatID, msgReq.SessionID)
	marshal := *h.marshal
	responseBody, err := marshal(&message)
	if err != nil {
		var errorPrefix = "error message marshaling"
		http.Error(w, errorPrefix, http.StatusInternalServerError)
		log.Printf("%s: %v\n", errorPrefix, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(responseBody)
	if err != nil {
		var errorPrefix = "can't write body to response"
		http.Error(w, errorPrefix, http.StatusInternalServerError)
		log.Printf("%s: %v\n", errorPrefix, err)
		return
	}
}

// SendImageRequest is the request for sending image to WhatsApp.
type SendImageRequest struct {
	SessionID string `json:"session_name"`
	ChatID    string `json:"chat_id"`
	ImageURL  string `json:"image_url"`
	Caption   string `json:"caption"`
}
