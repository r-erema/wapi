package message

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	httpInfra "github.com/r-erema/wapi/internal/infrastructure/http"
	jsonInfra "github.com/r-erema/wapi/internal/infrastructure/json"
	"github.com/r-erema/wapi/internal/service/auth"
	"github.com/r-erema/wapi/internal/service/supervisor"

	"github.com/Rhymen/go-whatsapp"
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
		handleError(w, "can't decode request", err, http.StatusBadRequest)
		return
	}

	sessConnDTO, err := h.connectionsSupervisor.AuthenticatedConnectionForSession(msgReq.SessionID)
	if err != nil {
		handleError(w, "session not registered", err, http.StatusBadRequest)
		return
	}
	wac := sessConnDTO.Wac()

	response, err := h.httpClient.Get(msgReq.ImageURL)
	if err != nil {
		handleError(w, "image url error", err, http.StatusBadRequest)
		return
	}
	defer func() {
		err = response.Body.Close()
		log.Printf("%s: %v\n", "response body closing error", err)
	}()

	img, err := ioutil.ReadAll(response.Body)
	if err != nil {
		handleError(w, "reading image error", err, http.StatusBadRequest)
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
		handleError(w, "sending message error", err, http.StatusInternalServerError)
	}
	log.Printf("message sent to %s by session %s \n", msgReq.ChatID, msgReq.SessionID)
	marshal := *h.marshal
	responseBody, err := marshal(&message)
	if err != nil {
		handleError(w, "error message marshaling", err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(responseBody)
	if err != nil {
		handleError(w, "can't write body to response", err, http.StatusInternalServerError)
		return
	}
}

func handleError(w http.ResponseWriter, errorPrefix string, err error, httpStatus int) {
	http.Error(w, errorPrefix, httpStatus)
	log.Printf("%s: %v\n", errorPrefix, err)
}

// SendImageRequest is the request for sending image to WhatsApp.
type SendImageRequest struct {
	SessionID string `json:"session_name"`
	ChatID    string `json:"chat_id"`
	ImageURL  string `json:"image_url"`
	Caption   string `json:"caption"`
}
