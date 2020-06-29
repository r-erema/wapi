package file

// Builds qr-code images path.
type QRFileResolver interface {
	// Returns path to image file of qr-code.
	ResolveQrFilePath(sessionID string) string
}
