package service

import (
	"fmt"

	"github.com/r-erema/wapi/internal/infrastructure/os"

	"github.com/pkg/errors"
)

// QRFileResolver builds qr-code images path.
type QRFileResolver interface {
	// Returns path to image file of qr-code.
	ResolveQrFilePath(sessionID string) string
}

// QRImgResolver builds qr-code images path.
type QRImgResolver struct {
	fs                os.FileSystem
	qrImagesFilesPath string
}

// NewQRImgResolver creates qr-image files resolver.
func NewQRImgResolver(qrImagesFilesPath string, fs os.FileSystem) (*QRImgResolver, error) {
	if _, err := fs.Stat(qrImagesFilesPath); fs.IsNotExist(err) {
		err := fs.MkdirAll(qrImagesFilesPath, os.ModePerm)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't create images path")
		}
	}

	return &QRImgResolver{fs: fs, qrImagesFilesPath: qrImagesFilesPath}, nil
}

// ResolveQrFilePath returns path to image file of qr-code.
func (q *QRImgResolver) ResolveQrFilePath(sessionID string) string {
	return fmt.Sprintf("%s/qr_%s.png", q.qrImagesFilesPath, sessionID)
}
