package invoiceCreator

import (
	"context"
	"fmt"
	"gift-buyer/internal/service/giftService/giftInterfaces"
	"gift-buyer/pkg/errors"
	"gift-buyer/pkg/utils"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/gotd/td/tg"
)

type InvoiceCreatorImpl struct {
	userReceiver    []int
	channelReceiver []int
	receiverType    []int
	idCache         giftInterfaces.UserCache
}

func NewInvoiceCreator(userReceiver []int, channelReceiver []int, receiverType []int, idCache giftInterfaces.UserCache) *InvoiceCreatorImpl {
	return &InvoiceCreatorImpl{
		userReceiver:    userReceiver,
		channelReceiver: channelReceiver,
		receiverType:    receiverType,
		idCache:         idCache,
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
func (ic *InvoiceCreatorImpl) CreateInvoice(gift *tg.StarGift) (*tg.InputInvoiceStarGift, error) {
	randReceiverType := utils.SelectRandomElementFast(ic.receiverType)

	switch randReceiverType {
	case 0:
		return ic.selfPurchase(gift)
	case 1:
		return ic.userPurchase(gift)
	case 2:
		return ic.channelPurchase(gift)
	default:
		return nil, errors.Wrap(errors.New("unexpected receiver type"),
			fmt.Sprintf("unexpected receiver type: %d", randReceiverType))
	}
}

func (ic *InvoiceCreatorImpl) selfPurchase(gift *tg.StarGift) (*tg.InputInvoiceStarGift, error) {
	invoice := &tg.InputInvoiceStarGift{
		Peer:     &tg.InputPeerSelf{},
		GiftID:   gift.ID,
		HideName: true,
		Message: tg.TextWithEntities{
			Text: fmt.Sprintf("By @chiefssq %s_%d_%s", utils.RandString5(10), time.Now().UnixNano(), uuid.New().String()[:6]),
		},
	}
	return invoice, nil
}

func (ic *InvoiceCreatorImpl) userPurchase(gift *tg.StarGift) (*tg.InputInvoiceStarGift, error) {
	userInfo, err := ic.getUserInfo(context.Background(), int64(utils.SelectRandomElementFast(ic.userReceiver)))
	if err != nil {
		return nil, errors.Wrap(err, "cannot create invoice without user access hash")
	}

	invoice := &tg.InputInvoiceStarGift{
		Peer:     &tg.InputPeerUser{UserID: userInfo.ID, AccessHash: userInfo.AccessHash},
		GiftID:   gift.ID,
		HideName: rand.Intn(2) == 0,
		Message: tg.TextWithEntities{
			Text: fmt.Sprintf("By @chiefssq %s_%d_%s", utils.RandString5(10), time.Now().UnixNano(), uuid.New().String()[:6]),
		},
	}
	return invoice, nil
}

func (ic *InvoiceCreatorImpl) channelPurchase(gift *tg.StarGift) (*tg.InputInvoiceStarGift, error) {
	channelInfo, err := ic.getChannelInfo(context.Background(), int64(utils.SelectRandomElementFast(ic.channelReceiver)))
	if err != nil {
		return nil, errors.Wrap(err, "cannot create invoice without channel access hash")
	}

	invoice := &tg.InputInvoiceStarGift{
		Peer: &tg.InputPeerChannel{
			ChannelID:  channelInfo.ID,
			AccessHash: channelInfo.AccessHash,
		},
		GiftID:   gift.ID,
		HideName: true,
		Message: tg.TextWithEntities{
			Text: fmt.Sprintf("By @chiefssq %s_%d", utils.RandString5(10), time.Now().UnixNano()),
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
func (ic *InvoiceCreatorImpl) getChannelInfo(ctx context.Context, channelID int64) (*tg.Channel, error) {
	channel, err := ic.idCache.GetChannel(channelID)
	if err == nil {
		return channel, nil
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
func (ic *InvoiceCreatorImpl) getUserInfo(ctx context.Context, userID int64) (*tg.User, error) {
	user, err := ic.idCache.GetUser(userID)
	if err == nil {
		return user, nil
	}

	return nil, errors.New(fmt.Sprintf("user %d not accessible: session hasn't met this user. See logs for solutions.", userID))
}
