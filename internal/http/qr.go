package http

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/r-erema/wapi/internal/infrastructure/os"
	"github.com/r-erema/wapi/internal/service"
)

// ImageHandler returns QR image.
type ImageHandler struct {
	fs             os.FileSystem
	qrFileResolver service.QRFileResolver
}

// NewInfo creates ImageHandler.
func NewQR(fs os.FileSystem, qrFileResolver service.QRFileResolver) *ImageHandler {
	return &ImageHandler{fs: fs, qrFileResolver: qrFileResolver}
}

func (h *ImageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sessionID := params["sessionID"]
	qrImagePath := h.qrFileResolver.ResolveQrFilePath(sessionID)
	if _, err := h.fs.Stat(qrImagePath); h.fs.IsNotExist(err) {
		errPrefix := "QR image not found"
		http.Error(w, errPrefix, http.StatusNotFound)
		log.Printf("%s: %v", errPrefix, err)
		return
	}

	f, err := h.fs.Open(qrImagePath)
	if err != nil {
		errPrefix := "can't open qr image file"
		http.Error(w, errPrefix, http.StatusInternalServerError)
		log.Printf("%s: %v", errPrefix, err)
		return
	}

	buffer := new(bytes.Buffer)
	_, err = io.Copy(buffer, f)
	if err != nil {
		errPrefix := "can't handle qr image file"
		http.Error(w, errPrefix, http.StatusInternalServerError)
		log.Printf("%s: %v", errPrefix, err)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		errPrefix := "unable to write image"
		http.Error(w, errPrefix, http.StatusInternalServerError)
		log.Printf("%s: %v", errPrefix, err)
		return
	}
}
