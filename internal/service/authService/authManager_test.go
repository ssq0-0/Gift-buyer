package authService

import (
	"context"
	"testing"

	"gift-buyer/internal/config"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthManager(t *testing.T) {
	sessionManager := &MockSessionManager{}
	apiChecker := &MockApiChecker{}
	tgSettings := &config.TgSettings{
		AppId:    123456,
		ApiHash:  "test_hash",
		Phone:    "+1234567890",
		Password: "test_password",
	}

	authManager := NewAuthManager(sessionManager, apiChecker, tgSettings)

	assert.NotNil(t, authManager)
	assert.Equal(t, sessionManager, authManager.sessionManager)
	assert.Equal(t, apiChecker, authManager.apiChecker)
	assert.Equal(t, tgSettings, authManager.cfg)
}

func TestNewAuthManager_NilDependencies(t *testing.T) {
	// Test with nil session manager
	authManager := NewAuthManager(nil, nil, nil)
	assert.NotNil(t, authManager)

	assert.Nil(t, authManager.sessionManager)
	assert.Nil(t, authManager.apiChecker)
	assert.Nil(t, authManager.cfg)
}

func TestAuthManagerImpl_SetApiChecker(t *testing.T) {
	sessionManager := &MockSessionManager{}
	tgSettings := &config.TgSettings{
		AppId:   123456,
		ApiHash: "test_hash",
		Phone:   "+1234567890",
	}

	authManager := NewAuthManager(sessionManager, nil, tgSettings)

	// Initially should be nil
	assert.Nil(t, authManager.apiChecker)

	// Set api checker
	apiChecker := &MockApiChecker{}
	authManager.SetApiChecker(apiChecker)

	// Should now be set
	assert.Equal(t, apiChecker, authManager.apiChecker)
}

func TestAuthManagerImpl_SetMonitor(t *testing.T) {
	sessionManager := &MockSessionManager{}
	tgSettings := &config.TgSettings{
		AppId:   123456,
		ApiHash: "test_hash",
		Phone:   "+1234567890",
	}

	authManager := NewAuthManager(sessionManager, nil, tgSettings)

	// Initially should be nil
	assert.Nil(t, authManager.monitor)

	// Set monitor
	monitor := &MockGiftMonitor{}
	authManager.SetMonitor(monitor)

	// Should now be set
	assert.Equal(t, monitor, authManager.monitor)
}

func TestAuthManagerImpl_InitClient_NilSettings(t *testing.T) {
	authManager := NewAuthManager(nil, nil, nil)

	ctx := context.Background()

	// Должен вернуть ошибку без паники при nil настройках
	assert.NotPanics(t, func() {
		client, err := authManager.InitClient(ctx)
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestAuthManagerImpl_InitBotClient_NilSettings(t *testing.T) {
	authManager := NewAuthManager(nil, nil, nil)

	ctx := context.Background()
	client, err := authManager.InitBotClient(ctx)

	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestAuthManagerImpl_InitBotClient_NilSessionManager(t *testing.T) {
	tgSettings := &config.TgSettings{
		AppId:    123456,
		ApiHash:  "test_hash",
		Phone:    "+1234567890",
		TgBotKey: "test_bot_key",
	}

	authManager := NewAuthManager(nil, nil, tgSettings)

	ctx := context.Background()
	client, err := authManager.InitBotClient(ctx)

	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestAuthManagerImpl_RunApiChecker_NilApiChecker(t *testing.T) {
	sessionManager := &MockSessionManager{}
	tgSettings := &config.TgSettings{
		AppId:   123456,
		ApiHash: "test_hash",
		Phone:   "+1234567890",
	}

	authManager := NewAuthManager(sessionManager, nil, tgSettings)

	ctx := context.Background()

	// Should not panic with nil api checker
	assert.NotPanics(t, func() {
		authManager.RunApiChecker(ctx)
	})
}

func TestAuthManagerImpl_RunApiChecker_WithApiChecker(t *testing.T) {
	sessionManager := &MockSessionManager{}
	apiChecker := &MockApiChecker{}
	tgSettings := &config.TgSettings{
		AppId:   123456,
		ApiHash: "test_hash",
		Phone:   "+1234567890",
	}

	authManager := NewAuthManager(sessionManager, apiChecker, tgSettings)

	ctx := context.Background()

	// Should not panic with api checker
	assert.NotPanics(t, func() {
		authManager.RunApiChecker(ctx)
	})
}

func TestAuthManagerImpl_GetApi(t *testing.T) {
	authManager := NewAuthManager(nil, nil, nil)

	// Initially should be nil
	api := authManager.GetApi()
	assert.Nil(t, api)
}

func TestAuthManagerImpl_GetBotApi(t *testing.T) {
	authManager := NewAuthManager(nil, nil, nil)

	// Initially should be nil
	botApi := authManager.GetBotApi()
	assert.Nil(t, botApi)
}

func TestAuthManagerImpl_Stop(t *testing.T) {
	authManager := NewAuthManager(nil, nil, nil)

	// Should not panic
	assert.NotPanics(t, func() {
		authManager.Stop()
	})
}

// Mock implementations for testing

type MockSessionManager struct{}

func (m *MockSessionManager) InitUserAPI(client *telegram.Client, ctx context.Context) (*tg.Client, error) {
	return nil, assert.AnError
}

func (m *MockSessionManager) InitBotAPI(ctx context.Context) (*tg.Client, error) {
	return nil, assert.AnError
}

func (m *MockSessionManager) CreateDeviceConfig() telegram.DeviceConfig {
	return telegram.DeviceConfig{
		DeviceModel:    "Test Device",
		SystemVersion:  "Test OS",
		AppVersion:     "1.0.0",
		SystemLangCode: "en",
		LangCode:       "en",
	}
}

type MockApiChecker struct{}

func (m *MockApiChecker) Run(ctx context.Context) error {
	return nil
}

func (m *MockApiChecker) Stop() {
	// Mock implementation
}

type MockGiftMonitor struct{}

func (m *MockGiftMonitor) Pause() {}

func (m *MockGiftMonitor) Resume() {}

func (m *MockGiftMonitor) IsPaused() bool {
	return false
}
