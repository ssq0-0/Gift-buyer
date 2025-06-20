// Package giftBuyer provides gift purchasing functionality for the gift buying system.
// It handles the complete purchase workflow including payment processing, retry logic,
// balance validation, and concurrent purchase management with configurable limits.
package giftBuyer

import (
	"context"
	"gift-buyer/internal/service/giftService/giftInterfaces"
	"gift-buyer/internal/service/giftService/giftTypes"
	"gift-buyer/pkg/errors"
	"gift-buyer/pkg/logger"
	"sync"

	"github.com/gotd/td/tg"
)

// GiftBuyerImpl implements the GiftBuyer interface for purchasing Telegram star gifts.
// It manages the complete purchase workflow including payment processing, retry logic,
// balance validation, and purchase counting with configurable limits.
type GiftBuyerImpl struct {
	// manager handles gift-related operations and API communication
	manager giftInterfaces.Giftmanager

	// idCache is the cache for user IDs
	idCache giftInterfaces.UserCache

	// notification sends purchase status updates and notifications
	notification giftInterfaces.NotificationService

	// invoiceCreator creates invoices for gift purchases
	invoiceCreator giftInterfaces.InvoiceCreator

	// api is the Telegram client used for payment operations
	api *tg.Client

	// userReceiver is the ID of the gift recipient
	userReceiver []int

	// channelReceiver is the ID of the gift recipient
	channelReceiver []int

	// receiverType specifies the type of receiver (1 for user, 2 for channel)
	receiverType []int

	// counter tracks and limits the total number of purchases
	counter giftInterfaces.Counter

	retryCount           int
	concurrentGifts      int
	concurrentOperations int

	// requestCounter provides unique identifiers for requests to avoid FormID duplicates
	requestCounter int64
	rateLimiter    giftInterfaces.RateLimiter

	purchaseProcessor giftInterfaces.PurchaseProcessor
	monitorProcessor  giftInterfaces.MonitorProcessor
}

// NewGiftBuyer creates a new GiftBuyer instance with the specified configuration.
// It initializes the buyer with API client, recipient information, and purchase limits.
//
// Parameters:
//   - api: configured Telegram API client for payment operations
//   - receiver: Telegram ID of the gift recipient
//   - receiverType: type of receiver (1 for user, 2 for channel)
//   - manager: gift manager for API operations
//   - notification: notification service for status updates
//   - maxBuyCount: maximum number of gifts that can be purchased
//   - concurrentGifts: maximum number of concurrent gift purchases
//   - concurrentOperations: maximum number of concurrent operations
//
// Returns:
//   - giftInterfaces.GiftBuyer: configured gift buyer instance
func NewGiftBuyer(
	api *tg.Client,
	userIds,
	channelIds,
	receiverType []int,
	manager giftInterfaces.Giftmanager,
	notification giftInterfaces.NotificationService,
	maxBuyCount int64,
	retryCount int,
	idCache giftInterfaces.UserCache,
	concurrentGifts int,
	rateLimiter giftInterfaces.RateLimiter,
	concurrentOperations int,
	invoiceCreator giftInterfaces.InvoiceCreator,
	purchaseProcessor giftInterfaces.PurchaseProcessor,
	monitorProcessor giftInterfaces.MonitorProcessor,
	counter giftInterfaces.Counter,
) giftInterfaces.GiftBuyer {
	return &GiftBuyerImpl{
		api:                  api,
		userReceiver:         userIds,
		channelReceiver:      channelIds,
		receiverType:         receiverType,
		manager:              manager,
		notification:         notification,
		counter:              counter,
		retryCount:           retryCount,
		idCache:              idCache,
		concurrentGifts:      concurrentGifts,
		concurrentOperations: concurrentOperations,
		requestCounter:       0,
		rateLimiter:          rateLimiter,
		invoiceCreator:       invoiceCreator,
		purchaseProcessor:    purchaseProcessor,
		monitorProcessor:     monitorProcessor,
	}
}

// BuyGift attempts to purchase the specified gifts with their respective quantities.
// It handles concurrent purchases, retry logic, balance validation, and purchase limits.
//
// The purchase process:
//  1. Validates that gifts are provided
//  2. Launches concurrent goroutines for each gift type
//  3. Attempts individual purchases with retry logic
//  4. Collects results and sends status notifications
//  5. Returns success or aggregated error information
//
// Parameters:
//   - ctx: context for request cancellation and timeout control
//   - gifts: map of gifts to their desired purchase quantities
//
// Returns:
//   - error: purchase error, payment failure, or aggregated error from multiple failures
func (gm *GiftBuyerImpl) BuyGift(ctx context.Context, gifts map[*tg.StarGift]int64) {
	var (
		wg        sync.WaitGroup
		sem       = make(chan struct{}, gm.concurrentGifts)
		resultsCh = make(chan giftTypes.GiftResult)
		doneCh    = make(chan struct{})
	)

	go gm.monitorProcessor.MonitorProcess(ctx, resultsCh, doneCh, gifts)

	for gift, count := range gifts {
		wg.Add(1)
		go func(gift *tg.StarGift, count int64) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			gm.buyGift(ctx, gift, count, resultsCh)
		}(gift, count)
	}

	go func() {
		wg.Wait()
		close(doneCh)
	}()
}

// buyGift attempts to purchase a specific gift multiple times with retry logic.
// It handles individual gift purchases, manages the purchase counter, and implements
// asynchronous retry logic where each attempt is a separate goroutine.
//
// Parameters:
//   - ctx: context for request cancellation and timeout control
//   - gift: the star gift to purchase
//   - count: number of times to purchase this gift
//
// Returns:
//   - int64: number of successful purchases completed
//   - error: purchase error after all retry attempts exhausted
func (gm *GiftBuyerImpl) buyGift(ctx context.Context, gift *tg.StarGift, count int64, resChan chan<- giftTypes.GiftResult) {
	var (
		wg  sync.WaitGroup
		sem = make(chan struct{}, gm.concurrentOperations)
	)

	for i := int64(0); i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			gm.buyGiftWithRetry(ctx, gift, resChan)
		}()
	}

	wg.Wait()
}

func (gm *GiftBuyerImpl) buyGiftWithRetry(ctx context.Context, gift *tg.StarGift, resChan chan<- giftTypes.GiftResult) {
	var lastErr error

	for j := 0; j < gm.retryCount; j++ {
		select {
		case <-ctx.Done():
			resChan <- giftTypes.GiftResult{
				GiftID:  gift.ID,
				Success: false,
				Err:     ctx.Err(),
			}
			return
		default:
		}

		if !gm.counter.TryIncrement() {
			lastErr = errors.New("max buy count reached")
			resChan <- giftTypes.GiftResult{
				GiftID:  gift.ID,
				Success: false,
				Err:     lastErr,
			}
			return
		}

		if err := gm.purchaseProcessor.PurchaseGift(ctx, gift); err != nil {
			gm.counter.Decrement()
			lastErr = err
			continue
		}

		logger.GlobalLogger.Infof("Successfully purchased gift %d on attempt %d", gift.ID, j+1)
		resChan <- giftTypes.GiftResult{
			GiftID:  gift.ID,
			Success: true,
			Err:     nil,
		}
		return
	}

	resChan <- giftTypes.GiftResult{
		GiftID:  gift.ID,
		Success: false,
		Err:     lastErr,
	}
}

func (gm *GiftBuyerImpl) Close() {
	gm.rateLimiter.Close()
}
