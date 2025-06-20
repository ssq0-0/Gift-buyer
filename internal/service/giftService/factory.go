// Package giftService provides the main orchestration layer for the gift buying system.
package giftService

import (
	"bufio"
	"context"
	"fmt"
	"gift-buyer/internal/config"
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
	"gift-buyer/pkg/logger"
	"os"
	"strings"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
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
func (f *Factory) CreateSystem() (GiftService, error) {
	deviceConfig := f.createDeviceConfig()
	client := telegram.NewClient(f.cfg.TgSettings.AppId, f.cfg.TgSettings.ApiHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{
			Path: "session.json",
		},
		Device: telegram.DeviceConfig{
			DeviceModel:    deviceConfig.DeviceModel,
			SystemVersion:  deviceConfig.SystemVersion,
			AppVersion:     deviceConfig.AppVersion,
			SystemLangCode: deviceConfig.SystemLangCode,
			LangCode:       deviceConfig.LangCode,
			LangPack:       deviceConfig.LangPack,
		},
	})
	ctx, cancel := context.WithCancel(context.Background())

	api, err := f.initClient(client, ctx)
	if err != nil {
		cancel()
		return nil, err
	}

	var botClient *tg.Client
	if f.cfg.TgSettings.TgBotKey != "" {
		botClient, err = f.createBotClient(ctx)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to create bot client: %w", err)
		}
	}

	validator := giftValidator.NewGiftValidator(f.cfg.Criterias, f.cfg.TotalStarCap, f.cfg.TestMode, f.cfg.LimitedStatus)
	manager := giftManager.NewGiftManager(client.API())
	cache := giftCache.NewGiftCache()
	userCache := idCache.NewIDCache()
	notification := giftNotification.NewNotification(botClient, &f.cfg.TgSettings)
	monitor := giftMonitor.NewGiftMonitor(cache, manager, validator, notification, time.Duration(f.cfg.Ticker*1000)*time.Millisecond)
	rl := rateLimiter.NewRateLimiter(f.cfg.RPCRateLimit)
	counter := atomicCounter.NewAtomicCounter(f.cfg.MaxBuyCount)
	invoiceCreator := invoiceCreator.NewInvoiceCreator(f.cfg.Receiver.UserReceiverID, f.cfg.Receiver.ChannelReceiverID, f.cfg.Receiver.Type, userCache)
	paymentProcessor := paymentProcessor.NewPaymentProcessor(client.API(), invoiceCreator, rl)
	purchaseProcessor := purchaseProcessor.NewPurchaseProcessor(client.API(), paymentProcessor)
	monitorProcessor := giftBuyerMonitoring.NewGiftBuyerMonitoring(client.API(), notification)
	accountManager := accountManager.NewAccountManager(client.API(), f.cfg.Receiver.UserReceiverID, f.cfg.Receiver.ChannelReceiverID, userCache)
	buyer := giftBuyer.NewGiftBuyer(client.API(), f.cfg.Receiver.UserReceiverID, f.cfg.Receiver.ChannelReceiverID, f.cfg.Receiver.Type, manager, notification, f.cfg.MaxBuyCount, f.cfg.RetryCount, userCache, f.cfg.ConcurrencyGiftCount, rl, f.cfg.ConcurrentOperations, invoiceCreator, purchaseProcessor, monitorProcessor, counter)

	service := NewGiftService(
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
	)

	return service, nil
}

func (f *Factory) createDeviceConfig() telegram.DeviceConfig {
	config := telegram.DeviceConfig{}

	config.SetDefaults()

	if f.cfg.TgSettings.DeviceModel != "" {
		config.DeviceModel = f.cfg.TgSettings.DeviceModel
	} else {
		config.DeviceModel = "MacBook Pro M1 Pro"
	}

	if f.cfg.TgSettings.SystemVersion != "" {
		config.SystemVersion = f.cfg.TgSettings.SystemVersion
	} else {
		config.SystemVersion = "macOS 14.1"
	}

	if f.cfg.TgSettings.AppVersion != "" {
		config.AppVersion = f.cfg.TgSettings.AppVersion
	} else {
		config.AppVersion = "11.9 (272031) APP_STORE"
	}

	if f.cfg.TgSettings.SystemLangCode != "" {
		config.SystemLangCode = f.cfg.TgSettings.SystemLangCode
	} else {
		config.SystemLangCode = "en"
	}

	if f.cfg.TgSettings.LangCode != "" {
		config.LangCode = f.cfg.TgSettings.LangCode
	} else {
		config.LangCode = "en"
	}

	if f.cfg.TgSettings.LangPack != "" {
		config.LangPack = f.cfg.TgSettings.LangPack
	} else {
		config.LangPack = "macos"
	}

	return config
}

