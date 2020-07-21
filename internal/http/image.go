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
	"github.com/pkg/errors"
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

// Handle sends message with client`s image to WhatsApp server.
func (h *SendImageHandler) Handle(w http.ResponseWriter, r *http.Request) *AppError {
	decoder := json.NewDecoder(r.Body)
	var msgReq SendImageRequest
	err := decoder.Decode(&msgReq)
	if err != nil {
		return &AppError{
			Error:       errors.Wrap(err, "can't decode request in image handler"),
			ResponseMsg: "can't decode request",
			Code:        http.StatusBadRequest,
		}
	}

	sessConnDTO, err := h.connectionsSupervisor.AuthenticatedConnectionForSession(msgReq.SessionID)
	if err != nil {
		return &AppError{
			Error:       errors.Wrap(err, "can't find session in image handler"),
			ResponseMsg: "session not registered",
			Code:        http.StatusBadRequest,
		}
	}
	wac := sessConnDTO.Wac()

	response, err := h.httpClient.Get(msgReq.ImageURL)
	if err != nil {
		return &AppError{
			Error:       errors.Wrap(err, "can't get image by this url"),
			ResponseMsg: "image url error",
			Code:        http.StatusInternalServerError,
		}
	}
	defer func() {
		err = response.Body.Close()
		log.Printf("%s: %v\n", "response body closing error", err)
	}()

	img, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return &AppError{
			Error:       errors.Wrap(err, "reading image error in image handler"),
			ResponseMsg: "reading image error",
			Code:        http.StatusInternalServerError,
		}
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
		return &AppError{
			Error:       errors.Wrap(err, "sending message error in image handler"),
			ResponseMsg: "sending message error",
			Code:        http.StatusInternalServerError,
		}
	}
	log.Printf("message sent to %s by session %s \n", msgReq.ChatID, msgReq.SessionID)
	if err := h.writeMsgToResponse(&message, w); err != nil {
		return &AppError{
			Error:       errors.Wrap(err, "error writing message to response"),
			ResponseMsg: "can't send image",
			Code:        http.StatusInternalServerError,
		}
	}
	return nil
}

func (h *SendImageHandler) writeMsgToResponse(msg *whatsapp.ImageMessage, w http.ResponseWriter) error {
	marshal := *h.marshal
	responseBody, err := marshal(msg)
	if err != nil {
		return errors.Wrap(err, "error message marshaling")
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(responseBody); err != nil {
		return errors.Wrap(err, "can't write body to response")
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
