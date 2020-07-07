package service

import (
	"fmt"
	"os"

	osInfra "github.com/r-erema/wapi/internal/infrastructure/os"
)

// Builds qr-code images path.
type QRFileResolver interface {
	// Returns path to image file of qr-code.
	ResolveQrFilePath(sessionID string) string
}

// Builds qr-code images path.
type QRImgResolver struct {
	fs                osInfra.FileSystem
	qrImagesFilesPath string
}

// Creates qr-image files resolver.
func NewQRImgResolver(qrImagesFilesPath string, fs osInfra.FileSystem) (*QRImgResolver, error) {
	if _, err := fs.Stat(qrImagesFilesPath); fs.IsNotExist(err) {
		err := fs.MkdirAll(qrImagesFilesPath, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	return &QRImgResolver{fs: fs, qrImagesFilesPath: qrImagesFilesPath}, nil
}

// Returns path to image file of qr-code.
func (q *QRImgResolver) ResolveQrFilePath(sessionID string) string {
	return fmt.Sprintf("%s/qr_%s.png", q.qrImagesFilesPath, sessionID)
}
