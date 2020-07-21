package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pkg/errors"
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

// NewTextHandler creates SendTextMessageHandler.j'
func NewTextHandler(
	authorizer service.Authorizer,
	connectionsSupervisor service.Connections,
	marshal *jsonInfra.MarshallCallback,
) *SendTextMessageHandler {
	return &SendTextMessageHandler{auth: authorizer, connectionsSupervisor: connectionsSupervisor, marshal: marshal}
}

// Handle sends text message to WhatsApp server.
func (handler *SendTextMessageHandler) Handle(w http.ResponseWriter, r *http.Request) *AppError {
	decoder := json.NewDecoder(r.Body)
	var msgReq SendMessageRequest
	err := decoder.Decode(&msgReq)
	if err != nil {
		return &AppError{
			Error:       errors.Wrap(err, "decoding error in text handler"),
			ResponseMsg: "can't decode request",
			Code:        http.StatusBadRequest,
		}
	}

	sessConnDTO, err := handler.connectionsSupervisor.AuthenticatedConnectionForSession(msgReq.SessionID)
	if err != nil {
		return &AppError{
			Error:       errors.Wrap(err, "session not registered in text handler"),
			ResponseMsg: "session not registered",
			Code:        http.StatusBadRequest,
		}
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
		return &AppError{
			Error:       errors.Wrap(err, "sending message error in text handler"),
			ResponseMsg: "sending message error",
			Code:        http.StatusInternalServerError,
		}
	}
	log.Printf("message sent to %s by session %s \n", msgReq.ChatID, msgReq.SessionID)
	marshal := *handler.marshal
	responseBody, err := marshal(&message)
	if err != nil {
		return &AppError{
			Error:       errors.Wrap(err, "error message marshaling in text handler"),
			ResponseMsg: "error message marshaling",
			Code:        http.StatusInternalServerError,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(responseBody); err != nil {
		return &AppError{
			Error:       errors.Wrap(err, "can't write body to response in text handler"),
			ResponseMsg: "can't write body to response",
			Code:        http.StatusInternalServerError,
		}
	}

	return nil
}

// SendMessageRequest is the request for sending text message to WhatsApp.
type SendMessageRequest struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	SessionID string `json:"session_name"`
}
