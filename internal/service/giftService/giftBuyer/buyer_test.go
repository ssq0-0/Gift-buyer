package giftBuyer

import (
	"context"
	"sync"
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockGiftManager struct {
	mock.Mock
}

func (m *MockGiftManager) GetAvailableGifts(ctx context.Context) ([]*tg.StarGift, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*tg.StarGift), args.Error(1)
}

// MockNotificationService is a mock implementation of NotificationService interface
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

func (m *MockNotificationService) SendBulkGiftNotification(ctx context.Context, gifts []*tg.StarGift) error {
	args := m.Called(ctx, gifts)
	return args.Error(0)
}

func (m *MockNotificationService) SetBot() bool {
	args := m.Called()
	return args.Bool(0)
}

func TestAtomicCounter(t *testing.T) {
	t.Run("NewAtomicCounter", func(t *testing.T) {
		counter := newAtomicCounter(5)
		assert.Equal(t, int64(0), counter.Get())
		assert.Equal(t, int64(5), counter.GetMax())
	})

	t.Run("TryIncrement_Success", func(t *testing.T) {
		counter := newAtomicCounter(3)

		assert.True(t, counter.TryIncrement())
		assert.Equal(t, int64(1), counter.Get())

		assert.True(t, counter.TryIncrement())
		assert.Equal(t, int64(2), counter.Get())

		assert.True(t, counter.TryIncrement())
		assert.Equal(t, int64(3), counter.Get())
	})

	t.Run("TryIncrement_Limit", func(t *testing.T) {
		counter := newAtomicCounter(2)

		assert.True(t, counter.TryIncrement())
		assert.True(t, counter.TryIncrement())

		assert.False(t, counter.TryIncrement())
		assert.Equal(t, int64(2), counter.Get())
	})

	t.Run("Decrement", func(t *testing.T) {
		counter := newAtomicCounter(5)
		counter.TryIncrement()
		counter.TryIncrement()
		assert.Equal(t, int64(2), counter.Get())

		counter.Decrement()
		assert.Equal(t, int64(1), counter.Get())

		counter.Decrement()
		assert.Equal(t, int64(0), counter.Get())
	})

	t.Run("Concurrent_Access", func(t *testing.T) {
		counter := newAtomicCounter(100)
		var wg sync.WaitGroup

		for i := 0; i < 200; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				counter.TryIncrement()
			}()
		}

		wg.Wait()

		assert.Equal(t, int64(100), counter.Get())
	})
}

func TestGiftBuyer(t *testing.T) {
	t.Run("NewGiftBuyer", func(t *testing.T) {
		mockManager := &MockGiftManager{}
		mockNotification := &MockNotificationService{}
		api := &tg.Client{}

		buyer := NewGiftBuyer(api, 123, 0, mockManager, mockNotification, 10).(*GiftBuyerImpl)

		assert.NotNil(t, buyer)
		assert.Equal(t, int64(0), buyer.GetTotalBuyCount())
	})

	t.Run("AtomicCounter_MaxCountReached", func(t *testing.T) {
		counter := newAtomicCounter(2)

		assert.True(t, counter.TryIncrement())
		assert.True(t, counter.TryIncrement())

		assert.False(t, counter.TryIncrement())
		assert.Equal(t, int64(2), counter.Get())
	})

	t.Run("AtomicCounter_Decrement", func(t *testing.T) {
		counter := newAtomicCounter(5)

		counter.TryIncrement()
		counter.TryIncrement()
		assert.Equal(t, int64(2), counter.Get())

		counter.Decrement()
		assert.Equal(t, int64(1), counter.Get())

		assert.True(t, counter.TryIncrement())
		assert.Equal(t, int64(2), counter.Get())
	})

	t.Run("AtomicCounter_ConcurrentAccess", func(t *testing.T) {
		counter := newAtomicCounter(50)
		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				counter.TryIncrement()
			}()
		}

		wg.Wait()

		assert.Equal(t, int64(50), counter.Get())
	})
}
