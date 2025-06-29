package sessions

import (
	"bufio"
	"context"
	"fmt"
	"gift-buyer/internal/config"
	"gift-buyer/pkg/logger"
	"os"
	"strings"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type sessionManagerImpl struct {
	cfg *config.TgSettings
}

func NewSessionManager(cfg *config.TgSettings) *sessionManagerImpl {
	return &sessionManagerImpl{
		cfg: cfg,
	}
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
func (f *sessionManagerImpl) InitUserAPI(client *telegram.Client, ctx context.Context) (*tg.Client, error) {
	authDone := make(chan *tg.Client, 1)
	errCh := make(chan error, 1)

	go func() {
		err := client.Run(ctx, func(ctx context.Context) error {
			status, err := client.Auth().Status(ctx)
			if err == nil && status.Authorized {
				logger.GlobalLogger.Info("Already authorized, using existing session")
				authDone <- client.API()
				<-ctx.Done()
				return ctx.Err()
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
				auth.Constant(f.cfg.Phone, f.cfg.Password, auth.CodeAuthenticatorFunc(codePrompt)),
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
			return ctx.Err()
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
	case <-time.After(560 * time.Second):
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
func (f *sessionManagerImpl) InitBotAPI(ctx context.Context) (*tg.Client, error) {
	if f.cfg.TgBotKey == "" {
		return nil, fmt.Errorf("bot token is not configured")
	}

	botClient := telegram.NewClient(f.cfg.AppId, f.cfg.ApiHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{
			Path: "bot_session.json",
		},
	})

	botAPI := make(chan *tg.Client, 1)
	errCh := make(chan error, 1)

	go func() {
		err := botClient.Run(ctx, func(ctx context.Context) error {
			_, err := botClient.Auth().Bot(ctx, f.cfg.TgBotKey)
			if err != nil {
				logger.GlobalLogger.Errorf("Bot authentication failed: %v", err)
				return err
			}

			logger.GlobalLogger.Info("Bot authenticated successfully!")
			botAPI <- botClient.API()
			<-ctx.Done()
			return ctx.Err()
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
