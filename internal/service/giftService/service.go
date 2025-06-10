// Package giftService provides the main orchestration layer for the gift buying system.
// It coordinates between monitoring, validation, purchasing, and notification components
// to create a complete automated gift buying solution.
package giftService

import (
	"context"
	"gift-buyer/internal/service/giftService/giftInterfaces"
	"gift-buyer/pkg/logger"
	"sync"

	"github.com/gotd/td/tg"
)

// GiftService defines the main interface for the gift buying service.
// It provides lifecycle management methods for starting and stopping the service.
type GiftService interface {
	// Start begins the gift monitoring and purchasing process.
	// This method runs continuously until stopped or context cancelled.
	Start()

	// Stop gracefully shuts down the gift service and all its components.
	Stop()
}

// GiftServiceImpl implements the GiftService interface and orchestrates all gift buying operations.
// It manages the lifecycle of monitoring, validation, purchasing, and notification components,
// providing a unified service that automatically discovers and purchases eligible gifts.
type GiftServiceImpl struct {
	// manager handles gift retrieval and API communication
	manager giftInterfaces.Giftmanager

	// validator evaluates gifts against purchase criteria
	validator giftInterfaces.GiftValidator

	// cache provides persistent storage for processed gifts
	cache giftInterfaces.GiftCache

	// notification sends alerts about discoveries and purchase status
	notification giftInterfaces.NotificationService

	// monitor continuously checks for new eligible gifts
	monitor giftInterfaces.GiftMonitor

	// buyer handles the actual gift purchase transactions
	buyer giftInterfaces.GiftBuyer

	// ctx provides cancellation context for the service
	ctx context.Context

	// cancel function to stop the service gracefully
	cancel context.CancelFunc

	// wg coordinates graceful shutdown of goroutines
	wg sync.WaitGroup

	// api is the main Telegram client for API operations
	api *tg.Client
}

// NewGiftService creates a new GiftService instance with all required dependencies.
// It wires together all components needed for automated gift buying operations.
//
// Parameters:
//   - manager: gift manager for API communication
//   - validator: gift validator for eligibility checking
//   - cache: gift cache for state persistence
//   - notification: notification service for alerts
//   - monitor: gift monitor for continuous discovery
//   - buyer: gift buyer for purchase operations
//   - ctx: context for cancellation control
//   - cancel: cancel function for graceful shutdown
//   - api: Telegram API client
//
// Returns:
//   - GiftService: configured gift service ready for operation
func NewGiftService(
	manager giftInterfaces.Giftmanager,
	validator giftInterfaces.GiftValidator,
	cache giftInterfaces.GiftCache,
	notification giftInterfaces.NotificationService,
	monitor giftInterfaces.GiftMonitor,
	buyer giftInterfaces.GiftBuyer,
	ctx context.Context,
	cancel context.CancelFunc,
	api *tg.Client,
) GiftService {
	return &GiftServiceImpl{
		manager:      manager,
		validator:    validator,
		cache:        cache,
		notification: notification,
		monitor:      monitor,
		buyer:        buyer,
		ctx:          ctx,
		cancel:       cancel,
		api:          api,
	}
}

// Start begins the main gift buying service loop.
// It continuously monitors for new gifts, validates them against criteria,
// and automatically purchases eligible gifts until the service is stopped.
//
// The service loop:
//  1. Monitors for new eligible gifts
//  2. Validates discovered gifts against criteria
//  3. Attempts to purchase eligible gifts
//  4. Handles errors and continues operation
//  5. Respects context cancellation for graceful shutdown
//
// This method blocks until the service is stopped or context is cancelled.
func (tc *GiftServiceImpl) Start() {
	for {
		select {
		case <-tc.ctx.Done():
			return
		default:
			newGifts, err := tc.monitor.Start(tc.ctx)
			if err != nil {
				if tc.ctx.Err() != nil {
					logger.GlobalLogger.Info("Context cancelled, stopping service")
					return
				}
				logger.GlobalLogger.Error("Error checking for new gifts", "error", err)
				continue
			}

			if len(newGifts) > 0 {
				err = tc.buyer.BuyGift(tc.ctx, newGifts)
				if err != nil {
					if tc.ctx.Err() != nil {
						logger.GlobalLogger.Info("Context cancelled, stopping service")
						return
					}
					logger.GlobalLogger.Error("Error buying gift", "error", err)
					continue
				}
				logger.GlobalLogger.Info("New gifts bought", "count", len(newGifts))
			}
		}
	}
}

// Stop gracefully shuts down the gift service.
// It cancels the service context and waits for all goroutines to complete
// before returning, ensuring clean shutdown of all components.
func (tc *GiftServiceImpl) Stop() {
	if tc.cancel != nil {
		tc.cancel()
	}
	tc.wg.Wait()
}
