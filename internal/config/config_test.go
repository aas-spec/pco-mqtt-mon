package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const testCfgJSON = `{
	"Common": {"LogFiles": 7},
	"MQTT": {"Host": "localhost", "Port": 1883, "User": "user", "Password": "pass"},
	"Services": []
}`

func TestLoadConfig(t *testing.T) {
	f, err := os.CreateTemp("", "pco-test-*.cfg")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	_, err = f.WriteString(testCfgJSON)
	require.NoError(t, err)
	f.Close()

	var cfg AppConfig
	cfg.LoadConfig(f.Name())

	require.Equal(t, int64(7), cfg.Common.LogFiles.Int64)
	require.Equal(t, "localhost", cfg.MQTT.Host)
	require.Equal(t, 1883, cfg.MQTT.Port)
	require.Equal(t, "user", cfg.MQTT.User)
	require.Equal(t, "pass", cfg.MQTT.Password)
	require.Empty(t, cfg.Services)
}

func TestLoadConfigMissingFile(t *testing.T) {
	var cfg AppConfig
	require.Panics(t, func() {
		cfg.LoadConfig("/nonexistent/pco-test.cfg")
	})
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	f, err := os.CreateTemp("", "pco-test-*.cfg")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	_, err = f.WriteString("not valid json")
	require.NoError(t, err)
	f.Close()

	var cfg AppConfig
	require.Panics(t, func() {
		cfg.LoadConfig(f.Name())
	})
}
