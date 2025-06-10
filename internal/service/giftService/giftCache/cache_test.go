package giftCache

import (
	"sync"
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
)

func TestNewGiftCache(t *testing.T) {
	cache := NewGiftCache()
	assert.NotNil(t, cache)

	// Verify it implements the interface
	_, ok := cache.(*GiftCacheImpl)
	assert.True(t, ok)
}

func TestGiftCache_SetAndGetGift(t *testing.T) {
	cache := NewGiftCache()

	gift := &tg.StarGift{
		ID:    123,
		Stars: 500,
	}

	// Set gift
	cache.SetGift(123, gift)

	// Get gift
	retrievedGift, err := cache.GetGift(123)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedGift)
	assert.Equal(t, gift.ID, retrievedGift.ID)
	assert.Equal(t, gift.Stars, retrievedGift.Stars)
}

func TestGiftCache_GetNonExistentGift(t *testing.T) {
	cache := NewGiftCache()

	gift, err := cache.GetGift(999)
	assert.NoError(t, err)
	assert.Nil(t, gift)
}

func TestGiftCache_HasGift(t *testing.T) {
	cache := NewGiftCache()

	gift := &tg.StarGift{
		ID:    456,
		Stars: 750,
	}

	// Initially should not have the gift
	assert.False(t, cache.HasGift(456))

	// Set gift
	cache.SetGift(456, gift)

	// Now should have the gift
	assert.True(t, cache.HasGift(456))

	// Should not have other gifts
	assert.False(t, cache.HasGift(789))
}

func TestGiftCache_DeleteGift(t *testing.T) {
	cache := NewGiftCache()

	gift := &tg.StarGift{
		ID:    789,
		Stars: 1000,
	}

	// Set gift
	cache.SetGift(789, gift)
	assert.True(t, cache.HasGift(789))

	// Delete gift
	cache.DeleteGift(789)
	assert.False(t, cache.HasGift(789))

	// Verify it's actually gone
	retrievedGift, err := cache.GetGift(789)
	assert.NoError(t, err)
	assert.Nil(t, retrievedGift)
}

func TestGiftCache_DeleteNonExistentGift(t *testing.T) {
	cache := NewGiftCache()

	// Should not panic when deleting non-existent gift
	assert.NotPanics(t, func() {
		cache.DeleteGift(999)
	})
}

func TestGiftCache_GetAllGifts(t *testing.T) {
	cache := NewGiftCache()

	// Initially should be empty
	allGifts := cache.GetAllGifts()
	assert.NotNil(t, allGifts)
	assert.Empty(t, allGifts)

	// Add some gifts
	gift1 := &tg.StarGift{ID: 1, Stars: 100}
	gift2 := &tg.StarGift{ID: 2, Stars: 200}
	gift3 := &tg.StarGift{ID: 3, Stars: 300}

	cache.SetGift(1, gift1)
	cache.SetGift(2, gift2)
	cache.SetGift(3, gift3)

	// Get all gifts
	allGifts = cache.GetAllGifts()
	assert.Len(t, allGifts, 3)
	assert.Equal(t, gift1, allGifts[1])
	assert.Equal(t, gift2, allGifts[2])
	assert.Equal(t, gift3, allGifts[3])
}

func TestGiftCache_GetAllGifts_IsolatedCopy(t *testing.T) {
	cache := NewGiftCache()

	gift := &tg.StarGift{ID: 1, Stars: 100}
	cache.SetGift(1, gift)

	// Get all gifts
	allGifts1 := cache.GetAllGifts()
	allGifts2 := cache.GetAllGifts()

	// Modify one copy
	allGifts1[999] = &tg.StarGift{ID: 999, Stars: 999}

	// Other copy should not be affected
	assert.Len(t, allGifts2, 1)
	assert.NotContains(t, allGifts2, int64(999))

	// Original cache should not be affected
	assert.False(t, cache.HasGift(999))
}

func TestGiftCache_Clear(t *testing.T) {
	cache := NewGiftCache()

	// Add some gifts
	cache.SetGift(1, &tg.StarGift{ID: 1, Stars: 100})
	cache.SetGift(2, &tg.StarGift{ID: 2, Stars: 200})
	cache.SetGift(3, &tg.StarGift{ID: 3, Stars: 300})

	// Verify gifts are there
	assert.True(t, cache.HasGift(1))
	assert.True(t, cache.HasGift(2))
	assert.True(t, cache.HasGift(3))

	// Clear cache
	cache.Clear()

	// Verify cache is empty
	assert.False(t, cache.HasGift(1))
	assert.False(t, cache.HasGift(2))
	assert.False(t, cache.HasGift(3))

	allGifts := cache.GetAllGifts()
	assert.Empty(t, allGifts)
}

func TestGiftCache_UpdateGift(t *testing.T) {
	cache := NewGiftCache()

	// Set initial gift
	originalGift := &tg.StarGift{ID: 1, Stars: 100}
	cache.SetGift(1, originalGift)

	// Update gift
	updatedGift := &tg.StarGift{ID: 1, Stars: 200}
	cache.SetGift(1, updatedGift)

	// Verify update
	retrievedGift, err := cache.GetGift(1)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedGift)
	assert.Equal(t, int64(200), retrievedGift.Stars)
}

func TestGiftCache_ConcurrentAccess(t *testing.T) {
	cache := NewGiftCache()
	const numGoroutines = 100
	const numOperations = 100

	var wg sync.WaitGroup

	// Concurrent writes
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				gift := &tg.StarGift{
					ID:    int64(id*numOperations + j),
					Stars: int64(j),
				}
				cache.SetGift(gift.ID, gift)
			}
		}(i)
	}
	wg.Wait()

	// Verify all gifts were set
	allGifts := cache.GetAllGifts()
	assert.Len(t, allGifts, numGoroutines*numOperations)

	// Concurrent reads
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				giftID := int64(id*numOperations + j)
				gift, err := cache.GetGift(giftID)
				assert.NoError(t, err)
				assert.NotNil(t, gift)
				assert.Equal(t, giftID, gift.ID)
				assert.Equal(t, int64(j), gift.Stars)
			}
		}(i)
	}
	wg.Wait()
}

func TestGiftCache_ZeroIDGift(t *testing.T) {
	cache := NewGiftCache()

	gift := &tg.StarGift{
		ID:    0, // Zero ID
		Stars: 100,
	}

	cache.SetGift(0, gift)
	assert.True(t, cache.HasGift(0))

	retrievedGift, err := cache.GetGift(0)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedGift)
	assert.Equal(t, int64(0), retrievedGift.ID)
}

func TestGiftCache_NegativeIDGift(t *testing.T) {
	cache := NewGiftCache()

	gift := &tg.StarGift{
		ID:    -1, // Negative ID
		Stars: 100,
	}

	cache.SetGift(-1, gift)
	assert.True(t, cache.HasGift(-1))

	retrievedGift, err := cache.GetGift(-1)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedGift)
	assert.Equal(t, int64(-1), retrievedGift.ID)
}
