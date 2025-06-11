// Package config provides application configuration structures and utilities
// for the Gift Buyer application. It defines the configuration schema
// for Telegram settings, gift criteria, and operational parameters.
package config

// AppConfig represents the main application configuration structure.
// It contains logger settings and software-specific configuration.
type AppConfig struct {
	// LoggerLevel specifies the logging level (debug, info, warn, error, fatal, panic)
	LoggerLevel string `json:"logger_level"`

	// SoftConfig contains the core application configuration
	SoftConfig SoftConfig `json:"soft_config"`
}

// SoftConfig contains the core operational configuration for the gift buying system.
// It includes Telegram settings, purchase criteria, and operational limits.
type SoftConfig struct {
	// TgSettings contains Telegram API and bot configuration
	TgSettings TgSettings `json:"tg_settings"`

	// TotalStarCap is the maximum total stars that can be spent across all gifts
	TotalStarCap int64 `json:"total_star_cap"`

	// Criterias defines the list of criteria for gift validation
	Criterias []Criterias `json:"criterias"`

	// Receiver specifies the target recipient for purchased gifts
	Receiver ReceiverParams `json:"receiver"`

	// Ticker is the monitoring interval in seconds
	Ticker float64 `json:"ticker"`

	// TestMode enables test mode which bypasses certain validations
	TestMode bool `json:"test_mode"`

	// MaxBuyCount is the maximum number of gifts that can be purchased
	MaxBuyCount int64 `json:"max_buy_count"`

	// LimitedStatus is the status of the limited gifts
	LimitedStatus bool `json:"limited_status"`
}

// TgSettings contains all Telegram-related configuration parameters.
// This includes API credentials, bot settings, and notification preferences.
type TgSettings struct {
	// AppId is the Telegram application ID obtained from my.telegram.org
	AppId int `json:"app_id"`

	// ApiHash is the Telegram API hash obtained from my.telegram.org
	ApiHash string `json:"api_hash"`

	// Phone is the phone number associated with the Telegram account
	Phone string `json:"phone"`

	// Password is the 2FA password for the Telegram account (if enabled)
	Password string `json:"password"`

	// TgBotKey is the bot token for sending notifications
	TgBotKey string `json:"tg_bot_key"`

	// NotificationChatID is the chat ID where notifications will be sent
	NotificationChatID int64 `json:"notification_chat_id"`
}

// Criterias defines the validation criteria for gift purchases.
// Multiple criteria can be defined, and gifts matching any criteria will be considered eligible.
type Criterias struct {
	// MinPrice is the minimum price in stars for eligible gifts
	MinPrice int64 `json:"min_price"`

	// MaxPrice is the maximum price in stars for eligible gifts
	MaxPrice int64 `json:"max_price"`

	// TotalSupply is the minimum total supply required for limited gifts
	TotalSupply int64 `json:"total_supply"`

	// Count is the number of gifts to purchase when this criteria matches
	Count int64 `json:"count"`
}

// ReceiverParams specifies the recipient configuration for purchased gifts.
type ReceiverParams struct {
	// Type specifies the receiver type (1 for user, 2 for channel)
	Type []int `json:"type"`

	// ReceiverID is the Telegram ID of the gift recipient
	ReceiverID []int `json:"receiver_id"`
}
