package usecase

import (
	"context"
	"fmt"
	"gift-buyer/internal/config"
	"gift-buyer/internal/infrastructure/gitVersion"
	"gift-buyer/internal/service/authService"
	"gift-buyer/internal/service/authService/apiChecker"
	"gift-buyer/internal/service/authService/sessions"
	"gift-buyer/internal/service/giftService/accountManager"
	"gift-buyer/internal/service/giftService/cache/giftCache"
	"gift-buyer/internal/service/giftService/cache/idCache"
	"gift-buyer/internal/service/giftService/giftBuyer"
	"gift-buyer/internal/service/giftService/giftBuyer/atomicCounter"
	"gift-buyer/internal/service/giftService/giftBuyer/giftBuyerMonitoring"
	"gift-buyer/internal/service/giftService/giftBuyer/invoiceCreator"
	"gift-buyer/internal/service/giftService/giftBuyer/paymentProcessor"
	"gift-buyer/internal/service/giftService/giftBuyer/purchaseProcessor"
	"gift-buyer/internal/service/giftService/giftManager"
	"gift-buyer/internal/service/giftService/giftMonitor"
	"gift-buyer/internal/service/giftService/giftNotification"
	"gift-buyer/internal/service/giftService/giftValidator"
	"gift-buyer/internal/service/giftService/rateLimiter"
	"time"

	"github.com/gotd/td/tg"
)

// Factory provides a centralized way to create and configure the complete gift buying system.
// It handles the complex initialization of all components including Telegram clients,
// authentication, and dependency wiring with proper error handling.
type Factory struct {
	// cfg contains the software configuration for the gift buying system
	cfg *config.SoftConfig
}

// NewFactory creates a new Factory instance with the specified configuration.
// The factory will use this configuration to initialize all system components.
//
// Parameters:
//   - cfg: software configuration containing Telegram settings, criteria, and operational parameters
//
// Returns:
//   - *Factory: configured factory instance ready to create the gift buying system
func NewFactory(cfg *config.SoftConfig) *Factory {
	return &Factory{cfg: cfg}
}

// CreateSystem creates and initializes the complete gift buying system.
// It sets up Telegram clients, handles authentication, creates all service components,
// and wires them together into a functional gift buying service.
//
// The initialization process:
//  1. Creates and configures Telegram user client
//  2. Handles user authentication (including 2FA if required)
//  3. Creates and authenticates bot client for notifications
//  4. Initializes all service components (validator, manager, cache, etc.)
//  5. Wires components together into the main service
//
// Returns:
//   - GiftService: fully configured and ready-to-use gift buying service
//   - error: initialization error, authentication failure, or configuration error
//
// Possible errors:
//   - Telegram authentication failures
//   - Bot client initialization errors
//   - Network connectivity issues
//   - Invalid configuration parameters
func (f *Factory) CreateSystem() (UseCase, error) {
	ctx, cancel := context.WithCancel(context.Background())

	sessionManager := sessions.NewSessionManager(&f.cfg.TgSettings)
	authManager := authService.NewAuthManager(sessionManager, nil, &f.cfg.TgSettings)
	api, err := authManager.InitClient(ctx)
	if err != nil {
		cancel()
		return nil, err
	}

	apiChecker := apiChecker.NewApiChecker(api, time.NewTicker(time.Duration(f.cfg.Ticker*1000)*time.Millisecond))
	authManager.SetApiChecker(apiChecker)
	authManager.RunApiChecker(ctx)

	var botClient *tg.Client
	if f.cfg.TgSettings.TgBotKey != "" {
		botClient, err = authManager.InitBotClient(ctx)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to create bot client: %w", err)
		}
	}

	validator := giftValidator.NewGiftValidator(f.cfg.Criterias, f.cfg.TotalStarCap, f.cfg.TestMode, f.cfg.LimitedStatus)
	manager := giftManager.NewGiftManager(api)
	cache := giftCache.NewGiftCache()
	userCache := idCache.NewIDCache()
	notification := giftNotification.NewNotification(botClient, &f.cfg.TgSettings)
	monitor := giftMonitor.NewGiftMonitor(cache, manager, validator, notification, time.Duration(f.cfg.Ticker*1000)*time.Millisecond)
	authManager.SetMonitor(monitor)
	rl := rateLimiter.NewRateLimiter(f.cfg.RPCRateLimit)
	counter := atomicCounter.NewAtomicCounter(f.cfg.MaxBuyCount)
	invoiceCreator := invoiceCreator.NewInvoiceCreator(f.cfg.Receiver.UserReceiverID, f.cfg.Receiver.ChannelReceiverID, f.cfg.Receiver.Type, userCache)
	paymentProcessor := paymentProcessor.NewPaymentProcessor(api, invoiceCreator, rl)
	purchaseProcessor := purchaseProcessor.NewPurchaseProcessor(api, paymentProcessor)
	monitorProcessor := giftBuyerMonitoring.NewGiftBuyerMonitoring(api, notification)
	accountManager := accountManager.NewAccountManager(api, f.cfg.Receiver.UserReceiverID, f.cfg.Receiver.ChannelReceiverID, userCache)
	buyer := giftBuyer.NewGiftBuyer(api, f.cfg.Receiver.UserReceiverID, f.cfg.Receiver.ChannelReceiverID, f.cfg.Receiver.Type, manager, notification, f.cfg.MaxBuyCount, f.cfg.RetryCount, userCache, f.cfg.ConcurrencyGiftCount, rl, f.cfg.ConcurrentOperations, invoiceCreator, purchaseProcessor, monitorProcessor, counter)
	gitVersion := gitVersion.NewGitVersionController(f.cfg.RepoOwner, f.cfg.RepoName, f.cfg.ApiLink)

	service := NewUseCase(
		manager,
		validator,
		cache,
		notification,
		monitor,
		buyer,
		ctx,
		cancel,
		api,
		accountManager,
		gitVersion,
		time.NewTicker(time.Duration(f.cfg.UpdateTicker)*time.Second),
	)

	return service, nil
}
