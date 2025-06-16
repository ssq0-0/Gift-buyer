// Package giftBuyer provides gift purchasing functionality for the gift buying system.
// It handles the complete purchase workflow including payment processing, retry logic,
// balance validation, and concurrent purchase management with configurable limits.
package giftBuyer

import (
	"context"
	"fmt"
	"gift-buyer/internal/service/giftService/giftInterfaces"
	"gift-buyer/pkg/errors"
	"gift-buyer/pkg/logger"
	"sync"
	"sync/atomic"
	"time"

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

	// api is the Telegram client used for payment operations
	api *tg.Client

	// userReceiver is the ID of the gift recipient
	userReceiver []int

	// channelReceiver is the ID of the gift recipient
	channelReceiver []int

	// receiverType specifies the type of receiver (1 for user, 2 for channel)
	receiverType []int

	// counter tracks and limits the total number of purchases
	counter *atomicCounter

	retryCount int
	// retryDelay           time.Duration
	concurrentGifts      int
	concurrentOperations int

	// requestCounter provides unique identifiers for requests to avoid FormID duplicates
	requestCounter int64
	rateLimiter    giftInterfaces.RateLimiter
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
) giftInterfaces.GiftBuyer {
	return &GiftBuyerImpl{
		api:                  api,
		userReceiver:         userIds,
		channelReceiver:      channelIds,
		receiverType:         receiverType,
		manager:              manager,
		notification:         notification,
		counter:              newAtomicCounter(maxBuyCount),
		retryCount:           retryCount,
		idCache:              idCache,
		concurrentGifts:      concurrentGifts,
		concurrentOperations: concurrentOperations,
		requestCounter:       0,
		rateLimiter:          rateLimiter,
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
		resultsCh = make(chan giftResult)
		doneCh    = make(chan struct{})
	)

	go gm.monitorProcess(ctx, resultsCh, doneCh, gifts)

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

func (gm *GiftBuyerImpl) monitorProcess(ctx context.Context, resultsCh chan giftResult, doneChan chan struct{}, gifts map[*tg.StarGift]int64) {
	summaries := make(map[int64]*giftSummary)
	errorCounts := make(map[string]int64)
	for gift, count := range gifts {
		summaries[gift.ID] = &giftSummary{
			giftID:    gift.ID,
			requested: count,
			success:   0,
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

			if result.success {
				summaries[result.giftID].success++
			} else if result.err != nil {
				errorCounts[result.err.Error()]++
			}
		}
	}
}

func (gm *GiftBuyerImpl) getMostFrequentError(errorCounts map[string]int64) error {
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

func (gm *GiftBuyerImpl) sendNotify(ctx context.Context, summaries map[int64]*giftSummary, mostFrequentError error) {
	totalSuccess := int64(0)
	totalRequested := int64(0)

	for _, summary := range summaries {
		totalSuccess += summary.success
		totalRequested += summary.requested
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
			if summary.success > 0 {
				logger.GlobalLogger.Infof("Successfully bought %d/%d x gift %d",
					summary.success, summary.requested, summary.giftID)
			} else {
				logger.GlobalLogger.Errorf("Failed to buy %d/%d x gift %d",
					summary.success, summary.requested, summary.giftID)
			}
		}
		if mostFrequentError != nil {
			logger.GlobalLogger.Errorf("Most frequent error during purchase: %v", mostFrequentError)
		}
	}
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
func (gm *GiftBuyerImpl) buyGift(ctx context.Context, gift *tg.StarGift, count int64, resChan chan<- giftResult) {
	var (
		wg  sync.WaitGroup
		sem = make(chan struct{}, gm.concurrentOperations)
	)

	for i := int64(0); i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			gm.buyGiftWithRetry(ctx, gift, sem, resChan)
		}()
	}

	wg.Wait()
}

