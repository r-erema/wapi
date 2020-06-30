package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	// Service HTTP host.
	ListenHTTPHost = "WAPI_INTERNAL_HOST"
	// Timeout of connection with WhatsApp service.
	WhatsAppConnectionTimeout = "WAPI_WHATSAPP_CONNECTION_TIMEOUT"
	// Directory of storing static files.
	FileSystemRootPoint = "WAPI_FILE_SYSTEM_ROOT_POINT_FULL_PATH"
	// Webhook url.
	WebHookURL = "WAPI_GETTING_MESSAGES_WEBHOOK"
	// Redis host.
	RedisHost = "WAPI_REDIS_HOST"
	// Wapi environment: dev or prod.
	Env = "WAPI_ENV"
	// Path to certificate file.
	CertFilePath = "WAPI_CERT_FILE_PATH"
	// Path to certificate key file.
	CertKeyPath = "WAPI_CERT_KEY_PATH"
	// Sentry connection string.
	SentryDSN = "WAPI_SENTRY_DSN"
	// Connections checkout durations in seconds.
	ConnectionsCheckoutDuration = "WAPI_CONNECTIONS_CHECKOUT_DURATION_SECS"

	// Development mode value of wapi environment.
	DevMode = "dev"
	// Production mode value of wapi environment.
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

// New creates common config contains all application parameters.
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
