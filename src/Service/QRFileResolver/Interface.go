package QRFileResolver

type Interface interface {
	ResolveQrFilePath(sessionId string) string
}
