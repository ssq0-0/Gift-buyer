// Package giftInterfaces defines the core interfaces for the gift buying system.
// These interfaces provide abstractions for gift management, validation, purchasing,
// caching, monitoring, and notifications, enabling loose coupling and testability.
package giftInterfaces

import (
	"context"

	"github.com/gotd/td/tg"
)

// Giftmanager defines the interface for managing gift operations with Telegram API.
// It provides methods to retrieve available gifts from the Telegram platform.
type Giftmanager interface {
	// GetAvailableGifts retrieves all currently available star gifts from Telegram.
	// It returns a slice of StarGift objects representing gifts that can be purchased.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeout control
	//
	// Returns:
	//   - []*tg.StarGift: slice of available star gifts
	//   - error: API communication error or parsing error
	GetAvailableGifts(ctx context.Context) ([]*tg.StarGift, error)
}

// GiftValidator defines the interface for validating gifts against purchase criteria.
// It determines whether a gift meets the configured requirements for automatic purchase.
type GiftValidator interface {
	// IsEligible checks if a gift meets the configured purchase criteria.
	// It evaluates the gift against price ranges, supply limits, and other constraints.
	//
	// Parameters:
	//   - gift: the star gift to validate
	//
	// Returns:
	//   - int64: number of gifts to purchase if eligible (0 if not eligible)
	//   - bool: true if the gift meets criteria, false otherwise
	IsEligible(gift *tg.StarGift) (int64, bool)
}

// GiftBuyer defines the interface for purchasing gifts through Telegram API.
// It handles the actual purchase transactions and manages purchase limits.
type GiftBuyer interface {
	// BuyGift attempts to purchase the specified gifts with their respective quantities.
	// It handles payment processing, retry logic, and purchase confirmation.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeout control
	//   - gifts: map of gifts to their desired purchase quantities
	//
	// Returns:
	//   - error: purchase error, payment failure, or API communication error
	BuyGift(ctx context.Context, gifts map[*tg.StarGift]int64) error
}

// GiftCache defines the interface for caching gift information.
// It provides persistent storage for gift data to avoid redundant API calls
// and maintain state across application restarts.
type GiftCache interface {
	// SetGift stores a gift in the cache with the specified ID as key.
	//
	// Parameters:
	//   - id: unique identifier for the gift
	//   - gift: the star gift object to cache
	SetGift(id int64, gift *tg.StarGift)

	// GetGift retrieves a cached gift by its ID.
	//
	// Parameters:
	//   - id: unique identifier of the gift to retrieve
	//
	// Returns:
	//   - *tg.StarGift: the cached gift object, nil if not found
	//   - error: retrieval error (currently always nil)
	GetGift(id int64) (*tg.StarGift, error)

	// GetAllGifts returns a copy of all cached gifts.
	//
	// Returns:
	//   - map[int64]*tg.StarGift: map of gift IDs to gift objects
	GetAllGifts() map[int64]*tg.StarGift

	// HasGift checks if a gift with the specified ID exists in the cache.
	//
	// Parameters:
	//   - id: unique identifier of the gift to check
	//
	// Returns:
	//   - bool: true if the gift exists in cache, false otherwise
	HasGift(id int64) bool

	// DeleteGift removes a gift from the cache.
	//
	// Parameters:
	//   - id: unique identifier of the gift to remove
	DeleteGift(id int64)

	// Clear removes all gifts from the cache.
	Clear()
}

// GiftMonitor defines the interface for monitoring new gifts.
// It continuously checks for new gifts and identifies those eligible for purchase.
type GiftMonitor interface {
	// Start begins the gift monitoring process and returns newly discovered eligible gifts.
	// It runs continuously until the context is cancelled, checking for new gifts
	// at configured intervals.
	//
	// Parameters:
	//   - ctx: context for cancellation and timeout control
	//
	// Returns:
	//   - map[*tg.StarGift]int64: map of eligible gifts to their purchase quantities
	//   - error: monitoring error, API communication error, or context cancellation
	Start(ctx context.Context) (map[*tg.StarGift]int64, error)
}

// NotificationService defines the interface for sending notifications.
// It provides methods to notify users about new gifts and purchase status updates.
type NotificationService interface {
	// SendNewGiftNotification sends a notification about a newly discovered gift.
	// The notification includes gift details such as price, supply, and availability.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeout control
	//   - gift: the star gift to notify about
	//
	// Returns:
	//   - error: notification sending error or API communication error
	SendNewGiftNotification(ctx context.Context, gift *tg.StarGift) error

	// SendBuyStatus sends a notification about the purchase operation status.
	// It reports successful purchases or error conditions to the configured chat.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeout control
	//   - status: human-readable status message
	//   - err: error that occurred during purchase (nil for success)
	//
	// Returns:
	//   - error: notification sending error or API communication error
	SendBuyStatus(ctx context.Context, status string, err error) error

	// SetBot sets the bot client
	SetBot() bool
}

// UserCache defines the interface for caching user information.
// It provides persistent storage for user data to avoid redundant API calls
// and maintain state across application restarts.
type UserCache interface {
	// SetUser stores a user in the cache with the specified ID as key.
	//
	// Parameters:
	//   - id: unique identifier for the user
	//   - user: the user object to cache
	SetUser(user *tg.User)

	// GetUser retrieves a cached user by their ID.
	//
	// Parameters:
	//   - id: unique identifier of the user to retrieve
	//
	// Returns:
	//   - *tg.User: the cached user object, nil if not found
	//   - error: retrieval error (currently always nil)
	GetUser(id int64) (*tg.User, error)

	// SetChannel stores a channel in the cache with the specified ID as key.
	//
	// Parameters:
	//   - id: unique identifier for the channel
	//   - channel: the channel object to cache
	SetChannel(channel *tg.Channel)

	// GetChannel retrieves a cached channel by its ID.
	//
	// Parameters:
	//   - id: unique identifier of the channel to retrieve
	//
	// Returns:
	//   - *tg.Channel: the cached channel object, nil if not found
	//   - error: retrieval error (currently always nil)
	GetChannel(id int64) (*tg.Channel, error)
}

// BalanceCache defines the interface for caching the balance of the user
type BalanceCache interface {
	// SetBalance sets the balance of the user
	//
	// Parameters:
	//   - balance: the balance to set
	SetBalance(balance int64)

	// GetBalance gets the balance of the user
	//
	// Returns:
	//   - int64: the balance of the user
	GetBalance() int64

	// TrimBalance trims the balance of the user
	//
	// Parameters:
	//   - deduction: the amount to trim from the balance
	TrimBalance(deduction int64)
}
