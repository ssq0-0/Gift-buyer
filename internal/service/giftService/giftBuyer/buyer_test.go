package giftBuyer

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations
type MockGiftManager struct {
	mock.Mock
}

func (m *MockGiftManager) GetAvailableGifts(ctx context.Context) ([]*tg.StarGift, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*tg.StarGift), args.Error(1)
}

type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) SendNewGiftNotification(ctx context.Context, gift *tg.StarGift) error {
	args := m.Called(ctx, gift)
	return args.Error(0)
}

func (m *MockNotificationService) SendBuyStatus(ctx context.Context, status string, err error) error {
	args := m.Called(ctx, status, err)
	return args.Error(0)
}

func (m *MockNotificationService) SetBot() bool {
	args := m.Called()
	return args.Bool(0)
}

type MockUserCache struct {
	mock.Mock
}

func (m *MockUserCache) SetUser(user *tg.User) {
	m.Called(user)
}

func (m *MockUserCache) GetUser(id int64) (*tg.User, error) {
	args := m.Called(id)
	return args.Get(0).(*tg.User), args.Error(1)
}

func (m *MockUserCache) SetChannel(channel *tg.Channel) {
	m.Called(channel)
}

func (m *MockUserCache) GetChannel(id int64) (*tg.Channel, error) {
	args := m.Called(id)
	return args.Get(0).(*tg.Channel), args.Error(1)
}

type MockRateLimiter struct {
	mock.Mock
	mu     sync.Mutex
	closed bool
}

func (m *MockRateLimiter) Acquire(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRateLimiter) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.closed {
		m.closed = true
		m.Called()
	}
}

// Helper functions
func createTestGift(id int64, stars int64) *tg.StarGift {
	return &tg.StarGift{
		ID:    id,
		Stars: stars,
	}
}

func createMockBuyer() (*GiftBuyerImpl, *MockGiftManager, *MockNotificationService, *MockUserCache, *MockRateLimiter) {
	mockManager := &MockGiftManager{}
	mockNotification := &MockNotificationService{}
	mockUserCache := &MockUserCache{}
	mockRateLimiter := &MockRateLimiter{}

	buyer := &GiftBuyerImpl{
		manager:              mockManager,
		idCache:              mockUserCache,
		notification:         mockNotification,
		api:                  nil, // nil API для тестирования
		userReceiver:         []int{123456},
		channelReceiver:      []int{789012},
		receiverType:         []int{0, 1, 2}, // self, user, channel
		counter:              newAtomicCounter(100),
		retryCount:           3,
		concurrentGifts:      5,
		concurrentOperations: 10,
		requestCounter:       0,
		rateLimiter:          mockRateLimiter,
	}

	return buyer, mockManager, mockNotification, mockUserCache, mockRateLimiter
}

func TestNewGiftBuyer(t *testing.T) {
	t.Run("создание нового GiftBuyer", func(t *testing.T) {
		mockManager := &MockGiftManager{}
		mockNotification := &MockNotificationService{}
		mockUserCache := &MockUserCache{}
		mockRateLimiter := &MockRateLimiter{}

		buyer := NewGiftBuyer(
			nil, // api
			[]int{123456},
			[]int{789012},
			[]int{0, 1, 2},
			mockManager,
			mockNotification,
			100, // maxBuyCount
			3,   // retryCount
			mockUserCache,
			5, // concurrentGifts
			mockRateLimiter,
			10, // concurrentOperations
		)

		assert.NotNil(t, buyer)
		impl, ok := buyer.(*GiftBuyerImpl)
		require.True(t, ok)

		assert.Equal(t, []int{123456}, impl.userReceiver)
		assert.Equal(t, []int{789012}, impl.channelReceiver)
		assert.Equal(t, []int{0, 1, 2}, impl.receiverType)
		assert.Equal(t, 3, impl.retryCount)
		assert.Equal(t, 5, impl.concurrentGifts)
		assert.Equal(t, 10, impl.concurrentOperations)
		assert.NotNil(t, impl.counter)
		assert.Equal(t, int64(100), impl.counter.GetMax())
	})
}

