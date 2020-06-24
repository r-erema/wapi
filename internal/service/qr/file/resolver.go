package file

type QRFileResolver interface {
	ResolveQrFilePath(sessionId string) string
}
