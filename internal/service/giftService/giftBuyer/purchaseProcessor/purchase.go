package purchaseProcessor

import (
	"context"
	"fmt"
	"gift-buyer/internal/service/giftService/giftInterfaces"
	"gift-buyer/pkg/errors"

	"github.com/gotd/td/tg"
)

type PurchaseProcessorImpl struct {
	api              *tg.Client
	paymentProcessor giftInterfaces.PaymentProcessor
}

func NewPurchaseProcessor(api *tg.Client, paymentProcessor giftInterfaces.PaymentProcessor) *PurchaseProcessorImpl {
	return &PurchaseProcessorImpl{
		api:              api,
		paymentProcessor: paymentProcessor,
	}
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
func (pp *PurchaseProcessorImpl) PurchaseGift(ctx context.Context, gift *tg.StarGift) error {
	if !pp.validatePurchase(gift) {
		return errors.New("insufficient balance to buy gift")
	}

	paymentForm, invoice, err := pp.paymentProcessor.CreatePaymentForm(ctx, gift)
	if err != nil {
		return errors.Wrap(err, "failed to send stars form")
	}

	switch form := paymentForm.(type) {
	case *tg.PaymentsPaymentFormStars:
		return pp.sendStarsForm(ctx, invoice, form.FormID)
	case *tg.PaymentsPaymentFormStarGift:
		return pp.sendStarsForm(ctx, invoice, form.FormID)
	case *tg.PaymentsPaymentForm:
		return errors.New("regular payment form not supported for star gifts")
	default:
		return errors.Wrap(errors.New("unexpected payment form type"),
			fmt.Sprintf("unexpected payment form type: %T", paymentForm))
	}
}

func (pp *PurchaseProcessorImpl) sendStarsForm(ctx context.Context, invoice *tg.InputInvoiceStarGift, id int64) error {
	sendStarsRequest := &tg.PaymentsSendStarsFormRequest{
		FormID:  id,
		Invoice: invoice,
	}

	_, err := pp.api.PaymentsSendStarsForm(ctx, sendStarsRequest)
	if err != nil {
		return errors.Wrap(err, "failed to send payment")
	}
	return nil
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
func (pp *PurchaseProcessorImpl) validatePurchase(gift *tg.StarGift) bool {
	balance, err := pp.api.PaymentsGetStarsStatus(context.Background(), &tg.InputPeerSelf{})
	if err != nil {
		return false
	}
	return balance.Balance.Amount >= gift.Stars
}