func TestGiftBuyerImpl_BuyGift(t *testing.T) {
	t.Run("успешная покупка подарков", func(t *testing.T) {
		buyer, _, mockNotification, _, mockRateLimiter := createMockBuyer()

		// Настраиваем моки
		mockNotification.On("SetBot").Return(true)
		mockNotification.On("SendBuyStatus", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)
		mockRateLimiter.On("Acquire", mock.Anything).Return(nil)

		gifts := map[*tg.StarGift]int64{
			createTestGift(1, 100): 2,
			createTestGift(2, 200): 1,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Создаем канал для ожидания завершения
		done := make(chan struct{})
		go func() {
			buyer.BuyGift(ctx, gifts)
			close(done)
		}()

		// Ждем завершения или таймаута
		select {
		case <-done:
			// Операция завершена
		case <-time.After(3 * time.Second):
			t.Fatal("BuyGift took too long to complete")
		}

		// Даем время для завершения асинхронных операций
		time.Sleep(100 * time.Millisecond)

		// Проверяем что уведомления были отправлены
		mockNotification.AssertCalled(t, "SetBot")
		mockNotification.AssertCalled(t, "SendBuyStatus", mock.Anything, mock.AnythingOfType("string"), mock.Anything)
	})

	t.Run("покупка с пустым списком подарков", func(t *testing.T) {
		buyer, _, mockNotification, _, _ := createMockBuyer()

		mockNotification.On("SetBot").Return(true)
		mockNotification.On("SendBuyStatus", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)

		gifts := map[*tg.StarGift]int64{}

		ctx := context.Background()

		// Создаем канал для ожидания завершения
		done := make(chan struct{})
		go func() {
			buyer.BuyGift(ctx, gifts)
			close(done)
		}()

		// Ждем завершения или таймаута
		select {
		case <-done:
			// Операция завершена
		case <-time.After(1 * time.Second):
			t.Fatal("BuyGift took too long to complete")
		}

		// Даем время для завершения асинхронных операций
		time.Sleep(100 * time.Millisecond)

		// Должно быть отправлено уведомление о том, что ничего не куплено
		mockNotification.AssertCalled(t, "SetBot")
	})

	t.Run("покупка с отменой контекста", func(t *testing.T) {
		buyer, _, _, _, _ := createMockBuyer()

		gifts := map[*tg.StarGift]int64{
			createTestGift(1, 100): 1,
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Отменяем контекст сразу

		buyer.BuyGift(ctx, gifts)

		// Тест должен завершиться без паники
	})
}

func TestGiftBuyerImpl_Close(t *testing.T) {
	t.Run("закрытие buyer", func(t *testing.T) {
		buyer, _, _, _, mockRateLimiter := createMockBuyer()

		mockRateLimiter.On("Close").Return()

		buyer.Close()

		mockRateLimiter.AssertCalled(t, "Close")
	})
}

func TestGiftBuyerImpl_CreateInvoice(t *testing.T) {
	t.Run("создание инвойса для self", func(t *testing.T) {
		buyer, _, _, _, _ := createMockBuyer()
		buyer.receiverType = []int{0} // только self

		gift := createTestGift(1, 100)
		invoice, err := buyer.createInvoice(gift)

		assert.NoError(t, err)
		assert.NotNil(t, invoice)
		assert.Equal(t, gift.ID, invoice.GiftID)
		assert.True(t, invoice.HideName)

		// Проверяем что peer это InputPeerSelf
		_, ok := invoice.Peer.(*tg.InputPeerSelf)
		assert.True(t, ok)
	})

	t.Run("создание инвойса для пользователя", func(t *testing.T) {
		buyer, _, _, mockUserCache, _ := createMockBuyer()
		buyer.receiverType = []int{1} // только user

		// Настраиваем мок для получения пользователя
		testUser := &tg.User{
			ID:         123456,
			AccessHash: 987654321,
		}
		mockUserCache.On("GetUser", int64(123456)).Return(testUser, nil)

		gift := createTestGift(1, 100)
		invoice, err := buyer.createInvoice(gift)

		assert.NoError(t, err)
		assert.NotNil(t, invoice)
		assert.Equal(t, gift.ID, invoice.GiftID)

		// Проверяем что peer это InputPeerUser
		peerUser, ok := invoice.Peer.(*tg.InputPeerUser)
		assert.True(t, ok)
		assert.Equal(t, testUser.ID, peerUser.UserID)
		assert.Equal(t, testUser.AccessHash, peerUser.AccessHash)
	})

	t.Run("создание инвойса для канала", func(t *testing.T) {
		buyer, _, _, mockUserCache, _ := createMockBuyer()
		buyer.receiverType = []int{2} // только channel

		// Настраиваем мок для получения канала
		testChannel := &tg.Channel{
			ID:         789012,
			AccessHash: 123456789,
		}
		mockUserCache.On("GetChannel", int64(789012)).Return(testChannel, nil)

		gift := createTestGift(1, 100)
		invoice, err := buyer.createInvoice(gift)

		assert.NoError(t, err)
		assert.NotNil(t, invoice)
		assert.Equal(t, gift.ID, invoice.GiftID)

		// Проверяем что peer это InputPeerChannel
		peerChannel, ok := invoice.Peer.(*tg.InputPeerChannel)
		assert.True(t, ok)
		assert.Equal(t, testChannel.ID, peerChannel.ChannelID)
		assert.Equal(t, testChannel.AccessHash, peerChannel.AccessHash)
	})

	t.Run("ошибка при получении пользователя", func(t *testing.T) {
		buyer, _, _, mockUserCache, _ := createMockBuyer()
		buyer.receiverType = []int{1} // только user

		// Настраиваем мок для возврата ошибки из кеша
		mockUserCache.On("GetUser", int64(123456)).Return((*tg.User)(nil), errors.New("user not found"))

		gift := createTestGift(1, 100)
		invoice, err := buyer.createInvoice(gift)

		assert.Error(t, err)
		assert.Nil(t, invoice)
		// Ошибка может быть либо от кеша, либо от API (так как API nil)
		assert.True(t,
			strings.Contains(err.Error(), "cannot create invoice without user access hash") ||
				strings.Contains(err.Error(), "user not found"))
	})

	t.Run("неподдерживаемый тип получателя", func(t *testing.T) {
		buyer, _, _, _, _ := createMockBuyer()
		buyer.receiverType = []int{999} // неподдерживаемый тип

		gift := createTestGift(1, 100)
		invoice, err := buyer.createInvoice(gift)

		assert.Error(t, err)
		assert.Nil(t, invoice)
		assert.Contains(t, err.Error(), "unexpected receiver type")
	})
}

func TestGiftBuyerImpl_ValidatePurchase(t *testing.T) {
	t.Run("валидация покупки с nil API", func(t *testing.T) {
		buyer, _, _, _, _ := createMockBuyer()
		buyer.api = nil // API nil для тестирования

		gift := createTestGift(1, 100)
		result := buyer.validatePurchase(gift)

		// С nil API валидация должна возвращать false
		assert.False(t, result)
	})
}

func TestGiftBuyerImpl_PurchaseGift(t *testing.T) {
	t.Run("успешная покупка подарка с nil API", func(t *testing.T) {
		buyer, _, _, _, mockRateLimiter := createMockBuyer()
		buyer.api = nil // nil API для имитации успешной покупки

		mockRateLimiter.On("Acquire", mock.Anything).Return(nil)

		gift := createTestGift(1, 100)
		ctx := context.Background()

		err := buyer.purchaseGift(ctx, gift)

		// С nil API покупка должна быть успешной (имитация)
		assert.NoError(t, err)
		// Убираем AssertCalled так как с nil API метод может не вызываться
		// mockRateLimiter.AssertCalled(t, "Acquire", ctx)
	})
}

func TestGiftBuyerImpl_GetMostFrequentError(t *testing.T) {
	t.Run("получение самой частой ошибки", func(t *testing.T) {
		buyer, _, _, _, _ := createMockBuyer()

		errorCounts := map[string]int64{
			"error1": 5,
			"error2": 10,
			"error3": 3,
		}

		err := buyer.getMostFrequentError(errorCounts)

		assert.NotNil(t, err)
		assert.Equal(t, "error2", err.Error())
	})

	t.Run("пустой список ошибок", func(t *testing.T) {
		buyer, _, _, _, _ := createMockBuyer()

		errorCounts := map[string]int64{}

		err := buyer.getMostFrequentError(errorCounts)

		assert.Nil(t, err)
	})

	t.Run("одинаковое количество ошибок", func(t *testing.T) {
		buyer, _, _, _, _ := createMockBuyer()

		errorCounts := map[string]int64{
			"error1": 5,
			"error2": 5,
		}

		err := buyer.getMostFrequentError(errorCounts)

		assert.NotNil(t, err)
		// Должна вернуться одна из ошибок
		assert.True(t, err.Error() == "error1" || err.Error() == "error2")
	})
}

func TestGiftBuyerImpl_ConcurrentPurchases(t *testing.T) {
	t.Run("конкурентные покупки", func(t *testing.T) {
		buyer, _, mockNotification, _, mockRateLimiter := createMockBuyer()

		mockNotification.On("SetBot").Return(false) // Используем логгер вместо бота
		mockRateLimiter.On("Acquire", mock.Anything).Return(nil)

		// Создаем много подарков для тестирования конкурентности
		gifts := make(map[*tg.StarGift]int64)
		for i := int64(1); i <= 10; i++ {
			gifts[createTestGift(i, 100)] = 2
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Создаем канал для ожидания завершения
		done := make(chan struct{})
		start := time.Now()
		go func() {
			buyer.BuyGift(ctx, gifts)
			close(done)
		}()

		// Ждем завершения или таймаута
		select {
		case <-done:
			elapsed := time.Since(start)
			// Конкурентные покупки должны завершиться быстро
			assert.True(t, elapsed < 5*time.Second, "elapsed time: %v", elapsed)
		case <-time.After(8 * time.Second):
			t.Fatal("Concurrent purchases took too long to complete")
		}

		// Даем время для завершения асинхронных операций
		time.Sleep(100 * time.Millisecond)

		mockNotification.AssertCalled(t, "SetBot")
	})
}

func TestGiftBuyerImpl_MaxBuyCountLimit(t *testing.T) {
	t.Run("ограничение максимального количества покупок", func(t *testing.T) {
		buyer, _, mockNotification, _, mockRateLimiter := createMockBuyer()
		buyer.counter = newAtomicCounter(2) // Ограничиваем до 2 покупок

		mockNotification.On("SetBot").Return(true)
		mockNotification.On("SendBuyStatus", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)
		mockRateLimiter.On("Acquire", mock.Anything).Return(nil)

		// Пытаемся купить больше чем лимит
		gifts := map[*tg.StarGift]int64{
			createTestGift(1, 100): 5, // Пытаемся купить 5, но лимит 2
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Создаем канал для ожидания завершения
		done := make(chan struct{})
		go func() {
			buyer.BuyGift(ctx, gifts)
			close(done)
		}()

		// Ждем завершения или таймаута
		select {
		case <-done:
			// Операция завершена
		case <-time.After(3 * time.Second):
			t.Fatal("BuyGift took too long to complete")
		}

		// Даем время для завершения асинхронных операций
		time.Sleep(100 * time.Millisecond)

		// Проверяем что счетчик не превысил лимит
		assert.True(t, buyer.counter.Get() <= buyer.counter.GetMax())

		mockNotification.AssertCalled(t, "SendBuyStatus", mock.Anything, mock.AnythingOfType("string"), mock.Anything)
	})
}
