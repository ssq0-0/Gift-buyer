// Package config provides application configuration management.
// It handles loading, parsing and validating application configuration from JSON files.
package config

import (
	"encoding/json"
	"gift-buyer/pkg/errors"
	"gift-buyer/pkg/logger"
	"os"
	"strconv"
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

	// tgSecrets, err := loadENVTgSetting()
	// if err != nil {
	// 	return nil, err
	// }

	appConfig := &AppConfig{}
	if err := json.Unmarshal(data, appConfig); err != nil {
		logger.GlobalLogger.Errorf("Failed to unmarshal config: %v", err)
		return nil, errors.Wrap(errors.ErrConfigParse, err.Error())
	}
	// appConfig.SoftConfig.TgSettings = tgSecrets
	return appConfig, nil
}

func loadENVTgSetting() (TgSettings, error) {
	settings := TgSettings{}
	if err := validateParams(
		os.Getenv("TG_APP_ID"),
		os.Getenv("TG_API_HASH"),
		os.Getenv("TG_PHONE"),
	); err != nil {
		return TgSettings{}, err
	}

	appId, err := strconv.Atoi(os.Getenv("TG_APP_ID"))
	if err != nil {
		return TgSettings{}, errors.Wrap(errors.ErrConfigParse, "invalid TG_APP_ID")
	}

	chatId, err := strconv.ParseInt(os.Getenv("TG_NOTIFICATION_CHAT_ID"), 10, 64)
	if err != nil && os.Getenv("TG_NOTIFICATION_CHAT_ID") != "" {
		return TgSettings{}, errors.Wrap(errors.ErrConfigParse, "invalid TG_NOTIFICATION_CHAT_ID")
	}

	settings.AppId = appId
	settings.ApiHash = os.Getenv("TG_API_HASH")
	settings.Phone = os.Getenv("TG_PHONE")
	settings.Password = os.Getenv("TG_PASSWORD")
	settings.TgBotKey = os.Getenv("TG_BOT_KEY")
	settings.NotificationChatID = chatId
	settings.DeviceModel = os.Getenv("DEVICE_MODEL")
	settings.SystemVersion = os.Getenv("SYSTEM_VERSION")
	settings.AppVersion = os.Getenv("APP_VERSION")
	settings.SystemLangCode = os.Getenv("SYSTEM_LANG_CODE")
	settings.LangCode = os.Getenv("LANG_CODE")
	settings.LangPack = os.Getenv("LANG_PACK")

	return settings, nil
}

func validateParams(value ...string) error {
	for _, v := range value {
		if v == "" {
			return errors.Wrap(errors.ErrConfigParse, "value is not set")
		}
	}
	return nil
}
