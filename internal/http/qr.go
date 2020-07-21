package http

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"github.com/r-erema/wapi/internal/infrastructure/os"
	"github.com/r-erema/wapi/internal/service"

	"github.com/gorilla/mux"
)

// QRHandler represents a handler for working with the QR image.
type QRHandler struct {
	fs             os.FileSystem
	qrFileResolver service.QRFileResolver
}

// NewQR creates QRHandler.
func NewQR(fs os.FileSystem, qrFileResolver service.QRFileResolver) *QRHandler {
	return &QRHandler{fs: fs, qrFileResolver: qrFileResolver}
}

// Handle sends QR-code image.
func (h *QRHandler) Handle(w http.ResponseWriter, r *http.Request) *AppError {
	params := mux.Vars(r)
	sessionID := params["sessionID"]
	qrImagePath := h.qrFileResolver.ResolveQrFilePath(sessionID)
	if _, err := h.fs.Stat(qrImagePath); h.fs.IsNotExist(err) {
		return &AppError{
			Error:       errors.Wrap(err, "QR image not found in qr handler"),
			ResponseMsg: "QR image not found",
			Code:        http.StatusNotFound,
		}
	}

	f, err := h.fs.Open(qrImagePath)
	if err != nil {
		return &AppError{
			Error:       errors.Wrap(err, "can't open qr image file in qr handler"),
			ResponseMsg: "can't open qr image file",
			Code:        http.StatusInternalServerError,
		}
	}

	buffer := new(bytes.Buffer)
	if _, err := io.Copy(buffer, f); err != nil {
		return &AppError{
			Error:       errors.Wrap(err, "can't Handle qr image file in qr handler"),
			ResponseMsg: "can't Handle qr image file",
			Code:        http.StatusInternalServerError,
		}
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		return &AppError{
			Error:       errors.Wrap(err, "unable to write image in qr handler"),
			ResponseMsg: "unable to write image",
			Code:        http.StatusInternalServerError,
		}
	}

	return nil
}
