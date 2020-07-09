package http

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

	"github.com/r-erema/wapi/internal/infrastructure/os"
	"github.com/r-erema/wapi/internal/service"

	"github.com/gorilla/mux"
)

// ImageHandler represents a handler for working with the QR image.
type ImageHandler struct {
	fs             os.FileSystem
	qrFileResolver service.QRFileResolver
}

// NewQR creates ImageHandler.
func NewQR(fs os.FileSystem, qrFileResolver service.QRFileResolver) *ImageHandler {
	return &ImageHandler{fs: fs, qrFileResolver: qrFileResolver}
}

// ServeHTTP sends QR-code image.
func (h *ImageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sessionID := params["sessionID"]
	qrImagePath := h.qrFileResolver.ResolveQrFilePath(sessionID)
	if _, err := h.fs.Stat(qrImagePath); h.fs.IsNotExist(err) {
		handleError(w, "QR image not found", err, http.StatusNotFound)
		return
	}

	f, err := h.fs.Open(qrImagePath)
	if err != nil {
		handleError(w, "can't open qr image file", err, http.StatusInternalServerError)
		return
	}

	buffer := new(bytes.Buffer)
	if _, err := io.Copy(buffer, f); err != nil {
		handleError(w, "can't handle qr image file", err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		handleError(w, "unable to write image", err, http.StatusInternalServerError)
		return
	}
}
