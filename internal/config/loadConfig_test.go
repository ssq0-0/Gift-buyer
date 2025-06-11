package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	pkgErrors "gift-buyer/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Success(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	config := &AppConfig{
		LoggerLevel: "info",
		SoftConfig: SoftConfig{
			TgSettings: TgSettings{
				AppId:    123456,
				ApiHash:  "test_api_hash",
				Phone:    "+1234567890",
				Password: "test_password",
			},
			Criterias: []Criterias{
				{
					MinPrice:    100,
					MaxPrice:    1000,
					TotalSupply: 50,
				},
			},
			Receiver: ReceiverParams{
				Type:       []int{1},
				ReceiverID: []int{987654321},
			},
			Ticker: 30.0,
		},
	}

	// Write config to file
	data, err := json.Marshal(config)
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	// Test loading config
	loadedConfig, err := LoadConfig(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, loadedConfig)
	assert.Equal(t, config.LoggerLevel, loadedConfig.LoggerLevel)
	assert.Equal(t, config.SoftConfig.TgSettings.AppId, loadedConfig.SoftConfig.TgSettings.AppId)
	assert.Equal(t, config.SoftConfig.TgSettings.ApiHash, loadedConfig.SoftConfig.TgSettings.ApiHash)
	assert.Equal(t, config.SoftConfig.TgSettings.Phone, loadedConfig.SoftConfig.TgSettings.Phone)
	assert.Equal(t, config.SoftConfig.TgSettings.Password, loadedConfig.SoftConfig.TgSettings.Password)
	assert.Equal(t, len(config.SoftConfig.Criterias), len(loadedConfig.SoftConfig.Criterias))
	assert.Equal(t, config.SoftConfig.Criterias[0].MinPrice, loadedConfig.SoftConfig.Criterias[0].MinPrice)
	assert.Equal(t, config.SoftConfig.Criterias[0].MaxPrice, loadedConfig.SoftConfig.Criterias[0].MaxPrice)
	assert.Equal(t, config.SoftConfig.Criterias[0].TotalSupply, loadedConfig.SoftConfig.Criterias[0].TotalSupply)
	assert.Equal(t, config.SoftConfig.Receiver, loadedConfig.SoftConfig.Receiver)
	assert.Equal(t, config.SoftConfig.Ticker, loadedConfig.SoftConfig.Ticker)
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	nonExistentPath := "/path/that/does/not/exist/config.json"

	config, err := LoadConfig(nonExistentPath)
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.True(t, errors.Is(err, pkgErrors.ErrConfigRead))
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid_config.json")

	// Write invalid JSON to file
	invalidJSON := `{
		"logger_level": "info",
		"soft_config": {
			"tg_settings": {
				"app_id": 123456,
				"api_hash": "test_hash"
				// missing comma and closing braces
	`
	err := os.WriteFile(configPath, []byte(invalidJSON), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configPath)
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.True(t, errors.Is(err, pkgErrors.ErrConfigParse))
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "empty_config.json")

	// Write empty file
	err := os.WriteFile(configPath, []byte(""), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configPath)
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.True(t, errors.Is(err, pkgErrors.ErrConfigParse))
}

func TestLoadConfig_PartialConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "partial_config.json")

	// Create partial config with only some fields
	partialConfig := map[string]interface{}{
		"logger_level": "debug",
		"soft_config": map[string]interface{}{
			"tg_settings": map[string]interface{}{
				"app_id":   123456,
				"api_hash": "test_hash",
			},
			"ticker": 60.0,
		},
	}

	data, err := json.Marshal(partialConfig)
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "debug", config.LoggerLevel)
	assert.Equal(t, 123456, config.SoftConfig.TgSettings.AppId)
	assert.Equal(t, "test_hash", config.SoftConfig.TgSettings.ApiHash)
	assert.Equal(t, 60.0, config.SoftConfig.Ticker)
	// Other fields should have zero values
	assert.Equal(t, "", config.SoftConfig.TgSettings.Phone)
	assert.Equal(t, "", config.SoftConfig.TgSettings.Password)
	assert.Equal(t, ReceiverParams{Type: []int(nil), ReceiverID: []int(nil)}, config.SoftConfig.Receiver)
	assert.Empty(t, config.SoftConfig.Criterias)
}

