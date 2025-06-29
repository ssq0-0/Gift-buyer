// Package config provides application configuration management.
// It handles loading, parsing and validating application configuration from JSON files.
package config

import (
	"encoding/json"
	"gift-buyer/pkg/errors"
	"gift-buyer/pkg/logger"
	"os"
)

// LoadConfig loads and parses the application configuration from the specified JSON file.
// It reads the configuration file, unmarshals the JSON content, and returns the parsed
// configuration structure.
//
// The configuration file should be in JSON format and contain all required settings
// including Telegram credentials, gift criteria, and operational parameters.
//
// Parameters:
//   - path: filesystem path to the configuration JSON file
//
// Returns:
//   - *AppConfig: parsed configuration structure containing all application settings
//   - error: configuration loading or parsing error, wrapped with context information
//
// Example usage:
//
//	cfg, err := LoadConfig("config/app.json")
//	if err != nil {
//	    log.Fatalf("Failed to load config: %v", err)
//	}
//
// Possible errors:
//   - ErrConfigRead: when the configuration file cannot be read
//   - ErrConfigParse: when the JSON content cannot be parsed
func LoadConfig(path string) (*AppConfig, error) {
	logger.GlobalLogger.Debugf("Loading config from: %s", path)

	data, err := os.ReadFile(path)
	if err != nil {
		logger.GlobalLogger.Errorf("Failed to read config file: %v", err)
		return nil, errors.Wrap(errors.ErrConfigRead, err.Error())
	}

	appConfig := &AppConfig{}
	if err := json.Unmarshal(data, appConfig); err != nil {
		logger.GlobalLogger.Errorf("Failed to unmarshal config: %v", err)
		return nil, errors.Wrap(errors.ErrConfigParse, err.Error())
	}
	return appConfig, nil
}
