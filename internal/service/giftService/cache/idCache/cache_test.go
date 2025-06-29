package idCache

import (
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
)

func TestNewIDCache(t *testing.T) {
	cache := NewIDCache()

	assert.NotNil(t, cache)
	// Проверяем методы интерфейса
	assert.NotNil(t, cache.SetUser)
	assert.NotNil(t, cache.GetUser)
	assert.NotNil(t, cache.SetChannel)
	assert.NotNil(t, cache.GetChannel)
}

func TestIDCacheImpl_SetAndGetUser(t *testing.T) {
	cache := NewIDCache()

	user := &tg.User{
		ID:        123456789,
		FirstName: "Test",
		LastName:  "User",
		Username:  "testuser",
	}

	// Set user
	cache.SetUser(user)

	// Get user
	retrievedUser, err := cache.GetUser(123456789)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedUser)
	assert.Equal(t, user.ID, retrievedUser.ID)
	assert.Equal(t, user.FirstName, retrievedUser.FirstName)
	assert.Equal(t, user.LastName, retrievedUser.LastName)
	assert.Equal(t, user.Username, retrievedUser.Username)
}

func TestIDCacheImpl_GetNonExistentUser(t *testing.T) {
	cache := NewIDCache()

	user, err := cache.GetUser(999999999)
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "user not found")
}

func TestIDCacheImpl_SetAndGetChannel(t *testing.T) {
	cache := NewIDCache()

	channel := &tg.Channel{
		ID:       987654321,
		Title:    "Test Channel",
		Username: "testchannel",
	}

	// Set channel
	cache.SetChannel(channel)

	// Get channel
	retrievedChannel, err := cache.GetChannel(987654321)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedChannel)
	assert.Equal(t, channel.ID, retrievedChannel.ID)
	assert.Equal(t, channel.Title, retrievedChannel.Title)
	assert.Equal(t, channel.Username, retrievedChannel.Username)
}

func TestIDCacheImpl_GetNonExistentChannel(t *testing.T) {
	cache := NewIDCache()

	channel, err := cache.GetChannel(999999999)
	assert.Error(t, err)
	assert.Nil(t, channel)
	assert.Contains(t, err.Error(), "channel not found")
}

func TestIDCacheImpl_SetNilUser(t *testing.T) {
	cache := NewIDCache()

	// Should not panic with nil user
	assert.NotPanics(t, func() {
		cache.SetUser(nil)
	})
}

func TestIDCacheImpl_SetNilChannel(t *testing.T) {
	cache := NewIDCache()

	// Should not panic with nil channel
	assert.NotPanics(t, func() {
		cache.SetChannel(nil)
	})
}

func TestIDCacheImpl_OverwriteUser(t *testing.T) {
	cache := NewIDCache()

	user1 := &tg.User{
		ID:        123456789,
		FirstName: "First",
		LastName:  "User",
	}

	user2 := &tg.User{
		ID:        123456789,
		FirstName: "Second",
		LastName:  "User",
	}

	// Set first user
	cache.SetUser(user1)

	// Overwrite with second user
	cache.SetUser(user2)

	// Should get the second user
	retrievedUser, err := cache.GetUser(123456789)
	assert.NoError(t, err)
	assert.Equal(t, "Second", retrievedUser.FirstName)
}

func TestIDCacheImpl_OverwriteChannel(t *testing.T) {
	cache := NewIDCache()

	channel1 := &tg.Channel{
		ID:    987654321,
		Title: "First Channel",
	}

	channel2 := &tg.Channel{
		ID:    987654321,
		Title: "Second Channel",
	}

	// Set first channel
	cache.SetChannel(channel1)

	// Overwrite with second channel
	cache.SetChannel(channel2)

	// Should get the second channel
	retrievedChannel, err := cache.GetChannel(987654321)
	assert.NoError(t, err)
	assert.Equal(t, "Second Channel", retrievedChannel.Title)
}

func TestIDCacheImpl_MultipleUsers(t *testing.T) {
	cache := NewIDCache()

	users := []*tg.User{
		{ID: 111, FirstName: "User1"},
		{ID: 222, FirstName: "User2"},
		{ID: 333, FirstName: "User3"},
	}

	// Set multiple users
	for _, user := range users {
		cache.SetUser(user)
	}

	// Get all users
	for _, expectedUser := range users {
		retrievedUser, err := cache.GetUser(expectedUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.FirstName, retrievedUser.FirstName)
	}
}

func TestIDCacheImpl_MultipleChannels(t *testing.T) {
	cache := NewIDCache()

	channels := []*tg.Channel{
		{ID: 111, Title: "Channel1"},
		{ID: 222, Title: "Channel2"},
		{ID: 333, Title: "Channel3"},
	}

	// Set multiple channels
	for _, channel := range channels {
		cache.SetChannel(channel)
	}

	// Get all channels
	for _, expectedChannel := range channels {
		retrievedChannel, err := cache.GetChannel(expectedChannel.ID)
		assert.NoError(t, err)
		assert.Equal(t, expectedChannel.Title, retrievedChannel.Title)
	}
}

func TestIDCacheImpl_InterfaceCompliance(t *testing.T) {
	cache := NewIDCache()

	// Verify that the cache has all required methods
	assert.NotNil(t, cache.SetUser)
	assert.NotNil(t, cache.GetUser)
	assert.NotNil(t, cache.SetChannel)
	assert.NotNil(t, cache.GetChannel)
}
