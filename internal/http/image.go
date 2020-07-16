package http

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	httpInfra "github.com/r-erema/wapi/internal/infrastructure/http"
	jsonInfra "github.com/r-erema/wapi/internal/infrastructure/json"
	"github.com/r-erema/wapi/internal/service"

	"github.com/Rhymen/go-whatsapp"
)

// SendImageHandler is responsible for sending images.
type SendImageHandler struct {
	auth                  service.Authorizer
	connectionsSupervisor service.Connections
	httpClient            httpInfra.Client
	marshal               *jsonInfra.MarshallCallback
}

// NewImageHandler creates SendImageHandler.
func NewImageHandler(
	authorizer service.Authorizer,
	connectionsSupervisor service.Connections,
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

// ServeHTTP sends message with client`s image to WhatsApp server.
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
		handleError(w, "image url error", err, http.StatusInternalServerError)
		return
	}
	defer func() {
		err = response.Body.Close()
		log.Printf("%s: %v\n", "response body closing error", err)
	}()

	img, err := ioutil.ReadAll(response.Body)
	if err != nil {
		handleError(w, "reading image error", err, http.StatusInternalServerError)
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

	if _, err = wac.Send(message); err != nil {
		handleError(w, "sending message error", err, http.StatusInternalServerError)
	}
	log.Printf("message sent to %s by session %s \n", msgReq.ChatID, msgReq.SessionID)
	if err := h.writeMsgToResponse(&message, w); err != nil {
		return
	}
}

func (h *SendImageHandler) writeMsgToResponse(msg *whatsapp.ImageMessage, w http.ResponseWriter) error {
	marshal := *h.marshal
	responseBody, err := marshal(msg)
	if err != nil {
		handleError(w, "error message marshaling", err, http.StatusInternalServerError)
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(responseBody); err != nil {
		handleError(w, "can't write body to response", err, http.StatusInternalServerError)
		return err
	}
	return nil
}

// SendImageRequest is the request for sending image to WhatsApp.
type SendImageRequest struct {
	SessionID string `json:"session_name"`
	ChatID    string `json:"chat_id"`
	ImageURL  string `json:"image_url"`
	Caption   string `json:"caption"`
}
