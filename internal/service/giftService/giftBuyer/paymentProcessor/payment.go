package paymentProcessor

import (
	"context"
	"gift-buyer/internal/service/giftService/giftInterfaces"
	"gift-buyer/pkg/errors"
	"sync/atomic"
	"time"

	"github.com/gotd/td/tg"
)

type PaymentProcessorImpl struct {
	api            *tg.Client
	invoiceCreator giftInterfaces.InvoiceCreator
	rateLimiter    giftInterfaces.RateLimiter
	requestCounter int64
}

func NewPaymentProcessor(api *tg.Client, invoiceCreator giftInterfaces.InvoiceCreator, rateLimiter giftInterfaces.RateLimiter) *PaymentProcessorImpl {
	return &PaymentProcessorImpl{
		api:            api,
		invoiceCreator: invoiceCreator,
		rateLimiter:    rateLimiter,
	}
}

func (pp *PaymentProcessorImpl) CreatePaymentForm(ctx context.Context, gift *tg.StarGift) (tg.PaymentsPaymentFormClass, *tg.InputInvoiceStarGift, error) {
	jitter := time.Duration(atomic.AddInt64(&pp.requestCounter, 1)%100) * time.Millisecond
	time.Sleep(jitter)

	invoice, err := pp.invoiceCreator.CreateInvoice(gift)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create invoice")
	}

	if err := pp.rateLimiter.Acquire(ctx); err != nil {
		return nil, nil, errors.Wrap(err, "failed to wait for rate limit")
	}
	paymentFormRequest := &tg.PaymentsGetPaymentFormRequest{
		Invoice: invoice,
	}
	paymentForm, err := pp.api.PaymentsGetPaymentForm(ctx, paymentFormRequest)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get payment form")
	}

	return paymentForm, invoice, nil
}
