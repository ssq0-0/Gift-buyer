package giftNotification

import (
	"context"
	"gift-buyer/internal/config"
	"gift-buyer/internal/service/giftService/giftInterfaces"
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

	_, ok := service.(giftInterfaces.NotificationService)
	assert.True(t, ok)
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
	_, ok := service.(interface {
		SendNewGiftNotification(ctx context.Context, gift *tg.StarGift) error
		SendBuyStatus(ctx context.Context, status string, err error) error
	})
	assert.True(t, ok, "NotificationServiceImpl should implement the NotificationService interface")
}

func TestNotificationService_Structure(t *testing.T) {
	mockClient := &tg.Client{}
	mockConfig := &config.TgSettings{
		NotificationChatID: 12345,
		TgBotKey:           "test_bot_token",
	}
	service := NewNotification(mockClient, mockConfig)

	// Cast to concrete type to verify internal structure
	ns, ok := service.(*NotificationServiceImpl)
	assert.True(t, ok)
	assert.Equal(t, mockClient, ns.Bot)
	assert.Equal(t, mockConfig, ns.Config)
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
	ns, ok := service.(*NotificationServiceImpl)
	assert.True(t, ok)
	assert.Nil(t, ns.Bot)
	assert.Equal(t, mockConfig, ns.Config)
}

func TestNotificationService_NilConfig(t *testing.T) {
	mockClient := &tg.Client{}
	// Test with nil config - should not panic during creation
	service := NewNotification(mockClient, nil)
	assert.NotNil(t, service)

	// Cast to concrete type to verify nil config is stored
	ns, ok := service.(*NotificationServiceImpl)
	assert.True(t, ok)
	assert.Equal(t, mockClient, ns.Bot)
	assert.Nil(t, ns.Config)
}
