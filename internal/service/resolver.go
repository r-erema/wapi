package service

import (
	"fmt"
	"os"
)

// Builds qr-code images path.
type QRFileResolver interface {
	// Returns path to image file of qr-code.
	ResolveQrFilePath(sessionID string) string
}

// Builds qr-code images path.
type QRImgResolver struct {
	qrImagesFilesPath string
}

// Creates qr-image files resolver.
func NewQRImgResolver(qrImagesFilesPath string) (*QRImgResolver, error) {
	if _, err := os.Stat(qrImagesFilesPath); os.IsNotExist(err) {
		err := os.MkdirAll(qrImagesFilesPath, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	return &QRImgResolver{qrImagesFilesPath: qrImagesFilesPath}, nil
}

// Returns path to image file of qr-code.
func (q *QRImgResolver) ResolveQrFilePath(sessionID string) string {
	return fmt.Sprintf("%s/qr_%s.png", q.qrImagesFilesPath, sessionID)
}
