package RouteHandler

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/r-erema/wapi/src/Service/QRFileResolver"
)

type GetQRImageHandler struct {
	qrFileResolver QRFileResolver.Interface
}

func NewGetQRImageHandler(qrFileResolver QRFileResolver.Interface) *GetQRImageHandler {
	return &GetQRImageHandler{qrFileResolver: qrFileResolver}
}

func (handler *GetQRImageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sessionId := params["sessionId"]
	qrImagePath := handler.qrFileResolver.ResolveQrFilePath(sessionId)
	if _, err := os.Stat(qrImagePath); os.IsNotExist(err) {
		errPrefix := "QR image not found"
		http.Error(w, errPrefix, http.StatusNotFound)
		log.Printf("%s: %v", errPrefix, err)
		return
	}

	file, err := os.Open(qrImagePath)
	if err != nil {
		errPrefix := "can't open qr image file"
		http.Error(w, errPrefix, http.StatusInternalServerError)
		log.Printf("%s: %v", errPrefix, err)
		return
	}

	buffer := new(bytes.Buffer)
	_, err = io.Copy(buffer, file)
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