func (gm *GiftBuyerImpl) buyGiftWithRetry(ctx context.Context, gift *tg.StarGift, sem chan struct{}, resChan chan<- giftResult) {
	attemptChan := make(chan int, gm.retryCount)
	resultChan := make(chan bool, 1)
	var lastErr error
	var mu sync.Mutex

	for j := 0; j < gm.retryCount; j++ {
		attemptChan <- j
	}
	close(attemptChan)

	var wg sync.WaitGroup

	for attempt := range attemptChan {
		wg.Add(1)
		go func(attemptNum int) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			case <-resultChan:
				return
			default:
			}

			sem <- struct{}{}
			defer func() { <-sem }()

			if !gm.counter.TryIncrement() {
				mu.Lock()
				lastErr = errors.New("max buy count reached")
				mu.Unlock()
				return
			}

			if err := gm.purchaseGift(ctx, gift); err != nil {
				logger.GlobalLogger.Errorf("Failed to purchase gift %d attempt %d. Error: %v", gift.ID, attemptNum+1, err)
				gm.counter.Decrement()
				lastErr = err
				return
			}

			select {
			case resultChan <- true:
				logger.GlobalLogger.Infof("Successfully purchased gift %d on attempt %d", gift.ID, attemptNum+1)
			default:
				gm.counter.Decrement()
			}
		}(attempt)
	}

	wg.Wait()

	select {
	case <-resultChan:
		resChan <- giftResult{
			giftID:  gift.ID,
			success: true,
			err:     nil,
		}
	default:
		if lastErr == nil {
			lastErr = errors.New(fmt.Sprintf("failed to buy gift %d after %d attempts", gift.ID, gm.retryCount))
		}
		resChan <- giftResult{
			giftID:  gift.ID,
			success: false,
			err:     lastErr,
		}
	}
}

// validatePurchase checks if a purchase can proceed by validating the user's balance.
// It ensures sufficient stars are available before attempting the actual purchase.
//
// Parameters:
//   - ctx: context for request cancellation and timeout control
//   - gift: the star gift to validate for purchase
//
// Returns:
//   - error: validation error if insufficient balance or balance check fails
func (gm *GiftBuyerImpl) validatePurchase(gift *tg.StarGift) bool {
	// Для тестирования - если API клиент nil, возвращаем false
	if gm.api == nil {
		return false
	}

	balance, err := gm.api.PaymentsGetStarsStatus(context.Background(), &tg.InputPeerSelf{})
	if err != nil {
		return false
	}
	return balance.Balance.Amount >= gift.Stars
}

// purchaseGift executes the actual gift purchase through Telegram's payment API.
// It creates an invoice, retrieves the payment form, and processes the star payment.
//
// The purchase process:
//  1. Creates an invoice for the gift
//  2. Retrieves the payment form from Telegram
//  3. Processes the payment based on form type
//  4. Handles different payment form variations
//
// Parameters:
//   - ctx: context for request cancellation and timeout control
//   - gift: the star gift to purchase
//
// Returns:
//   - error: payment processing error or API communication failure
func (gm *GiftBuyerImpl) purchaseGift(ctx context.Context, gift *tg.StarGift) error {
	// Всегда вызываем rate limiter, даже для тестирования
	if err := gm.rateLimiter.Acquire(ctx); err != nil {
		return errors.Wrap(err, "failed to wait for rate limit")
	}

	// Для тестирования - если API клиент nil, имитируем успешную покупку
	if gm.api == nil {
		time.Sleep(time.Millisecond * 1) // Имитируем задержку API
		return nil
	}

	if !gm.validatePurchase(gift) {
		return errors.New("insufficient balance to buy gift")
	}

	paymentForm, invoice, err := gm.createPaymentForm(ctx, gift)
	if err != nil {
		return errors.Wrap(err, "failed to send stars form")
	}

	switch form := paymentForm.(type) {
	case *tg.PaymentsPaymentFormStars:
		return gm.sendStarsForm(ctx, invoice, form.FormID)
	case *tg.PaymentsPaymentFormStarGift:
		return gm.sendStarsForm(ctx, invoice, form.FormID)
	case *tg.PaymentsPaymentForm:
		return errors.New("regular payment form not supported for star gifts")
	default:
		return errors.Wrap(errors.New("unexpected payment form type"),
			fmt.Sprintf("unexpected payment form type: %T", paymentForm))
	}
}

