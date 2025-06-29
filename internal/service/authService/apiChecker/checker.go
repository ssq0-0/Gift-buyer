package apiChecker

import (
	"context"
	"errors"
	"gift-buyer/pkg/logger"
	"time"

	"github.com/gotd/td/tg"
)

type apiCheckerImpl struct {
	api    *tg.Client
	ticker *time.Ticker
}

func NewApiChecker(api *tg.Client, ticker *time.Ticker) *apiCheckerImpl {
	return &apiCheckerImpl{
		api:    api,
		ticker: ticker,
	}
}

func (f *apiCheckerImpl) Run(ctx context.Context) error {
	return f.ping(ctx)
}

func (f *apiCheckerImpl) ping(ctx context.Context) error {
	if f.api == nil {
		return errors.New("API client is nil")
	}

	api, err := f.api.AccountGetAuthorizations(ctx)
	if err != nil {
		return err
	}
	if len(api.Authorizations) == 0 {
		return errors.New("no authorizations found")
	}
	return nil
}

func (f *apiCheckerImpl) Stop() {
	if f.ticker != nil {
		f.ticker.Stop()
		logger.GlobalLogger.Info("API checker ticker stopped")
	}
}
