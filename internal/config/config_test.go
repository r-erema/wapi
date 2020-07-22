package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var defaultEnvs = map[string]string{
	ListenHTTPHost:              "localhost",
	WhatsAppConnectionTimeout:   "10",
	FileSystemRootPoint:         "/tmp",
	WebHookURL:                  "/wh/",
	RedisHost:                   "localhost:6379",
	Env:                         "",
	CertFilePath:                "/tmp/cert.crt",
	CertKeyPath:                 "/tmp/cert.key",
	SentryDSN:                   "dsn@sentry.io/test",
	ConnectionsCheckoutDuration: "60",
}

func setEnvs(customEnvs map[string]string, excludedEnvs []string) (err error) {
	os.Clearenv()
	envsToSet := make(map[string]string)
	for env, val := range defaultEnvs {
		envsToSet[env] = val
	}
	for env, val := range customEnvs {
		if _, ok := envsToSet[env]; ok {
			envsToSet[env] = val
		}
	}
	for _, env := range excludedEnvs {
		delete(envsToSet, env)
	}
	for env, val := range envsToSet {
		err = os.Setenv(env, val)
		if err != nil {
			return errors.Wrap(err, "setting env var failed")
		}
	}
	return nil
}

func TestNotValidEnvVars(t *testing.T) {
	tests := []struct {
		name            string
		envVars         map[string]string
		excludedEnvVars []string
	}{
		{
			name:            fmt.Sprintf("Invalid `%s` env variable", Env),
			envVars:         map[string]string{Env: "invalid_val"},
			excludedEnvVars: []string{},
		},
		{
			name:            fmt.Sprintf("Empty `%s` env variable", ListenHTTPHost),
			envVars:         map[string]string{},
			excludedEnvVars: []string{ListenHTTPHost},
		},
		{
			name:            fmt.Sprintf("Empty `%s` env variable", FileSystemRootPoint),
			envVars:         map[string]string{},
			excludedEnvVars: []string{FileSystemRootPoint},
		},
		{
			name:            fmt.Sprintf("Empty `%s` env variable", RedisHost),
			envVars:         map[string]string{},
			excludedEnvVars: []string{RedisHost},
		},
		{
			name:            fmt.Sprintf("Empty `%s` env variable", WebHookURL),
			envVars:         map[string]string{},
			excludedEnvVars: []string{WebHookURL},
		},
		{
			name:            fmt.Sprintf("Var `%s` must contain triling slash", WebHookURL),
			envVars:         map[string]string{WebHookURL: "/wh"},
			excludedEnvVars: []string{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := setEnvs(tt.envVars, tt.excludedEnvVars)
			require.Nil(t, err)

			config, err := New()

			assert.Nil(t, config)
			assert.NotNil(t, err)
		})
	}
}

func TestDefaultEnvParam(t *testing.T) {
	var conf *Config

	err := setEnvs(map[string]string{}, []string{Env})
	require.Nil(t, err)

	conf, err = New()
	require.Nil(t, err)

	assert.Equal(t, ProdMode, conf.Env)
}

func TestDefaultCheckoutDurationParam(t *testing.T) {
	var conf *Config

	err := setEnvs(map[string]string{}, []string{ConnectionsCheckoutDuration})
	require.Nil(t, err)

	conf, err = New()
	require.Nil(t, err)

	assert.Equal(t, DefaultConnectionsCheckoutDuration, conf.ConnectionsCheckoutDuration)
}

func TestDefaultConnectionTimeoutParam(t *testing.T) {
	var conf *Config
	err := setEnvs(map[string]string{}, []string{WhatsAppConnectionTimeout})
	if err != nil {
		t.Error(err)
	}
	conf, err = New()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, DefaultConnectionTimeout, conf.ConnectionTimeout)
}
