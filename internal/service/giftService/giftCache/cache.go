// Package giftCache provides persistent caching functionality for the gift buying system.
// It implements thread-safe caching of gift data with automatic persistence to disk,
// enabling the system to maintain state across restarts and avoid redundant processing.
package giftCache

import (
	"gift-buyer/internal/service/giftService/giftInterfaces"
	"sync"
	"time"

	"github.com/gotd/td/tg"
)

// GiftCacheImpl implements the GiftCache interface for thread-safe gift caching.
// It provides in-memory storage with automatic periodic persistence to disk,
// ensuring data durability and fast access to cached gift information.
type GiftCacheImpl struct {
	// cache stores the in-memory gift data indexed by gift ID
	cache map[int64]*tg.StarGift

	// stopCh signals the periodic save goroutine to stop
	stopCh chan struct{}

	// interval defines how often the cache is persisted to disk
	interval time.Duration

	// mu provides thread-safe access to the cache map
	mu sync.RWMutex
}

// NewGiftCache creates a new GiftCache instance with automatic persistence.
// It initializes the cache, loads existing data from disk, and starts
// a background goroutine for periodic saving.
//
// The cache automatically:
//   - Loads existing data from cache.json on startup
//   - Saves new data to disk every 5 seconds
//   - Provides thread-safe concurrent access
//
// Returns:
//   - giftInterfaces.GiftCache: configured and initialized gift cache instance
func NewGiftCache() giftInterfaces.GiftCache {
	gc := &GiftCacheImpl{
		cache:    make(map[int64]*tg.StarGift),
		stopCh:   make(chan struct{}),
		interval: 5 * time.Second,
	}

	gc.loadFromFile()

	go gc.startPeriodicSave()

	return gc
}

// startPeriodicSave runs a background goroutine that periodically saves the cache to disk.
// It saves the cache at the configured interval and performs a final save when stopped.
// This goroutine runs until the stopCh channel is closed.
func (gc *GiftCacheImpl) startPeriodicSave() {
	ticker := time.NewTicker(gc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			gc.saveToFile()
		case <-gc.stopCh:
			gc.saveToFile()
			return
		}
	}
}

// SetGift stores a gift in the cache with the specified ID as the key.
// This operation is thread-safe and will trigger persistence on the next save cycle.
//
// Parameters:
//   - id: unique identifier for the gift (typically gift.ID)
//   - gift: the star gift object to cache
func (gc *GiftCacheImpl) SetGift(id int64, gift *tg.StarGift) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	gc.cache[id] = gift
}

// GetGift retrieves a cached gift by its ID.
// This operation is thread-safe and uses a read lock for optimal performance.
//
// Parameters:
//   - id: unique identifier of the gift to retrieve
//
// Returns:
//   - *tg.StarGift: the cached gift object, nil if not found
//   - error: always nil in current implementation
func (gc *GiftCacheImpl) GetGift(id int64) (*tg.StarGift, error) {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	gift, exists := gc.cache[id]
	if !exists {
		return nil, nil
	}
	return gift, nil
}

// GetAllGifts returns a copy of all cached gifts.
// This operation is thread-safe and returns a new map to prevent external modification.
//
// Returns:
//   - map[int64]*tg.StarGift: map of gift IDs to gift objects (copy of internal cache)
func (gc *GiftCacheImpl) GetAllGifts() map[int64]*tg.StarGift {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	result := make(map[int64]*tg.StarGift, len(gc.cache))
	for id, gift := range gc.cache {
		result[id] = gift
	}
	return result
}

// HasGift checks if a gift with the specified ID exists in the cache.
// This operation is thread-safe and uses a read lock for optimal performance.
//
// Parameters:
//   - id: unique identifier of the gift to check
//
// Returns:
//   - bool: true if the gift exists in cache, false otherwise
func (gc *GiftCacheImpl) HasGift(id int64) bool {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	_, exists := gc.cache[id]
	return exists
}

// DeleteGift removes a gift from the cache.
// This operation is thread-safe and will be reflected in the next save cycle.
//
// Parameters:
//   - id: unique identifier of the gift to remove
func (gc *GiftCacheImpl) DeleteGift(id int64) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	delete(gc.cache, id)
}

// Clear removes all gifts from the cache.
// This operation is thread-safe and creates a new empty cache map.
func (gc *GiftCacheImpl) Clear() {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	gc.cache = make(map[int64]*tg.StarGift)
}
