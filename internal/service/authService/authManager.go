package authService

import (
	"context"
	"errors"
	"gift-buyer/internal/config"
	"gift-buyer/internal/service/authService/apiChecker"
	"gift-buyer/internal/service/authService/authInterfaces"
	"gift-buyer/pkg/logger"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type AuthManagerImpl struct {
	api            *tg.Client
	botApi         *tg.Client
	mu             sync.RWMutex
	sessionManager authInterfaces.SessionManager
	apiChecker     authInterfaces.ApiChecker
	cfg            *config.TgSettings
	reconnect      chan struct{}
	stopCh         chan struct{}
	wg             sync.WaitGroup
	monitor        authInterfaces.GiftMonitorAndAuthController
}

func NewAuthManager(sessionManager authInterfaces.SessionManager, apiChecker authInterfaces.ApiChecker, cfg *config.TgSettings) *AuthManagerImpl {
	return &AuthManagerImpl{
		sessionManager: sessionManager,
		apiChecker:     apiChecker,
		cfg:            cfg,
		reconnect:      make(chan struct{}, 1),
		stopCh:         make(chan struct{}),
	}
}

func (f *AuthManagerImpl) InitClient(ctx context.Context) (*tg.Client, error) {
	if f.cfg == nil {
		return nil, errors.New("configuration is nil")
	}

	if f.sessionManager == nil {
		return nil, errors.New("session manager is nil")
	}

	deviceConfig := f.sessionManager.CreateDeviceConfig()
	client := telegram.NewClient(f.cfg.AppId, f.cfg.ApiHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{
			Path: "session.json",
		},
		Device: deviceConfig,
	})

	api, err := f.sessionManager.InitUserAPI(client, ctx)
	if err != nil {
		return nil, err
	}
	f.mu.Lock()
	f.api = api
	f.mu.Unlock()
	return api, nil
}

func (f *AuthManagerImpl) RunApiChecker(ctx context.Context) {
	if f.apiChecker == nil {
		logger.GlobalLogger.Warn("API checker is nil, skipping")
		return
	}

	logger.GlobalLogger.Info("Starting API monitoring")

	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.GlobalLogger.Info("API monitoring stopped due to context cancellation")
				return
			case <-f.stopCh:
				logger.GlobalLogger.Info("API monitoring stopped due to stop signal")
				return
			case <-ticker.C:
				if err := f.apiChecker.Run(ctx); err != nil {
					logger.GlobalLogger.Errorf("API check failed: %v", err)
					if f.isCriticalError(err) {
						logger.GlobalLogger.Errorf("Critical API error detected, triggering reconnect: %v", err)
						select {
						case f.reconnect <- struct{}{}:
							f.stopCh <- struct{}{}
							logger.GlobalLogger.Info("Reconnect signal sent")
						default:
							logger.GlobalLogger.Warn("Reconnect channel is full")
						}
					}
				} else {
					logger.GlobalLogger.Debug("API check successful")
				}
			}
		}
	}()

	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		f.handleReconnectSignals(ctx)
	}()
}

func (f *AuthManagerImpl) isCriticalError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	return strings.Contains(errStr, "auth_key_unregistered") ||
		strings.Contains(errStr, "connection_not_inited") ||
		strings.Contains(errStr, "session_revoked")
}

func (f *AuthManagerImpl) handleReconnectSignals(ctx context.Context) {
	for {
		select {
		case <-f.reconnect:
			logger.GlobalLogger.Info("Processing reconnect signal")

			if f.monitor != nil {
				logger.GlobalLogger.Info("Pausing gift monitoring during reconnection")
				f.monitor.Pause()
			}

			if _, err := f.Reconnect(ctx); err != nil {
				logger.GlobalLogger.Errorf("Reconnect failed: %v", err)
				if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline") {
					logger.GlobalLogger.Error("Reconnection timeout, exiting program")
					os.Exit(1)
				}
			} else {
				if f.monitor != nil {
					logger.GlobalLogger.Info("Resuming gift monitoring after reconnection")
					f.monitor.Resume()
				}
				<-f.stopCh
				logger.GlobalLogger.Info("Reconnection completed successfully")
			}
		case <-f.stopCh:
			logger.GlobalLogger.Info("Stopping reconnect handler")
			return
		case <-ctx.Done():
			logger.GlobalLogger.Info("Context cancelled, stopping reconnect handler")
			return
		}
	}
}

func (f *AuthManagerImpl) InitBotClient(ctx context.Context) (*tg.Client, error) {
	if f.sessionManager == nil {
		return nil, errors.New("session manager is nil")
	}

	botApi, err := f.sessionManager.InitBotAPI(ctx)
	if err != nil {
		return nil, err
	}
	f.mu.Lock()
	f.botApi = botApi
	f.mu.Unlock()
	return botApi, nil
}

func (f *AuthManagerImpl) GetBotApi() *tg.Client {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.botApi
}

func (f *AuthManagerImpl) GetApi() *tg.Client {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.api
}

func (f *AuthManagerImpl) Reconnect(ctx context.Context) (*tg.Client, error) {
	logger.GlobalLogger.Info("Starting reconnection process")

	tgc, err := f.InitClient(ctx)
	if err != nil {
		return nil, err
	}

	if f.apiChecker != nil {
		newApiChecker := apiChecker.NewApiChecker(tgc, time.NewTicker(2*time.Second))
		f.SetApiChecker(newApiChecker)
		// logger.GlobalLogger.Info("API checker updated with new client")
	}

	logger.GlobalLogger.Info("Reconnection successful")
	return tgc, nil
}

func (f *AuthManagerImpl) Stop() {
	logger.GlobalLogger.Info("Stopping AuthManager...")

	close(f.stopCh)

	if f.apiChecker != nil {
		f.apiChecker.Stop()
	}

	close(f.reconnect)

	f.wg.Wait()

	logger.GlobalLogger.Info("AuthManager stopped")
}

func (f *AuthManagerImpl) SetApiChecker(apiChecker authInterfaces.ApiChecker) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.apiChecker = apiChecker
}

func (f *AuthManagerImpl) SetMonitor(monitor authInterfaces.GiftMonitorAndAuthController) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.monitor = monitor
	logger.GlobalLogger.Info("Gift monitor set for auth manager")
}
