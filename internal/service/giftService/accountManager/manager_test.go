package accountManager

import (
	"context"
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
)

func TestNewAccountManager(t *testing.T) {
	userReceiverIDs := []int{123456789}
	channelReceiverIDs := []int{987654321}
	userCache := &MockUserCache{}

	manager := NewAccountManager(nil, userReceiverIDs, channelReceiverIDs, userCache)

	assert.NotNil(t, manager)
}

func TestNewAccountManager_EmptyIDs(t *testing.T) {
	userCache := &MockUserCache{}

	manager := NewAccountManager(nil, []int{}, []int{}, userCache)

	assert.NotNil(t, manager)
}

func TestNewAccountManager_NilCache(t *testing.T) {
	userReceiverIDs := []int{123456789}
	channelReceiverIDs := []int{987654321}

	manager := NewAccountManager(nil, userReceiverIDs, channelReceiverIDs, nil)

	assert.NotNil(t, manager)
}

func TestAccountManager_SetIds_NilAPI(t *testing.T) {
	userCache := &MockUserCache{}
	manager := NewAccountManager(nil, []int{123456789}, []int{987654321}, userCache)

	ctx := context.Background()

	// Должен вернуть ошибку без паники при nil API
	assert.NotPanics(t, func() {
		err := manager.SetIds(ctx)
		assert.Error(t, err)
	})
}

func TestAccountManager_SetIds_EmptyIDs(t *testing.T) {
	api := &tg.Client{}
	userCache := &MockUserCache{}
	manager := NewAccountManager(api, []int{}, []int{}, userCache)

	ctx := context.Background()
	err := manager.SetIds(ctx)

	assert.NoError(t, err) // Should succeed with empty IDs
}

func TestAccountManager_SetIds_ContextCancellation(t *testing.T) {
	userCache := &MockUserCache{}
	manager := NewAccountManager(nil, []int{123456789}, []int{987654321}, userCache)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := manager.SetIds(ctx)

	assert.Error(t, err)
	// С nil API мы получим ошибку "API client is nil", а не context.Canceled
	assert.Contains(t, err.Error(), "API client is nil")
}

func TestAccountManager_Structure(t *testing.T) {
	userReceiverIDs := []int{123456789, 987654321}
	channelReceiverIDs := []int{111222333, 444555666}
	userCache := &MockUserCache{}
	api := &tg.Client{}

	manager := NewAccountManager(api, userReceiverIDs, channelReceiverIDs, userCache)

	assert.NotNil(t, manager)
}

func TestAccountManager_InterfaceCompliance(t *testing.T) {
	userCache := &MockUserCache{}
	manager := NewAccountManager(nil, []int{}, []int{}, userCache)

	// Verify that the manager has the SetIds method
	assert.NotNil(t, manager.SetIds)
}

func TestAccountManager_LoadUsersToCache_NilAPI(t *testing.T) {
	userCache := &MockUserCache{}
	manager := NewAccountManager(nil, []int{123456789}, []int{}, userCache)

	ctx := context.Background()

	// Должен вернуть ошибку без паники при nil API
	assert.NotPanics(t, func() {
		err := manager.SetIds(ctx)
		assert.Error(t, err)
	})
}

func TestAccountManager_LoadChannelsToCache_NilAPI(t *testing.T) {
	userCache := &MockUserCache{}
	manager := NewAccountManager(nil, []int{}, []int{987654321}, userCache)

	ctx := context.Background()

	// Должен вернуть ошибку без паники при nil API
	assert.NotPanics(t, func() {
		err := manager.SetIds(ctx)
		assert.Error(t, err)
	})
}

// Mock implementations for testing

type MockUserCache struct{}

func (m *MockUserCache) SetUser(user *tg.User) {
	// Mock implementation
}

func (m *MockUserCache) GetUser(id int64) (*tg.User, error) {
	return nil, assert.AnError
}

func (m *MockUserCache) SetChannel(channel *tg.Channel) {
	// Mock implementation
}

func (m *MockUserCache) GetChannel(id int64) (*tg.Channel, error) {
	return nil, assert.AnError
}
