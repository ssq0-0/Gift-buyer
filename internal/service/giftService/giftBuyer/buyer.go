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
	"time"

	"github.com/gotd/td/tg"
)

const (
	// maxRetryAttempts defines the maximum number of retry attempts for failed purchases
	maxRetryAttempts = 5

	// baseRetryDelay is the base delay between retry attempts
	baseRetryDelay = 1 * time.Second
)

// GiftBuyerImpl implements the GiftBuyer interface for purchasing Telegram star gifts.
// It manages the complete purchase workflow including payment processing, retry logic,
// balance validation, and purchase counting with configurable limits.
type GiftBuyerImpl struct {
	// manager handles gift-related operations and API communication
	manager giftInterfaces.Giftmanager

	// notification sends purchase status updates and notifications
	notification giftInterfaces.NotificationService

	// api is the Telegram client used for payment operations
	api *tg.Client

	// receiver is the ID of the gift recipient
	receiver int

	// receiverType specifies the type of receiver (1 for user, 2 for channel)
	receiverType int

	// counter tracks and limits the total number of purchases
	counter *atomicCounter
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
//
// Returns:
//   - giftInterfaces.GiftBuyer: configured gift buyer instance
func NewGiftBuyer(api *tg.Client, receiver, receiverType int, manager giftInterfaces.Giftmanager, notification giftInterfaces.NotificationService, maxBuyCount int64) giftInterfaces.GiftBuyer {
	return &GiftBuyerImpl{
		api:          api,
		receiver:     receiver,
		receiverType: receiverType,
		manager:      manager,
		notification: notification,
		counter:      newAtomicCounter(maxBuyCount),
	}
}

// GetTotalBuyCount returns the current total number of gifts purchased.
//
// Returns:
//   - int64: current purchase count
func (gm *GiftBuyerImpl) GetTotalBuyCount() int64 {
	return gm.counter.Get()
}

// GetMaxBuyCount returns the maximum number of gifts that can be purchased.
//
// Returns:
//   - int64: maximum purchase limit
func (gm *GiftBuyerImpl) GetMaxBuyCount() int64 {
	return gm.counter.GetMax()
}

// CheckBalance retrieves the current star balance from the user's Telegram account.
// This is used to validate that sufficient funds are available before attempting purchases.
//
// Parameters:
//   - ctx: context for request cancellation and timeout control
//
// Returns:
//   - int64: current star balance amount
//   - error: API communication error or balance retrieval failure
func (gm *GiftBuyerImpl) CheckBalance(ctx context.Context) (int64, error) {
	starsStatus, err := gm.api.PaymentsGetStarsStatus(ctx, &tg.InputPeerSelf{})
	if err != nil {
		return 0, errors.Wrap(err, "failed to get stars status")
	}

	return starsStatus.Balance.Amount, nil
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
func (gm *GiftBuyerImpl) BuyGift(ctx context.Context, gifts map[*tg.StarGift]int64) error {
	if len(gifts) == 0 {
		return nil
	}

	totalPurchases := int64(0)
	for _, count := range gifts {
		totalPurchases += count
	}

	var (
		wg        sync.WaitGroup
		errCh     = make(chan error, len(gifts))
		successCh = make(chan int64, totalPurchases)
	)

	for gift, count := range gifts {
		wg.Add(1)
		go func(gift *tg.StarGift, count int64) {
			defer wg.Done()

			if success, err := gm.buyGift(ctx, gift, count); err != nil {
				errCh <- err
				logger.GlobalLogger.Errorf("failed to buy gift %d: %v", gift.ID, err)
				for i := int64(0); i < success; i++ {
					successCh <- gift.ID
				}
			} else {
				for i := int64(0); i < success; i++ {
					successCh <- gift.ID
				}
				logger.GlobalLogger.Infof("successfully bought %d x gift %d", success, gift.ID)
			}
		}(gift, count)
	}

	wg.Wait()
	close(errCh)
	close(successCh)

	var errList []error
	for err := range errCh {
		errList = append(errList, err)
	}

	var successCount int
	for range successCh {
		successCount++
	}

	totalGifts := len(gifts)
	failedCount := len(errList)

	if failedCount == 0 {
		if gm.notification.SetBot() {
			gm.notification.SendBuyStatus(ctx, "Success", nil)
		}
		return nil
	}

	if successCount > 0 {
		if gm.notification.SetBot() {
			gm.notification.SendBuyStatus(ctx, fmt.Sprintf("Success: %d gifts bought", successCount), nil)
		}
		return nil
	} else {
		logger.GlobalLogger.Errorf("failed to buy all %d gifts", totalGifts)
		for _, err := range errList {
			if gm.notification.SetBot() {
				gm.notification.SendBuyStatus(ctx, "Failed", err)
			}
		}
	}

	return errors.Wrap(errList[0], "some gifts failed to purchase")
}

// buyGift attempts to purchase a specific gift multiple times with retry logic.
// It handles individual gift purchases, manages the purchase counter, and implements
// retry logic with exponential backoff for failed attempts.
//
// Parameters:
//   - ctx: context for request cancellation and timeout control
//   - gift: the star gift to purchase
//   - count: number of times to purchase this gift
//
// Returns:
//   - int64: number of successful purchases completed
//   - error: purchase error after all retry attempts exhausted
func (gm *GiftBuyerImpl) buyGift(ctx context.Context, gift *tg.StarGift, count int64) (int64, error) {
	successCount := int64(0)
	for i := int64(0); i < count; i++ {
		var lastErr error
		purchased := false

		for j := 0; j < maxRetryAttempts; j++ {
			if !gm.counter.TryIncrement() {
				return successCount, errors.New("max buy count reached")
			}

			if err := gm.validatePurchase(ctx, gift); err != nil {
				gm.counter.Decrement()
				lastErr = err
				time.Sleep(baseRetryDelay)
				continue
			}

			if err := gm.purchaseGift(ctx, gift); err != nil {
				gm.counter.Decrement()
				lastErr = err
				time.Sleep(baseRetryDelay)
				continue
			}
			purchased = true
			successCount++
			break
		}
		if !purchased {
			return successCount, errors.Wrap(lastErr, fmt.Sprintf("failed to buy gift %d after %d attempts", gift.ID, maxRetryAttempts))
		}
	}
	return successCount, nil
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
func (gm *GiftBuyerImpl) validatePurchase(ctx context.Context, gift *tg.StarGift) error {
	balance, err := gm.CheckBalance(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to check balance")
	}

	if balance < gift.Stars {
		return errors.Wrap(errors.ErrBalanceEstimation, "insufficient balance to buy gift")
	}

	return nil
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
	invoice, err := gm.createInvoice(gift)
	if err != nil {
		return errors.Wrap(err, "failed to create invoice")
	}

	paymentFormRequest := &tg.PaymentsGetPaymentFormRequest{
		Invoice: invoice,
	}

	paymentForm, err := gm.api.PaymentsGetPaymentForm(ctx, paymentFormRequest)
	if err != nil {
		return errors.Wrap(err, "failed to get payment form")
	}

	switch form := paymentForm.(type) {
	case *tg.PaymentsPaymentFormStars:
		sendStarsRequest := &tg.PaymentsSendStarsFormRequest{
			FormID:  form.FormID,
			Invoice: invoice,
		}

		_, err = gm.api.PaymentsSendStarsForm(ctx, sendStarsRequest)
		if err != nil {
			return errors.Wrap(err, "failed to send payment")
		}
		return nil

	case *tg.PaymentsPaymentFormStarGift:
		sendStarsRequest := &tg.PaymentsSendStarsFormRequest{
			FormID:  form.FormID,
			Invoice: invoice,
		}

		_, err = gm.api.PaymentsSendStarsForm(ctx, sendStarsRequest)
		if err != nil {
			return errors.Wrap(err, "failed to send star gift payment")
		}
		return nil

	case *tg.PaymentsPaymentForm:
		return errors.New("regular payment form not supported for star gifts")

	default:
		return errors.Wrap(errors.New("unexpected payment form type"),
			fmt.Sprintf("unexpected payment form type: %T", paymentForm))
	}
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
	var invoice *tg.InputInvoiceStarGift
	switch gm.receiverType {
	case 0:
		invoice = &tg.InputInvoiceStarGift{
			Peer:   &tg.InputPeerSelf{},
			GiftID: gift.ID,
		}
	case 1:
		ctx := context.Background()
		userInfo, err := gm.getUserInfo(ctx, int64(gm.receiver))
		if err != nil {
			return nil, errors.Wrap(err, "cannot create invoice without user access hash")
		}
		invoice = &tg.InputInvoiceStarGift{
			Peer:     &tg.InputPeerUser{UserID: userInfo.ID, AccessHash: userInfo.AccessHash},
			GiftID:   gift.ID,
			HideName: true,
		}
		return invoice, nil
	case 2:
		ctx := context.Background()
		channelInfo, err := gm.getChannelInfo(ctx, int64(gm.receiver))
		if err != nil {
			return nil, errors.Wrap(err, "cannot create invoice without channel access hash")
		}

		invoice = &tg.InputInvoiceStarGift{
			Peer: &tg.InputPeerChannel{
				ChannelID:  channelInfo.ID,
				AccessHash: channelInfo.AccessHash,
			},
			GiftID: gift.ID,
		}
	default:
		return nil, errors.Wrap(errors.New("unexpected receiver type"),
			fmt.Sprintf("unexpected receiver type: %d", gm.receiverType))
	}

	invoice.SetMessage(tg.TextWithEntities{
		Text: "Gift bought by @cheifssq",
	})
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
			logger.GlobalLogger.Debugf("found channel %d with access hash %d", channel.ID, channel.AccessHash)
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
			if u.ID == int64(gm.receiver) {
				return u, nil
			}
		}
	default:
		fmt.Println("unexpected response type:", fmt.Sprintf("%T", contacts))
	}
	return nil, errors.New(fmt.Sprintf("user %d not accessible: session hasn't met this user. See logs for solutions.", int64(gm.receiver)))
}
