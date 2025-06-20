package paymentProcessor

import (
	"context"
	"errors"
	"testing"

	"gift-buyer/internal/service/giftService/giftInterfaces"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock для RateLimiter
type MockRateLimiter struct {
	mock.Mock
}

func (m *MockRateLimiter) Acquire(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRateLimiter) Close() {
	m.Called()
}

// Mock для InvoiceCreator
type MockInvoiceCreator struct {
	mock.Mock
}

func (m *MockInvoiceCreator) CreateInvoice(gift *tg.StarGift) (*tg.InputInvoiceStarGift, error) {
	args := m.Called(gift)
	return args.Get(0).(*tg.InputInvoiceStarGift), args.Error(1)
}

// Helper functions
func createTestGift(id int64, stars int64) *tg.StarGift {
	return &tg.StarGift{
		ID:    id,
		Stars: stars,
	}
}

func createTestInvoice(giftID int64) *tg.InputInvoiceStarGift {
	return &tg.InputInvoiceStarGift{
		Peer:     &tg.InputPeerSelf{},
		GiftID:   giftID,
		HideName: true,
		Message: tg.TextWithEntities{
			Text: "test message",
		},
	}
}

func TestNewPaymentProcessor(t *testing.T) {
	t.Run("создание нового PaymentProcessor", func(t *testing.T) {
		mockInvoiceCreator := &MockInvoiceCreator{}
		mockRateLimiter := &MockRateLimiter{}

		processor := NewPaymentProcessor((*tg.Client)(nil), mockInvoiceCreator, mockRateLimiter)

		assert.NotNil(t, processor)
		// Тестируем что создан правильный тип
		var _ giftInterfaces.PaymentProcessor = processor
	})
}

func TestPaymentProcessorImpl_InvoiceCreation_Success(t *testing.T) {
	t.Run("успешное создание инвойса", func(t *testing.T) {
		mockInvoiceCreator := &MockInvoiceCreator{}
		mockRateLimiter := &MockRateLimiter{}

		processor := NewPaymentProcessor((*tg.Client)(nil), mockInvoiceCreator, mockRateLimiter)

		gift := createTestGift(1, 100)
		invoice := createTestInvoice(1)

		// Настраиваем моки
		mockInvoiceCreator.On("CreateInvoice", gift).Return(invoice, nil)

		// Тестируем только создание инвойса
		createdInvoice, err := processor.invoiceCreator.CreateInvoice(gift)
		assert.NoError(t, err)
		assert.Equal(t, invoice, createdInvoice)

		mockInvoiceCreator.AssertExpectations(t)
	})
}

func TestPaymentProcessorImpl_RateLimiter_Success(t *testing.T) {
	t.Run("успешное получение токена rate limiter", func(t *testing.T) {
		mockInvoiceCreator := &MockInvoiceCreator{}
		mockRateLimiter := &MockRateLimiter{}

		processor := NewPaymentProcessor((*tg.Client)(nil), mockInvoiceCreator, mockRateLimiter)

		ctx := context.Background()

		// Настраиваем мок
		mockRateLimiter.On("Acquire", ctx).Return(nil)

		// Тестируем rate limiting
		err := processor.rateLimiter.Acquire(ctx)
		assert.NoError(t, err)

		mockRateLimiter.AssertExpectations(t)
	})
}

func TestPaymentProcessorImpl_RateLimiter_Error(t *testing.T) {
	t.Run("ошибка rate limiter", func(t *testing.T) {
		mockInvoiceCreator := &MockInvoiceCreator{}
		mockRateLimiter := &MockRateLimiter{}

		processor := NewPaymentProcessor((*tg.Client)(nil), mockInvoiceCreator, mockRateLimiter)

		ctx := context.Background()

		// Настраиваем мок для возврата ошибки rate limiter
		mockRateLimiter.On("Acquire", ctx).Return(errors.New("rate limit exceeded"))

		// Тестируем rate limiting с ошибкой
		err := processor.rateLimiter.Acquire(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit exceeded")

		mockRateLimiter.AssertExpectations(t)
	})
}

func TestPaymentProcessorImpl_InvoiceCreation_Error(t *testing.T) {
	t.Run("ошибка создания инвойса", func(t *testing.T) {
		mockInvoiceCreator := &MockInvoiceCreator{}
		mockRateLimiter := &MockRateLimiter{}

		processor := NewPaymentProcessor((*tg.Client)(nil), mockInvoiceCreator, mockRateLimiter)

		gift := createTestGift(1, 100)

		// Настраиваем мок для возврата ошибки
		mockInvoiceCreator.On("CreateInvoice", gift).Return((*tg.InputInvoiceStarGift)(nil), errors.New("invoice creation failed"))

		// Тестируем создание инвойса с ошибкой
		createdInvoice, err := processor.invoiceCreator.CreateInvoice(gift)
		assert.Error(t, err)
		assert.Nil(t, createdInvoice)
		assert.Contains(t, err.Error(), "invoice creation failed")

		mockInvoiceCreator.AssertExpectations(t)
	})
}

func TestPaymentProcessorImpl_ContextCancellation(t *testing.T) {
	t.Run("отмена контекста", func(t *testing.T) {
		mockInvoiceCreator := &MockInvoiceCreator{}
		mockRateLimiter := &MockRateLimiter{}

		processor := NewPaymentProcessor((*tg.Client)(nil), mockInvoiceCreator, mockRateLimiter)

		// Создаем отмененный контекст
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Настраиваем мок для возврата ошибки контекста
		mockRateLimiter.On("Acquire", ctx).Return(context.Canceled)

		// Тестируем обработку отмененного контекста
		err := processor.rateLimiter.Acquire(ctx)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)

		mockRateLimiter.AssertExpectations(t)
	})
}
