package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	ListenHttpHostEnvVarKey     = "WAPI_INTERNAL_HOST"
	WhatsAppConnectionTimeout   = "WAPI_WHATSAPP_CONNECTION_TIMEOUT"
	FileSystemRootPoint         = "WAPI_FILE_SYSTEM_ROOT_POINT_FULL_PATH"
	WebHookUrl                  = "WAPI_GETTING_MESSAGES_WEBHOOK"
	RedisHost                   = "WAPI_REDIS_HOST"
	Env                         = "WAPI_ENV"
	CertFilePath                = "WAPI_CERT_FILE_PATH"
	CertKeyPath                 = "WAPI_CERT_KEY_PATH"
	SentryDSN                   = "WAPI_SENTRY_DSN"
	ConnectionsCheckoutDuration = "WAPI_CONNECTIONS_CHECKOUT_DURATION_SECS"

	DevMode  = "dev"
	ProdMode = "prod"
)

type Config struct {
	ListenHttpHost,
	FileSystemRootPath,
	WebHookUrl,
	RedisHost,
	Env,
	CertFilePath,
	HttpStaticFiles,
	SentryDSN,
	CertKeyPath string
	ConnectionsCheckoutDuration,
	ConnectionTimeout int64
}

func New() (*Config, error) {
	env := os.Getenv(Env)
	if env == "" {
		env = ProdMode
	}
	if env != DevMode && env != ProdMode {
		return nil, fmt.Errorf("`%s` param allowed values: `%s`, `%s`", Env, DevMode, ProdMode)
	}

	listenHost := os.Getenv(ListenHttpHostEnvVarKey)
	if listenHost == "" {
		return nil, fmt.Errorf("required evironment variable `%s` isn't set", ListenHttpHostEnvVarKey)
	}

	checkoutDuration, err := strconv.ParseInt(os.Getenv(ConnectionsCheckoutDuration), 10, 64)
	if err != nil {
		checkoutDuration = 600
	}

	connectionTimeout, err := strconv.ParseInt(os.Getenv(WhatsAppConnectionTimeout), 10, 64)
	if err != nil {
		connectionTimeout = 20
	}

	filesRootPath := os.Getenv(FileSystemRootPoint)
	if filesRootPath == "" {
		return nil, fmt.Errorf("required evironment variable `%s` isn't set", FileSystemRootPoint)
	}

	redisHost := os.Getenv(RedisHost)
	if redisHost == "" {
		return nil, fmt.Errorf("required evironment variable `%s` isn't set", RedisHost)
	}

	webHookUrl := os.Getenv(WebHookUrl)
	if webHookUrl == "" {
		return nil, fmt.Errorf("required evironment variable `%s` isn't set", WebHookUrl)
	}
	if webHookUrl[len(webHookUrl)-1:] != "/" {
		return nil, fmt.Errorf("variable `%s` must contain trailing slash", WebHookUrl)
	}

	return &Config{
		ListenHttpHost:              listenHost,
		ConnectionTimeout:           connectionTimeout,
		FileSystemRootPath:          filesRootPath,
		WebHookUrl:                  webHookUrl,
		RedisHost:                   redisHost,
		Env:                         env,
		CertFilePath:                os.Getenv(CertFilePath),
		CertKeyPath:                 os.Getenv(CertKeyPath),
		SentryDSN:                   os.Getenv(SentryDSN),
		ConnectionsCheckoutDuration: checkoutDuration,
	}, nil
}
