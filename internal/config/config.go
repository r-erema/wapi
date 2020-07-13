package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	ListenHTTPHost = "WAPI_INTERNAL_HOST" // Service HTTP host aka API url.
	// WhatsAppConnectionTimeout represents timeout of establishing connection with WhatsApp service in seconds.
	WhatsAppConnectionTimeout   = "WAPI_WHATSAPP_CONNECTION_TIMEOUT"
	FileSystemRootPoint         = "WAPI_FILE_SYSTEM_ROOT_POINT_FULL_PATH"           // Path of storing static files.
	WebHookURL                  = "WAPI_GETTING_MESSAGES_WEBHOOK"                   // Base webhook url.
	RedisHost                   = "WAPI_REDIS_HOST"                                 // Redis host.
	Env                         = "WAPI_ENV"                                        // Wapi environment: dev or prod.
	CertFilePath                = "WAPI_CERT_FILE_PATH"                             // Path to certificate file.
	CertKeyPath                 = "WAPI_CERT_KEY_PATH"                              // Path to certificate key file.
	SentryDSN                   = "WAPI_SENTRY_DSN"                                 // Sentry connection string.
	ConnectionsCheckoutDuration = "WAPI_CONNECTIONS_CHECKOUT_DURATION_MILLISECONDS" // Connections checkout durations in seconds.

	DevMode  = "dev"  // Development mode value of wapi environment.
	ProdMode = "prod" // Production mode value of wapi environment.

	DefaultConnectionsCheckoutDuration = 60 // Default timeout of establishing connection with WhatsApp service in seconds.
	DefaultConnectionTimeout           = 20 // Default connections checkout durations in seconds.
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

	var listenHost string
	var ok bool
	if listenHost, ok = os.LookupEnv(ListenHTTPHost); !ok {
		return nil, fmt.Errorf("required evironment variable `%s` isn't set", ListenHTTPHost)
	}

	checkoutDuration, err := strconv.Atoi(os.Getenv(ConnectionsCheckoutDuration))
	if err != nil {
		checkoutDuration = DefaultConnectionsCheckoutDuration
	}

	connectionTimeout, err := strconv.Atoi(os.Getenv(WhatsAppConnectionTimeout))
	if err != nil {
		connectionTimeout = DefaultConnectionTimeout
	}

	var filesRootPath string
	if filesRootPath, ok = os.LookupEnv(FileSystemRootPoint); !ok {
		return nil, fmt.Errorf("required evironment variable `%s` isn't set", FileSystemRootPoint)
	}

	var redisHost string
	if redisHost, ok = os.LookupEnv(RedisHost); !ok {
		return nil, fmt.Errorf("required evironment variable `%s` isn't set", RedisHost)
	}

	var webHookURL string
	if webHookURL, ok = os.LookupEnv(WebHookURL); !ok {
		return nil, fmt.Errorf("required evironment variable `%s` isn't set", WebHookURL)
	}
	if webHookURL[len(webHookURL)-1:] != "/" {
		return nil, fmt.Errorf("variable `%s` must contain trailing slash", WebHookURL)
	}

	return &Config{
		ListenHTTPHost:              listenHost,
		ConnectionTimeout:           connectionTimeout,
		FileSystemRootPath:          filesRootPath,
		WebHookURL:                  webHookURL,
		RedisHost:                   redisHost,
		Env:                         env,
		CertFilePath:                os.Getenv(CertFilePath),
		CertKeyPath:                 os.Getenv(CertKeyPath),
		SentryDSN:                   os.Getenv(SentryDSN),
		ConnectionsCheckoutDuration: checkoutDuration,
	}, nil
}
