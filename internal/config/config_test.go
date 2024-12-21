package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func createTestConfigFile(content string) (string, error) {
	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func TestLoadConfig_Success(t *testing.T) {
	content := `
server:
  host: localhost
  port: 8080
log:
  level: info
  path: /log/app.log
`
	tempFile, err := createTestConfigFile(content)
	if err != nil {
		t.Fatal("Failed to create temp config file:", err)
	}
	defer os.Remove(tempFile)

	viper.SetConfigFile(tempFile)
	config, err := LoadConfig(tempFile)

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "localhost", config.Server.Host)
	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, "info", config.Log.Level)
	assert.Equal(t, "/log/app.log", config.Log.Path)
}

func TestLoadConfig_ReadInConfigError(t *testing.T) {
	tempFile, err := createTestConfigFile("invalid config")
	if err != nil {
		t.Fatal("Failed to create temp config file:", err)
	}
	defer os.Remove(tempFile)

	viper.SetConfigFile(tempFile)
	config, err := LoadConfig(tempFile)

	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestLoadConfig_UnmarshalError(t *testing.T) {
	content := `
server:
  host: localhost
  port: "invalid"
log:
  level: info
  path: /log/app.log
`
	tempFile, err := createTestConfigFile(content)
	if err != nil {
		t.Fatal("Failed to create temp config file:", err)
	}
	defer os.Remove(tempFile)

	viper.SetConfigFile(tempFile)
	config, err := LoadConfig(tempFile)

	assert.Error(t, err)
	assert.Nil(t, config)
}