// initClient initializes and authenticates the main Telegram user client.
// It handles the complete authentication flow including 2FA, session management,
// and interactive code input when required.
//
// The authentication process:
//  1. Checks for existing valid session
//  2. Initiates authentication flow if needed
//  3. Handles phone number and password authentication
//  4. Prompts for verification code interactively
//  5. Manages session persistence and recovery
//
// Parameters:
//   - client: configured Telegram client instance
//   - ctx: context for cancellation and timeout control
//
// Returns:
//   - *tg.Client: authenticated Telegram API client
//   - error: authentication error, network error, or timeout
func (f *Factory) initClient(client *telegram.Client, ctx context.Context) (*tg.Client, error) {
	authDone := make(chan *tg.Client, 1)
	errCh := make(chan error, 1)

	go func() {
		err := client.Run(ctx, func(ctx context.Context) error {
			status, err := client.Auth().Status(ctx)
			if err == nil && status.Authorized {
				logger.GlobalLogger.Info("Already authorized, using existing session")
				authDone <- client.API()
				<-ctx.Done()
				return nil
			}

			logger.GlobalLogger.Info("Starting Telegram authentication...")
			// codePrompt provides interactive code input for 2FA verification
			codePrompt := func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
				fmt.Print("Enter code: ")
				code, err := bufio.NewReader(os.Stdin).ReadString('\n')
				if err != nil {
					return "", err
				}
				return strings.TrimSpace(code), nil
			}

			err = auth.NewFlow(
				auth.Constant(f.cfg.TgSettings.Phone, f.cfg.TgSettings.Password, auth.CodeAuthenticatorFunc(codePrompt)),
				auth.SendCodeOptions{},
			).Run(ctx, client.Auth())
			if err != nil {
				logger.GlobalLogger.Errorf("Authentication failed: %v", err)
				if strings.Contains(err.Error(), "AUTH_RESTART") {
					logger.GlobalLogger.Warn("AUTH_RESTART received, clearing session file")
					if removeErr := os.Remove("session.json"); removeErr != nil {
						logger.GlobalLogger.Warnf("Failed to remove session file: %v", removeErr)
					}
				}
				return err
			}

			logger.GlobalLogger.Info("Authentication successful!")
			authDone <- client.API()
			<-ctx.Done()
			return nil
		})
		if err != nil {
			errCh <- err
		}
	}()

	select {
	case api := <-authDone:
		logger.GlobalLogger.Info("Ready to start gift service")
		return api, nil
	case err := <-errCh:
		return nil, fmt.Errorf("telegram client initialization failed: %w", err)
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled during authentication")
	case <-time.After(120 * time.Second):
		return nil, fmt.Errorf("authentication timeout")
	}
}

// createBotClient creates and authenticates a Telegram bot client for notifications.
// It initializes a separate bot session for sending notifications and status updates
// to the configured chat, independent of the main user client.
//
// The bot authentication process:
//  1. Creates bot client with separate session storage
//  2. Authenticates using the bot token
//  3. Verifies bot permissions and access
//  4. Returns ready-to-use bot API client
//
// Parameters:
//   - ctx: context for cancellation and timeout control
//
// Returns:
//   - *tg.Client: authenticated bot API client for notifications
//   - error: bot authentication error, invalid token, or network error
func (f *Factory) createBotClient(ctx context.Context) (*tg.Client, error) {
	if f.cfg.TgSettings.TgBotKey == "" {
		return nil, fmt.Errorf("bot token is not configured")
	}

	botClient := telegram.NewClient(f.cfg.TgSettings.AppId, f.cfg.TgSettings.ApiHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{
			Path: "bot_session.json",
		},
	})

	botAPI := make(chan *tg.Client, 1)
	errCh := make(chan error, 1)

	go func() {
		err := botClient.Run(ctx, func(ctx context.Context) error {
			_, err := botClient.Auth().Bot(ctx, f.cfg.TgSettings.TgBotKey)
			if err != nil {
				logger.GlobalLogger.Errorf("Bot authentication failed: %v", err)
				return err
			}

			logger.GlobalLogger.Info("Bot authenticated successfully!")
			botAPI <- botClient.API()
			<-ctx.Done()
			return nil
		})
		if err != nil {
			logger.GlobalLogger.Errorf("Bot client error: %v", err)
			errCh <- err
		}
	}()

	select {
	case api := <-botAPI:
		logger.GlobalLogger.Info("Bot ready for notifications")
		return api, nil
	case err := <-errCh:
		return nil, fmt.Errorf("bot client initialization failed: %w", err)
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled during bot authentication")
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("bot authentication timeout")
	}
}
