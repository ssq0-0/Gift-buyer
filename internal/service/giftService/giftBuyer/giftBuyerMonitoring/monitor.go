package giftBuyerMonitoring

import (
	"context"
	"fmt"
	"gift-buyer/internal/service/giftService/giftInterfaces"
	"gift-buyer/internal/service/giftService/giftTypes"
	"gift-buyer/pkg/errors"
	"gift-buyer/pkg/logger"

	"github.com/gotd/td/tg"
)

type GiftBuyerMonitoringImpl struct {
	api          *tg.Client
	notification giftInterfaces.NotificationService
}

func NewGiftBuyerMonitoring(api *tg.Client, notification giftInterfaces.NotificationService) *GiftBuyerMonitoringImpl {
	return &GiftBuyerMonitoringImpl{
		api:          api,
		notification: notification,
	}
}

func (gm *GiftBuyerMonitoringImpl) MonitorProcess(ctx context.Context, resultsCh chan giftTypes.GiftResult, doneChan chan struct{}, gifts map[*tg.StarGift]int64) {
	summaries := make(map[int64]*giftTypes.GiftSummary)
	errorCounts := make(map[string]int64)
	for gift, count := range gifts {
		summaries[gift.ID] = &giftTypes.GiftSummary{
			GiftID:    gift.ID,
			Requested: count,
			Success:   0,
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-doneChan:
			mostFrequentError := gm.getMostFrequentError(errorCounts)
			gm.sendNotify(ctx, summaries, mostFrequentError)
			return
		case result, ok := <-resultsCh:
			if !ok {
				return
			}

			if result.Success {
				summaries[result.GiftID].Success++
			} else if result.Err != nil {
				errorCounts[result.Err.Error()]++
			}
		}
	}
}

func (gm *GiftBuyerMonitoringImpl) getMostFrequentError(errorCounts map[string]int64) error {
	if len(errorCounts) == 0 {
		return nil
	}

	var mostFrequentError string
	var maxCount int64

	for errorMsg, count := range errorCounts {
		if count > maxCount {
			maxCount = count
			mostFrequentError = errorMsg
		}
	}

	return errors.New(mostFrequentError)
}

func (gm *GiftBuyerMonitoringImpl) sendNotify(ctx context.Context, summaries map[int64]*giftTypes.GiftSummary, mostFrequentError error) {
	totalSuccess := int64(0)
	totalRequested := int64(0)

	for _, summary := range summaries {
		totalSuccess += summary.Success
		totalRequested += summary.Requested
	}

	if gm.notification.SetBot() {
		if totalSuccess == totalRequested {
			gm.notification.SendBuyStatus(ctx,
				fmt.Sprintf("✅ Успешно куплено %d подарков", totalSuccess), nil)
		} else if totalSuccess > 0 {
			message := fmt.Sprintf("⚠️ Частично выполнено: %d/%d подарков куплено",
				totalSuccess, totalRequested)
			gm.notification.SendBuyStatus(ctx, message, nil)
		} else {
			message := fmt.Sprintf("❌ Не удалось купить ни одного подарка из %d", totalRequested)
			errorToSend := mostFrequentError
			if errorToSend == nil {
				errorToSend = errors.New("все покупки неудачны")
			}
			gm.notification.SendBuyStatus(ctx, message, errorToSend)
		}
	} else {
		if totalSuccess == totalRequested {
			logger.GlobalLogger.Infof("✅ Successfully bought all %d gifts", totalSuccess)
		} else if totalSuccess > 0 {
			logger.GlobalLogger.Warnf("⚠️ Partially completed: %d/%d gifts bought", totalSuccess, totalRequested)
		} else {
			logger.GlobalLogger.Errorf("❌ Failed to buy any gifts out of %d requested", totalRequested)
		}

		for _, summary := range summaries {
			if summary.Success > 0 {
				logger.GlobalLogger.Infof("Successfully bought %d/%d x gift %d",
					summary.Success, summary.Requested, summary.GiftID)
			} else {
				logger.GlobalLogger.Errorf("Failed to buy %d/%d x gift %d",
					summary.Success, summary.Requested, summary.GiftID)
			}
		}
		if mostFrequentError != nil {
			logger.GlobalLogger.Errorf("Most frequent error during purchase: %v", mostFrequentError)
		}
	}
}