func TestLoadConfig_MultipleCriterias(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "multi_criteria_config.json")

	config := &AppConfig{
		LoggerLevel: "warn",
		SoftConfig: SoftConfig{
			TgSettings: TgSettings{
				AppId:    789012,
				ApiHash:  "multi_test_hash",
				Phone:    "+9876543210",
				Password: "multi_password",
			},
			Criterias: []Criterias{
				{
					MinPrice:    50,
					MaxPrice:    500,
					TotalSupply: 25,
				},
				{
					MinPrice:    1000,
					MaxPrice:    5000,
					TotalSupply: 100,
				},
				{
					MinPrice:    10000,
					MaxPrice:    50000,
					TotalSupply: 10,
				},
			},
			Receiver: ReceiverParams{
				Type:       []int{1},
				ReceiverID: []int{111222333},
			},
			Ticker: 15.0,
		},
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	loadedConfig, err := LoadConfig(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, loadedConfig)
	assert.Equal(t, 3, len(loadedConfig.SoftConfig.Criterias))

	// Check first criteria
	assert.Equal(t, int64(50), loadedConfig.SoftConfig.Criterias[0].MinPrice)
	assert.Equal(t, int64(500), loadedConfig.SoftConfig.Criterias[0].MaxPrice)
	assert.Equal(t, int64(25), loadedConfig.SoftConfig.Criterias[0].TotalSupply)

	// Check second criteria
	assert.Equal(t, int64(1000), loadedConfig.SoftConfig.Criterias[1].MinPrice)
	assert.Equal(t, int64(5000), loadedConfig.SoftConfig.Criterias[1].MaxPrice)
	assert.Equal(t, int64(100), loadedConfig.SoftConfig.Criterias[1].TotalSupply)

	// Check third criteria
	assert.Equal(t, int64(10000), loadedConfig.SoftConfig.Criterias[2].MinPrice)
	assert.Equal(t, int64(50000), loadedConfig.SoftConfig.Criterias[2].MaxPrice)
	assert.Equal(t, int64(10), loadedConfig.SoftConfig.Criterias[2].TotalSupply)
}

func TestLoadConfig_ZeroValues(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "zero_config.json")

	config := &AppConfig{
		LoggerLevel: "",
		SoftConfig: SoftConfig{
			TgSettings: TgSettings{
				AppId:    0,
				ApiHash:  "",
				Phone:    "",
				Password: "",
			},
			Criterias: []Criterias{
				{
					MinPrice:    0,
					MaxPrice:    0,
					TotalSupply: 0,
				},
			},
			Receiver: ReceiverParams{
				Type:       []int{0},
				ReceiverID: []int{0},
			},
			Ticker: 0.0,
		},
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	loadedConfig, err := LoadConfig(configPath)
	assert.NoError(t, err)
	assert.NotNil(t, loadedConfig)
	assert.Equal(t, "", loadedConfig.LoggerLevel)
	assert.Equal(t, 0, loadedConfig.SoftConfig.TgSettings.AppId)
	assert.Equal(t, "", loadedConfig.SoftConfig.TgSettings.ApiHash)
	assert.Equal(t, "", loadedConfig.SoftConfig.TgSettings.Phone)
	assert.Equal(t, "", loadedConfig.SoftConfig.TgSettings.Password)
	assert.Equal(t, 1, len(loadedConfig.SoftConfig.Criterias))
	assert.Equal(t, int64(0), loadedConfig.SoftConfig.Criterias[0].MinPrice)
	assert.Equal(t, int64(0), loadedConfig.SoftConfig.Criterias[0].MaxPrice)
	assert.Equal(t, int64(0), loadedConfig.SoftConfig.Criterias[0].TotalSupply)
	assert.Equal(t, ReceiverParams{Type: []int{0}, ReceiverID: []int{0}}, loadedConfig.SoftConfig.Receiver)
	assert.Equal(t, 0.0, loadedConfig.SoftConfig.Ticker)
}

func TestLoadConfig_PermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "no_permission_config.json")

	// Create file and remove read permissions
	err := os.WriteFile(configPath, []byte(`{"logger_level": "info"}`), 0644)
	require.NoError(t, err)
	err = os.Chmod(configPath, 0000)
	require.NoError(t, err)

	// Restore permissions after test
	defer func() {
		os.Chmod(configPath, 0644)
	}()

	config, err := LoadConfig(configPath)
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.True(t, errors.Is(err, pkgErrors.ErrConfigRead))
}
