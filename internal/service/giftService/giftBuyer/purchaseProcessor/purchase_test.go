package purchaseProcessor

import (
	"context"
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPaymentProcessor для тестирования
type MockPaymentProcessor struct {
	mock.Mock
}

func (m *MockPaymentProcessor) CreatePaymentForm(ctx context.Context, gift *tg.StarGift) (tg.PaymentsPaymentFormClass, *tg.InputInvoiceStarGift, error) {
	args := m.Called(ctx, gift)
	if args.Get(0) == nil || args.Get(1) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(tg.PaymentsPaymentFormClass), args.Get(1).(*tg.InputInvoiceStarGift), args.Error(2)
}

func createTestGift(id int64, stars int64) *tg.StarGift {
	return &tg.StarGift{
		ID:    id,
		Stars: stars,
	}
}

func createTestInvoice(giftID int64) *tg.InputInvoiceStarGift {
	return &tg.InputInvoiceStarGift{
		Peer:   &tg.InputPeerSelf{},
		GiftID: giftID,
		Message: tg.TextWithEntities{
			Text: "Test message",
		},
	}
}

func TestNewPurchaseProcessor(t *testing.T) {
	mockPaymentProcessor := &MockPaymentProcessor{}

	processor := NewPurchaseProcessor(nil, mockPaymentProcessor)

	assert.NotNil(t, processor)
	assert.Nil(t, processor.api)
	assert.Equal(t, mockPaymentProcessor, processor.paymentProcessor)
}

func TestPurchaseProcessorImpl_PurchaseGift_ErrorCases(t *testing.T) {
	t.Run("ошибка при неподдерживаемом типе формы", func(t *testing.T) {
		mockPaymentProcessor := &MockPaymentProcessor{}

		processor := &PurchaseProcessorImpl{
			api:              nil,
			paymentProcessor: mockPaymentProcessor,
		}

		gift := createTestGift(1, 100)
		invoice := createTestInvoice(gift.ID)
		paymentForm := &tg.PaymentsPaymentForm{} // Неподдерживаемый тип

		// Настраиваем моки
		mockPaymentProcessor.On("CreatePaymentForm", mock.Anything, gift).Return(paymentForm, invoice, nil)

		ctx := context.Background()
		err := processor.PurchaseGift(ctx, gift)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "regular payment form not supported for star gifts")

		mockPaymentProcessor.AssertCalled(t, "CreatePaymentForm", ctx, gift)
	})

	t.Run("ошибка при создании payment form", func(t *testing.T) {
		mockPaymentProcessor := &MockPaymentProcessor{}

		processor := &PurchaseProcessorImpl{
			api:              nil,
			paymentProcessor: mockPaymentProcessor,
		}

		gift := createTestGift(1, 100)

		// Настраиваем мок для возврата ошибки
		mockPaymentProcessor.On("CreatePaymentForm", mock.Anything, gift).Return(nil, nil, assert.AnError)

		ctx := context.Background()
		err := processor.PurchaseGift(ctx, gift)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send stars form")

		mockPaymentProcessor.AssertCalled(t, "CreatePaymentForm", ctx, gift)
	})

	t.Run("неизвестный тип payment form", func(t *testing.T) {
		mockPaymentProcessor := &MockPaymentProcessor{}

		processor := &PurchaseProcessorImpl{
			api:              nil,
			paymentProcessor: mockPaymentProcessor,
		}

		gift := createTestGift(1, 100)
		invoice := createTestInvoice(gift.ID)

		// Используем nil для имитации неизвестного типа
		// Настраиваем моки
		mockPaymentProcessor.On("CreatePaymentForm", mock.Anything, gift).Return(nil, invoice, nil)

		ctx := context.Background()
		err := processor.PurchaseGift(ctx, gift)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected payment form type")

		mockPaymentProcessor.AssertCalled(t, "CreatePaymentForm", ctx, gift)
	})
}

func TestPurchaseProcessorImpl_SendStarsForm(t *testing.T) {
	t.Run("тестирование с nil API", func(t *testing.T) {
		processor := &PurchaseProcessorImpl{
			api:              nil,
			paymentProcessor: nil,
		}

		invoice := createTestInvoice(1)
		ctx := context.Background()

		// С nil API должна возникнуть паника или ошибка
		assert.Panics(t, func() {
			processor.sendStarsForm(ctx, invoice, 12345)
		})
	})
}

func TestPurchaseProcessorImpl_ValidatePurchase_NilAPI(t *testing.T) {
	t.Run("валидация с nil API", func(t *testing.T) {
		processor := &PurchaseProcessorImpl{
			api:              nil,
			paymentProcessor: nil,
		}

		gift := createTestGift(1, 100)

		// С nil API валидация должна вызвать панику, ловим её
		assert.Panics(t, func() {
			processor.validatePurchase(gift)
		})
	})
}