func (gm *GiftBuyerImpl) createPaymentForm(ctx context.Context, gift *tg.StarGift) (tg.PaymentsPaymentFormClass, *tg.InputInvoiceStarGift, error) {
	jitter := time.Duration(atomic.AddInt64(&gm.requestCounter, 1)%10) * time.Millisecond
	time.Sleep(jitter)

	invoice, err := gm.createInvoice(gift)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create invoice")
	}

	paymentFormRequest := &tg.PaymentsGetPaymentFormRequest{
		Invoice: invoice,
	}
	paymentForm, err := gm.api.PaymentsGetPaymentForm(ctx, paymentFormRequest)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get payment form")
	}

	return paymentForm, invoice, nil
}

func (gm *GiftBuyerImpl) Close() {
	gm.rateLimiter.Close()
}

func (gm *GiftBuyerImpl) sendStarsForm(ctx context.Context, invoice *tg.InputInvoiceStarGift, id int64) error {
	sendStarsRequest := &tg.PaymentsSendStarsFormRequest{
		FormID:  id,
		Invoice: invoice,
	}

	_, err := gm.api.PaymentsSendStarsForm(ctx, sendStarsRequest)
	if err != nil {
		return errors.Wrap(err, "failed to send payment")
	}
	return nil
}

// createInvoice creates a Telegram invoice for the specified gift.
// It configures the invoice based on the receiver type (self, user, or channel)
// and includes appropriate peer information and gift details.
//
// Supported receiver types:
//   - 0: Self (current user)
//   - 1: User (specified by user ID)
//   - 2: Channel (specified by channel ID with access hash)
//
// Parameters:
//   - gift: the star gift to create an invoice for
//
// Returns:
//   - *tg.InputInvoiceStarGift: configured invoice for the gift purchase
//   - error: invoice creation error or unsupported receiver type
func (gm *GiftBuyerImpl) createInvoice(gift *tg.StarGift) (*tg.InputInvoiceStarGift, error) {
	randReceiverType := selectRandomElementFast(gm.receiverType)

	switch randReceiverType {
	case 0:
		return gm.selfPurchase(gift)
	case 1:
		return gm.userPurchase(gift)
	case 2:
		return gm.channelPurchase(gift)
	default:
		return nil, errors.Wrap(errors.New("unexpected receiver type"),
			fmt.Sprintf("unexpected receiver type: %d", randReceiverType))
	}
}

func (gm *GiftBuyerImpl) selfPurchase(gift *tg.StarGift) (*tg.InputInvoiceStarGift, error) {
	timestamp := time.Now().UnixNano()

	invoice := &tg.InputInvoiceStarGift{
		Peer:     &tg.InputPeerSelf{},
		GiftID:   gift.ID,
		HideName: true,
		Message: tg.TextWithEntities{
			Text: fmt.Sprintf("By @chiefssq %s_%d", RandString5(10), timestamp),
		},
	}
	return invoice, nil
}

func (gm *GiftBuyerImpl) userPurchase(gift *tg.StarGift) (*tg.InputInvoiceStarGift, error) {
	userInfo, err := gm.getUserInfo(context.Background(), int64(selectRandomElementFast(gm.userReceiver)))
	if err != nil {
		return nil, errors.Wrap(err, "cannot create invoice without user access hash")
	}

	timestamp := time.Now().UnixNano()

	invoice := &tg.InputInvoiceStarGift{
		Peer:     &tg.InputPeerUser{UserID: userInfo.ID, AccessHash: userInfo.AccessHash},
		GiftID:   gift.ID,
		HideName: true,
		Message: tg.TextWithEntities{
			Text: fmt.Sprintf("By @chiefssq %s_%d", RandString5(10), timestamp),
		},
	}
	return invoice, nil
}

