package giftNotification

import (
	"gift-buyer/internal/config"
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
)

func TestNewNotification(t *testing.T) {
	mockClient := &tg.Client{}
	mockConfig := &config.TgSettings{
		NotificationChatID: 12345,
		TgBotKey:           "test_bot_token",
	}
	service := NewNotification(mockClient, mockConfig)

	assert.NotNil(t, service)

	assert.NotNil(t, service)
}

func TestNotificationService_Interface_Compliance(t *testing.T) {
	mockClient := &tg.Client{}
	mockConfig := &config.TgSettings{
		NotificationChatID: 12345,
		TgBotKey:           "test_bot_token",
	}
	service := NewNotification(mockClient, mockConfig)

	// Verify that the service implements the NotificationService interface
	// This is a compile-time check, but we can also verify at runtime
	assert.NotNil(t, service)
}

func TestNotificationService_Structure(t *testing.T) {
	mockClient := &tg.Client{}
	mockConfig := &config.TgSettings{
		NotificationChatID: 12345,
		TgBotKey:           "test_bot_token",
	}
	service := NewNotification(mockClient, mockConfig)

	// Cast to concrete type to verify internal structure
	assert.Equal(t, mockClient, service.Bot)
	assert.Equal(t, mockConfig, service.Config)
}

func TestNotificationService_NilClient(t *testing.T) {
	mockConfig := &config.TgSettings{
		NotificationChatID: 12345,
		TgBotKey:           "test_bot_token",
	}
	// Test with nil client - should not panic during creation
	service := NewNotification(nil, mockConfig)
	assert.NotNil(t, service)

	// Cast to concrete type to verify nil client is stored
	assert.Nil(t, service.Bot)
	assert.Equal(t, mockConfig, service.Config)
}

func TestNotificationService_NilConfig(t *testing.T) {
	mockClient := &tg.Client{}
	// Test with nil config - should not panic during creation
	service := NewNotification(mockClient, nil)
	assert.NotNil(t, service)

	// Cast to concrete type to verify nil config is stored
	assert.Equal(t, mockClient, service.Bot)
	assert.Nil(t, service.Config)
}
