// Package giftCache provides persistent caching functionality for the gift buying system.
package giftCache

import (
	"encoding/json"
	"gift-buyer/pkg/logger"
	"os"
	"strconv"

	"github.com/gotd/td/tg"
)

// CachedGift represents a simplified gift structure for JSON serialization.
// It contains only the essential fields needed for persistence and comparison,
// reducing storage overhead and avoiding complex nested structures.
type CachedGift struct {
	// ID is the unique identifier of the gift
	ID int64 `json:"id"`

	// Stars is the price of the gift in Telegram stars
	Stars int64 `json:"stars"`
}

// loadFromFile loads cached gift data from the cache.json file.
// It reads the JSON file, parses the cached gifts, and populates the in-memory cache.
// If the file doesn't exist or contains invalid data, it logs a warning and continues
// with an empty cache.
//
// The method is called during cache initialization to restore previously cached gifts.
// It reconstructs StarGift objects from the simplified CachedGift structures.
func (gc *GiftCacheImpl) loadFromFile() {
	data, err := os.ReadFile("cache.json")
	if err != nil {
		if !os.IsNotExist(err) {
			logger.GlobalLogger.Warnf("Failed to read cache file: %v", err)
		}
		return
	}

	var cachedGifts map[string]CachedGift
	if err := json.Unmarshal(data, &cachedGifts); err != nil {
		logger.GlobalLogger.Warnf("Failed to unmarshal cache file: %v", err)
		return
	}

	gc.mu.Lock()
	count := 0
	for _, cached := range cachedGifts {
		gift := &tg.StarGift{
			ID:    cached.ID,
			Stars: cached.Stars,
		}
		gc.cache[cached.ID] = gift
		count++
	}
	gc.mu.Unlock()

	logger.GlobalLogger.Infof("Loaded %d gifts from cache file", count)
}

// saveToFile persists the current cache state to the cache.json file.
// It merges new gifts with existing cached data to avoid overwriting
// previously saved gifts, then writes the updated data to disk.
//
// The method:
//  1. Reads existing cache file to preserve previously saved data
//  2. Identifies new gifts that haven't been saved yet
//  3. Merges new gifts with existing cached data
//  4. Writes the complete dataset to cache.json
//
// Only new gifts are added to the file to optimize I/O operations.
// If no new gifts are found, the save operation is skipped.
func (gc *GiftCacheImpl) saveToFile() {
	var existingCache map[string]CachedGift
	if data, err := os.ReadFile("cache.json"); err == nil {
		if unmarshalErr := json.Unmarshal(data, &existingCache); unmarshalErr != nil {
			logger.GlobalLogger.Warnf("Failed to unmarshal existing cache: %v", unmarshalErr)
		}
	}
	if existingCache == nil {
		existingCache = make(map[string]CachedGift)
	}

	gc.mu.RLock()
	newGifts := 0
	for id, gift := range gc.cache {
		key := strconv.FormatInt(id, 10)
		if _, exists := existingCache[key]; !exists {
			existingCache[key] = CachedGift{
				ID:    gift.ID,
				Stars: gift.Stars,
			}
			newGifts++
		}
	}
	gc.mu.RUnlock()

	if newGifts == 0 {
		return
	}

	jsonData, err := json.MarshalIndent(existingCache, "", "  ")
	if err != nil {
		logger.GlobalLogger.Errorf("Failed to marshal cache: %v", err)
		return
	}

	if err := os.WriteFile("cache.json", jsonData, 0600); err != nil {
		logger.GlobalLogger.Errorf("Failed to write cache to file: %v", err)
		return
	}

	logger.GlobalLogger.Infof("Saved %d new gifts to cache file", newGifts)
}
