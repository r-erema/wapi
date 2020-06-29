package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	ListenHTTPHost              = "WAPI_INTERNAL_HOST"
	WhatsAppConnectionTimeout   = "WAPI_WHATSAPP_CONNECTION_TIMEOUT"
	FileSystemRootPoint         = "WAPI_FILE_SYSTEM_ROOT_POINT_FULL_PATH"
	WebHookURL                  = "WAPI_GETTING_MESSAGES_WEBHOOK"
	RedisHost                   = "WAPI_REDIS_HOST"
	Env                         = "WAPI_ENV"
	CertFilePath                = "WAPI_CERT_FILE_PATH"
	CertKeyPath                 = "WAPI_CERT_KEY_PATH"
	SentryDSN                   = "WAPI_SENTRY_DSN"
	ConnectionsCheckoutDuration = "WAPI_CONNECTIONS_CHECKOUT_DURATION_SECS"

	DevMode  = "dev"
	ProdMode = "prod"

	DefaultConnectionsCheckoutDuration = 60
	DefaultConnectionTimeout           = 20
)

// Config stores all application parameters.
type Config struct {
	ListenHTTPHost,
	FileSystemRootPath,
	WebHookURL,
	RedisHost,
	Env,
	CertFilePath,
	HTTPStaticFiles,
	SentryDSN,
	CertKeyPath string
	ConnectionsCheckoutDuration,
	ConnectionTimeout int
}

// Creates common config contains all application parameters.
func New() (*Config, error) {
	env := os.Getenv(Env)
	if env == "" {
		env = ProdMode
	}
	if env != DevMode && env != ProdMode {
		return nil, fmt.Errorf("`%s` param allowed values: `%s`, `%s`", Env, DevMode, ProdMode)
	}

	listenHost := os.Getenv(ListenHTTPHost)
	if listenHost == "" {
		return nil, fmt.Errorf("required evironment variable `%s` isn't set", ListenHTTPHost)
	}

	checkoutDuration, err := strconv.ParseInt(os.Getenv(ConnectionsCheckoutDuration), 10, 64)
	if err != nil {
		checkoutDuration = DefaultConnectionsCheckoutDuration
	}

	connectionTimeout, err := strconv.ParseInt(os.Getenv(WhatsAppConnectionTimeout), 10, 64)
	if err != nil {
		connectionTimeout = DefaultConnectionTimeout
	}

	filesRootPath := os.Getenv(FileSystemRootPoint)
	if filesRootPath == "" {
		return nil, fmt.Errorf("required evironment variable `%s` isn't set", FileSystemRootPoint)
	}

	redisHost := os.Getenv(RedisHost)
	if redisHost == "" {
		return nil, fmt.Errorf("required evironment variable `%s` isn't set", RedisHost)
	}

	webHookURL := os.Getenv(WebHookURL)
	if webHookURL == "" {
		return nil, fmt.Errorf("required evironment variable `%s` isn't set", WebHookURL)
	}
	if webHookURL[len(webHookURL)-1:] != "/" {
		return nil, fmt.Errorf("variable `%s` must contain trailing slash", WebHookURL)
	}

	return &Config{
		ListenHTTPHost:              listenHost,
		ConnectionTimeout:           int(connectionTimeout),
		FileSystemRootPath:          filesRootPath,
		WebHookURL:                  webHookURL,
		RedisHost:                   redisHost,
		Env:                         env,
		CertFilePath:                os.Getenv(CertFilePath),
		CertKeyPath:                 os.Getenv(CertKeyPath),
		SentryDSN:                   os.Getenv(SentryDSN),
		ConnectionsCheckoutDuration: int(checkoutDuration),
	}, nil
}