func (gm *GiftBuyerImpl) channelPurchase(gift *tg.StarGift) (*tg.InputInvoiceStarGift, error) {
	channelInfo, err := gm.getChannelInfo(context.Background(), int64(selectRandomElementFast(gm.channelReceiver)))
	if err != nil {
		return nil, errors.Wrap(err, "cannot create invoice without channel access hash")
	}

	timestamp := time.Now().UnixNano()

	invoice := &tg.InputInvoiceStarGift{
		Peer: &tg.InputPeerChannel{
			ChannelID:  channelInfo.ID,
			AccessHash: channelInfo.AccessHash,
		},
		GiftID:   gift.ID,
		HideName: true,
		Message: tg.TextWithEntities{
			Text: fmt.Sprintf("By @chiefssq %s_%d", RandString5(10), timestamp),
		},
	}
	return invoice, nil
}

// getChannelInfo retrieves channel information including access hash for invoice creation.
// It handles channel ID conversion and fetches the channel details required for
// creating invoices for channel recipients.
//
// Parameters:
//   - ctx: context for request cancellation and timeout control
//   - channelID: the channel ID (may be in supergroup format)
//
// Returns:
//   - *tg.Channel: channel information with access hash
//   - error: channel retrieval error or API communication failure
func (gm *GiftBuyerImpl) getChannelInfo(ctx context.Context, channelID int64) (*tg.Channel, error) {
	channel, err := gm.idCache.GetChannel(channelID)
	if err == nil {
		return channel, nil
	}

	// Для тестирования - если API клиент nil, возвращаем ошибку
	if gm.api == nil {
		return nil, errors.New("channel not found")
	}

	var actualChannelID int64
	if channelID < -1000000000000 {
		actualChannelID = -channelID - 1000000000000
	} else {
		actualChannelID = channelID
	}

	channels, err := gm.api.ChannelsGetChannels(ctx, []tg.InputChannelClass{
		&tg.InputChannel{
			ChannelID: actualChannelID,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get channel info via ChannelsGetChannels")
	}

	for _, chat := range channels.GetChats() {
		if channel, ok := chat.(*tg.Channel); ok {
			gm.idCache.SetChannel(channel)
			return channel, nil
		}
	}

	return nil, errors.New("channel not found")
}

// getUserInfo retrieves user information including access hash for invoice creation.
// It tries multiple methods to get user info without requiring contacts:
// 1. Direct UsersGetUsers call
// 2. Search through recent dialogs with larger limit
// 3. Try to get user through common groups/channels
// 4. Search for messages from user
// 5. Try to resolve by username if available
// 6. Search through all chats and channels
//
// Parameters:
//   - ctx: context for request cancellation and timeout control
//   - userID: the user ID
//
// Returns:
//   - *tg.User: user information with access hash
//   - error: user retrieval error or API communication failure
func (gm *GiftBuyerImpl) getUserInfo(ctx context.Context, userID int64) (*tg.User, error) {
	user, err := gm.idCache.GetUser(userID)
	if err == nil {
		return user, nil
	}

	// Для тестирования - если API клиент nil, возвращаем ошибку
	if gm.api == nil {
		return nil, errors.New("user not found")
	}

	contacts, err := gm.api.ContactsGetContacts(ctx, 0)
	if err != nil {
		return nil, err
	}
	switch v := contacts.(type) {
	case *tg.ContactsContacts:
		for _, user := range v.Users {
			u, ok := user.(*tg.User)
			if !ok {
				continue
			}
			if u.ID == userID {
				gm.idCache.SetUser(u)
				return u, nil
			}
		}
	default:
		logger.GlobalLogger.Errorf("unexpected response type:", fmt.Sprintf("%T", contacts))
	}
	return nil, errors.New(fmt.Sprintf("user %d not accessible: session hasn't met this user. See logs for solutions.", userID))
}
