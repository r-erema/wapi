package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
			return err
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
			fmt.Sprintf("Invalid `%s` env variable", Env),
			map[string]string{Env: "invalid_val"},
			[]string{},
		},
		{
			fmt.Sprintf("Empty `%s` env variable", ListenHTTPHost),
			map[string]string{},
			[]string{ListenHTTPHost},
		},
		{
			fmt.Sprintf("Empty `%s` env variable", FileSystemRootPoint),
			map[string]string{},
			[]string{FileSystemRootPoint},
		},
		{
			fmt.Sprintf("Empty `%s` env variable", RedisHost),
			map[string]string{},
			[]string{RedisHost},
		},
		{
			fmt.Sprintf("Empty `%s` env variable", WebHookURL),
			map[string]string{},
			[]string{WebHookURL},
		},
		{
			fmt.Sprintf("Var `%s` must contain triling slash", WebHookURL),
			map[string]string{WebHookURL: "/wh"},
			[]string{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := setEnvs(tt.envVars, tt.excludedEnvVars)
			if err != nil {
				t.Error(err)
			}

			config, err := New()

			assert.Nil(t, config)
			assert.NotNil(t, err)
		})
	}
}

func TestDefaultEnvParam(t *testing.T) {
	var conf *Config
	err := setEnvs(map[string]string{}, []string{Env})
	if err != nil {
		t.Error(err)
	}
	conf, err = New()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, ProdMode, conf.Env)
}

func TestDefaultCheckoutDurationParam(t *testing.T) {
	var conf *Config
	err := setEnvs(map[string]string{}, []string{ConnectionsCheckoutDuration})
	if err != nil {
		t.Error(err)
	}
	conf, err = New()
	if err != nil {
		t.Error(err)
	}
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
