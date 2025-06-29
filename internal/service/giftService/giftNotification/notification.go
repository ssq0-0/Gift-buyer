// Package giftNotification provides notification functionality for the gift buying system.
// It handles sending formatted notifications about new gifts and purchase status updates
// through Telegram bot API with retry logic and error handling.
package giftNotification

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"gift-buyer/internal/config"
	"gift-buyer/pkg/logger"
	mathRand "math/rand"
	"strings"
	"time"

	"github.com/gotd/td/tg"
)

// cryptoRandomInt63 генерирует криптографически стойкое случайное число
func cryptoRandomInt63() int64 {
	var randomBytes [8]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		// Fallback на math/rand если crypto/rand недоступен
		return mathRand.Int63()
	}
	// Безопасное преобразование с маскированием старшего бита
	val := binary.BigEndian.Uint64(randomBytes[:])
	return int64(val >> 1) // Сдвиг вправо гарантирует положительное значение
}

// NotificationServiceImpl implements the NotificationService interface for sending
// Telegram notifications about gift discoveries and purchase status updates.
// It provides formatted messages with retry logic and flood protection.
type notificationServiceImpl struct {
	// Bot is the Telegram bot client used for sending notifications
	Bot *tg.Client

	// Config contains Telegram settings including notification chat ID
	Config *config.TgSettings
}

// NewNotification creates a new NotificationService instance with the specified bot client and configuration.
// The service will use the bot to send notifications to the configured chat.
//
// Parameters:
//   - bot: configured Telegram bot client for sending messages
//   - config: Telegram settings containing notification chat ID and other parameters
//
// Returns:
//   - giftInterfaces.NotificationService: configured notification service instance
func NewNotification(bot *tg.Client, config *config.TgSettings) *notificationServiceImpl {
	return &notificationServiceImpl{
		Bot:    bot,
		Config: config,
	}
}

// sendNotification sends a message to the configured notification chat with retry logic.
// It handles flood protection, implements exponential backoff, and provides error recovery.
//
// The retry mechanism:
//   - Maximum 3 retry attempts
//   - Special handling for FLOOD_WAIT errors with 5-second delay
//   - Exponential backoff for other errors (2, 4, 6 seconds)
//   - Logs errors and continues operation on failure
//
// Parameters:
//   - ctx: context for request cancellation and timeout control
//   - message: the message text to send
//
// Returns:
//   - error: notification sending error after all retries exhausted
func (ns *notificationServiceImpl) sendNotification(ctx context.Context, message string) error {
	if ns.Bot == nil || ns.Config == nil || ns.Config.NotificationChatID == 0 {
		logger.GlobalLogger.Warn("Bot client or notification chat ID not configured")
		return nil
	}

	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		_, err := ns.Bot.MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
			Peer: &tg.InputPeerUser{
				UserID: ns.Config.NotificationChatID,
			},
			Message:  message,
			RandomID: cryptoRandomInt63(),
		})

		if err == nil {
			return nil
		}

		if strings.Contains(err.Error(), "FLOOD_WAIT") {
			time.Sleep(5 * time.Second)
			continue
		}

		if attempt < maxRetries-1 {
			time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
			continue
		}

		logger.GlobalLogger.Errorf("Failed to send notification: %v", err)
		return err
	}

	return nil
}

// SendNewGiftNotification sends a formatted notification about a newly discovered gift.
// It creates a detailed message including gift information, pricing, availability,
// and current timestamp for tracking purposes.
//
// The notification includes:
//   - Gift title and ID
//   - Total and available supply with percentage
//   - Purchase price and conversion price in stars
//   - Current UTC timestamp
//   - Formatted numbers for better readability
//
// Parameters:
//   - ctx: context for request cancellation and timeout control
//   - gift: the star gift to notify about
//
// Returns:
//   - error: notification sending error or formatting error
func (ns *notificationServiceImpl) SendNewGiftNotification(ctx context.Context, gift *tg.StarGift) error {
	giftTitle, hasTitle := gift.GetTitle()
	if !hasTitle {
		giftTitle = "Unknown Gift"
	}

	giftID := gift.GetID()
	giftPrice := gift.GetStars()
	convertPrice := gift.GetConvertStars()
	giftSupply, hasSupply := gift.GetAvailabilityTotal()

	var availableAmount int
	var percentage float64

	if gift.Limited {
		remains, hasRemains := gift.GetAvailabilityRemains()
		if hasRemains {
			availableAmount = remains
			if hasSupply && giftSupply > 0 {
				percentage = float64(remains) / float64(giftSupply) * 100
			}
		}
	} else {
		availableAmount = giftSupply
		percentage = 100.0
	}

	currentTime := time.Now().UTC().Format("02-01-2006 15:04:05")

	message := fmt.Sprintf(`🎁 New gift detected!
%s (%d)

🎯 Total amount: %s
❓ Available amount: %d (%.0f%%, updated at %s UTC)

💎 Price: %s ⭐️
♻️ Convert price: %s ⭐️`,
		giftTitle,
		giftID,
		formatNumber(giftSupply),
		availableAmount,
		percentage,
		currentTime,
		formatNumber(int(giftPrice)),
		formatNumber(int(convertPrice)),
	)

	return ns.sendNotification(ctx, message)
}

// SendBuyStatus sends a notification about the purchase operation status.
// It reports successful purchases or error conditions with appropriate formatting
// and emoji indicators for quick visual identification.
//
// Parameters:
//   - ctx: context for request cancellation and timeout control
//   - status: human-readable status message describing the operation result
//   - err: error that occurred during purchase (nil for successful operations)
//
// Returns:
//   - error: notification sending error
func (ns *notificationServiceImpl) SendBuyStatus(ctx context.Context, status string, err error) error {
	var message string
	if err != nil {
		message = fmt.Sprintf("📊 Buy Status: %s\n❌ Error: %s", status, err.Error())
	} else {
		message = fmt.Sprintf("📊 Buy Status: %s\n✅ Success", status)
	}

	return ns.sendNotification(ctx, message)
}

func (ns *notificationServiceImpl) SendErrorNotification(ctx context.Context, err error) error {
	return ns.sendNotification(ctx, err.Error())
}

// SetBot sets the bot client
func (ns *notificationServiceImpl) SetBot() bool {
	return ns.Bot != nil
}

func (ns *notificationServiceImpl) SendUpdateNotification(ctx context.Context, version, message string) error {
	return ns.sendNotification(ctx, fmt.Sprintf("🆕 New version available: %s\n%s", version, message))
}

// formatNumber formats integers with comma separators for better readability.
// It adds commas every three digits to make large numbers easier to read.
//
// Examples:
//   - 1000 -> "1,000"
//   - 1234567 -> "1,234,567"
//   - 0 -> "0"
//
// Parameters:
//   - num: the integer to format
//
// Returns:
//   - string: formatted number with comma separators
func formatNumber(num int) string {
	if num == 0 {
		return "0"
	}

	str := fmt.Sprintf("%d", num)
	if len(str) <= 3 {
		return str
	}

	result := ""
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(digit)
	}

	return result
}
