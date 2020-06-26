package file

type QRFileResolver interface {
	ResolveQrFilePath(sessionID string) string
}
