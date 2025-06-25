// Package giftService provides the main orchestration layer for the gift buying system.
// It coordinates between monitoring, validation, purchasing, and notification components
// to create a complete automated gift buying solution.
package giftService

import (
	"context"
	"fmt"
	"gift-buyer/internal/infrastructure/gitVersion/gitInterfaces"
	"gift-buyer/internal/service/giftService/giftInterfaces"
	"gift-buyer/pkg/logger"
	"sync"
	"time"

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

	// SetIds sets the IDs of the accounts
	SetIds(ctx context.Context) error

	// CheckForUpdates checks for updates and sends a notification if available
	CheckForUpdates()
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

	// accountManager handles account-related operations
	accountManager giftInterfaces.AccountManager

	// version is the current version of the service
	gitVersion   gitInterfaces.GitVersionController
	updateTicker *time.Ticker
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
	accountManager giftInterfaces.AccountManager,
	gitVersion gitInterfaces.GitVersionController,
	updateTicker *time.Ticker,
) GiftService {
	return &GiftServiceImpl{
		manager:        manager,
		validator:      validator,
		cache:          cache,
		notification:   notification,
		monitor:        monitor,
		buyer:          buyer,
		ctx:            ctx,
		cancel:         cancel,
		api:            api,
		accountManager: accountManager,
		gitVersion:     gitVersion,
		updateTicker:   updateTicker,
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
			tc.wg.Wait()
			return
		default:
			newGifts, err := tc.monitor.Start(tc.ctx)
			if err != nil {
				if tc.ctx.Err() != nil {
					logger.GlobalLogger.Info("Context cancelled, stopping service")
					tc.wg.Wait()
					return
				}
				logger.GlobalLogger.Error("Error checking for new gifts", "error", err)
				continue
			}

			if len(newGifts) > 0 {
				logger.GlobalLogger.Infof("Found %d new gift types to process", len(newGifts))
				tc.wg.Add(2)
				go func() {
					defer tc.wg.Done()
					for gift, count := range newGifts {
						if err := tc.notification.SendNewGiftNotification(tc.ctx, gift); err != nil {
							logger.GlobalLogger.Error("Error sending notification", "error", err, "gift_id", gift.ID, "count", count)
						}
					}
				}()
				go func() {
					defer tc.wg.Done()
					tc.buyer.BuyGift(tc.ctx, newGifts)
				}()

				continue
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

	// Закрываем buyer и освобождаем ресурсы (rate limiter)
	if tc.buyer != nil {
		tc.buyer.Close()
	}
}

func (tc *GiftServiceImpl) SetIds(ctx context.Context) error {
	return tc.accountManager.SetIds(ctx)
}

func (tc *GiftServiceImpl) CheckForUpdates() {
	if err := tc.checkNewUpdates(); err != nil {
		logger.GlobalLogger.Error("Error checking for updates", "error", err)
	}
	for {
		select {
		case <-tc.ctx.Done():
			return
		case <-tc.updateTicker.C:
			if err := tc.checkNewUpdates(); err != nil {
				logger.GlobalLogger.Error("Error checking for updates", "error", err)
			}
		}
	}
}

func (tc *GiftServiceImpl) checkNewUpdates() error {
	localVersion, err := tc.gitVersion.GetCurrentVersion()
	if err != nil {
		logger.GlobalLogger.Error("Error getting current version", "error", err)
		return err
	}

	remoteVersion, err := tc.gitVersion.GetLatestVersion()
	if err != nil {
		logger.GlobalLogger.Error("Error getting latest version", "error", err)
		return err
	}

	ok, err := tc.gitVersion.CompareVersions(localVersion, remoteVersion.TagName)
	if err != nil {
		logger.GlobalLogger.Error("Error comparing versions", "error", err)
		return err
	}

	if ok {
		if err := tc.notification.SendUpdateNotification(tc.ctx, remoteVersion.TagName, fmt.Sprintf("%s\n", remoteVersion.Body)); err != nil {
			logger.GlobalLogger.Error("Error sending update notification", "error", err)
		}
	}
	return nil
}
