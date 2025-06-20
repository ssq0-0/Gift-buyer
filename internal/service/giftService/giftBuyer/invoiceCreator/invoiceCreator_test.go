package invoiceCreator

import (
	"context"
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserCache для тестирования
type MockUserCache struct {
	mock.Mock
}

func (m *MockUserCache) SetUser(user *tg.User) {
	m.Called(user)
}

func (m *MockUserCache) GetUser(id int64) (*tg.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tg.User), args.Error(1)
}

func (m *MockUserCache) SetChannel(channel *tg.Channel) {
	m.Called(channel)
}

func (m *MockUserCache) GetChannel(id int64) (*tg.Channel, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tg.Channel), args.Error(1)
}

func createTestGift(id int64, stars int64) *tg.StarGift {
	return &tg.StarGift{
		ID:    id,
		Stars: stars,
	}
}

func TestNewInvoiceCreator(t *testing.T) {
	mockCache := &MockUserCache{}
	userReceiver := []int{123456}
	channelReceiver := []int{789012}
	receiverType := []int{0, 1, 2}

	creator := NewInvoiceCreator(userReceiver, channelReceiver, receiverType, mockCache)

	assert.NotNil(t, creator)
	assert.Equal(t, userReceiver, creator.userReceiver)
	assert.Equal(t, channelReceiver, creator.channelReceiver)
	assert.Equal(t, receiverType, creator.receiverType)
	assert.Equal(t, mockCache, creator.idCache)
}

func TestInvoiceCreatorImpl_CreateInvoice(t *testing.T) {
	t.Run("создание инвойса для разных типов получателей", func(t *testing.T) {
		mockCache := &MockUserCache{}

		// Настраиваем тестовые данные
		testUser := &tg.User{
			ID:         123456,
			AccessHash: 987654321,
		}
		testChannel := &tg.Channel{
			ID:         789012,
			AccessHash: 123456789,
		}

		mockCache.On("GetUser", int64(123456)).Return(testUser, nil)
		mockCache.On("GetChannel", int64(789012)).Return(testChannel, nil)

		creator := NewInvoiceCreator(
			[]int{123456},
			[]int{789012},
			[]int{0, 1, 2}, // все типы
			mockCache,
		)

		gift := createTestGift(1, 100)

		// Тестируем создание инвойса (тип получателя выбирается случайно)
		invoice, err := creator.CreateInvoice(gift)

		assert.NoError(t, err)
		assert.NotNil(t, invoice)
		assert.Equal(t, gift.ID, invoice.GiftID)
		assert.NotEmpty(t, invoice.Message.Text)

		// Проверяем что peer установлен
		assert.NotNil(t, invoice.Peer)
	})
}

func TestInvoiceCreatorImpl_SelfPurchase(t *testing.T) {
	t.Run("создание инвойса для себя", func(t *testing.T) {
		mockCache := &MockUserCache{}
		creator := NewInvoiceCreator([]int{}, []int{}, []int{0}, mockCache)

		gift := createTestGift(1, 100)
		invoice, err := creator.selfPurchase(gift)

		assert.NoError(t, err)
		assert.NotNil(t, invoice)
		assert.Equal(t, gift.ID, invoice.GiftID)
		assert.True(t, invoice.HideName)
		assert.NotEmpty(t, invoice.Message.Text)

		// Проверяем что peer это InputPeerSelf
		_, ok := invoice.Peer.(*tg.InputPeerSelf)
		assert.True(t, ok)
	})
}

func TestInvoiceCreatorImpl_UserPurchase(t *testing.T) {
	t.Run("создание инвойса для пользователя", func(t *testing.T) {
		mockCache := &MockUserCache{}
		testUser := &tg.User{
			ID:         123456,
			AccessHash: 987654321,
		}
		mockCache.On("GetUser", int64(123456)).Return(testUser, nil)

		creator := NewInvoiceCreator([]int{123456}, []int{}, []int{1}, mockCache)

		gift := createTestGift(1, 100)
		invoice, err := creator.userPurchase(gift)

		assert.NoError(t, err)
		assert.NotNil(t, invoice)
		assert.Equal(t, gift.ID, invoice.GiftID)
		assert.NotEmpty(t, invoice.Message.Text)

		// Проверяем что peer это InputPeerUser
		peerUser, ok := invoice.Peer.(*tg.InputPeerUser)
		assert.True(t, ok)
		assert.Equal(t, testUser.ID, peerUser.UserID)
		assert.Equal(t, testUser.AccessHash, peerUser.AccessHash)

		mockCache.AssertCalled(t, "GetUser", int64(123456))
	})

	t.Run("ошибка при получении пользователя", func(t *testing.T) {
		mockCache := &MockUserCache{}
		mockCache.On("GetUser", int64(123456)).Return(nil, assert.AnError)

		creator := NewInvoiceCreator([]int{123456}, []int{}, []int{1}, mockCache)

		gift := createTestGift(1, 100)
		invoice, err := creator.userPurchase(gift)

		assert.Error(t, err)
		assert.Nil(t, invoice)
		assert.Contains(t, err.Error(), "cannot create invoice without user access hash")

		mockCache.AssertCalled(t, "GetUser", int64(123456))
	})
}

func TestInvoiceCreatorImpl_ChannelPurchase(t *testing.T) {
	t.Run("создание инвойса для канала", func(t *testing.T) {
		mockCache := &MockUserCache{}
		testChannel := &tg.Channel{
			ID:         789012,
			AccessHash: 123456789,
		}
		mockCache.On("GetChannel", int64(789012)).Return(testChannel, nil)

		creator := NewInvoiceCreator([]int{}, []int{789012}, []int{2}, mockCache)

		gift := createTestGift(1, 100)
		invoice, err := creator.channelPurchase(gift)

		assert.NoError(t, err)
		assert.NotNil(t, invoice)
		assert.Equal(t, gift.ID, invoice.GiftID)
		assert.True(t, invoice.HideName)
		assert.NotEmpty(t, invoice.Message.Text)

		// Проверяем что peer это InputPeerChannel
		peerChannel, ok := invoice.Peer.(*tg.InputPeerChannel)
		assert.True(t, ok)
		assert.Equal(t, testChannel.ID, peerChannel.ChannelID)
		assert.Equal(t, testChannel.AccessHash, peerChannel.AccessHash)

		mockCache.AssertCalled(t, "GetChannel", int64(789012))
	})

	t.Run("ошибка при получении канала", func(t *testing.T) {
		mockCache := &MockUserCache{}
		mockCache.On("GetChannel", int64(789012)).Return(nil, assert.AnError)

		creator := NewInvoiceCreator([]int{}, []int{789012}, []int{2}, mockCache)

		gift := createTestGift(1, 100)
		invoice, err := creator.channelPurchase(gift)

		assert.Error(t, err)
		assert.Nil(t, invoice)
		assert.Contains(t, err.Error(), "cannot create invoice without channel access hash")

		mockCache.AssertCalled(t, "GetChannel", int64(789012))
	})
}

func TestInvoiceCreatorImpl_GetUserInfo(t *testing.T) {
	t.Run("успешное получение пользователя из кеша", func(t *testing.T) {
		mockCache := &MockUserCache{}
		testUser := &tg.User{
			ID:         123456,
			AccessHash: 987654321,
		}
		mockCache.On("GetUser", int64(123456)).Return(testUser, nil)

		creator := NewInvoiceCreator([]int{123456}, []int{}, []int{1}, mockCache)

		user, err := creator.getUserInfo(context.Background(), 123456)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.AccessHash, user.AccessHash)

		mockCache.AssertCalled(t, "GetUser", int64(123456))
	})

	t.Run("пользователь не найден в кеше", func(t *testing.T) {
		mockCache := &MockUserCache{}
		mockCache.On("GetUser", int64(123456)).Return(nil, assert.AnError)

		creator := NewInvoiceCreator([]int{123456}, []int{}, []int{1}, mockCache)

		user, err := creator.getUserInfo(context.Background(), 123456)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user 123456 not accessible")

		mockCache.AssertCalled(t, "GetUser", int64(123456))
	})
}

func TestInvoiceCreatorImpl_GetChannelInfo(t *testing.T) {
	t.Run("успешное получение канала из кеша", func(t *testing.T) {
		mockCache := &MockUserCache{}
		testChannel := &tg.Channel{
			ID:         789012,
			AccessHash: 123456789,
		}
		mockCache.On("GetChannel", int64(789012)).Return(testChannel, nil)

		creator := NewInvoiceCreator([]int{}, []int{789012}, []int{2}, mockCache)

		channel, err := creator.getChannelInfo(context.Background(), 789012)

		assert.NoError(t, err)
		assert.NotNil(t, channel)
		assert.Equal(t, testChannel.ID, channel.ID)
		assert.Equal(t, testChannel.AccessHash, channel.AccessHash)

		mockCache.AssertCalled(t, "GetChannel", int64(789012))
	})

	t.Run("канал не найден в кеше", func(t *testing.T) {
		mockCache := &MockUserCache{}
		mockCache.On("GetChannel", int64(789012)).Return(nil, assert.AnError)

		creator := NewInvoiceCreator([]int{}, []int{789012}, []int{2}, mockCache)

		channel, err := creator.getChannelInfo(context.Background(), 789012)

		assert.Error(t, err)
		assert.Nil(t, channel)
		assert.Equal(t, "channel not found", err.Error())

		mockCache.AssertCalled(t, "GetChannel", int64(789012))
	})
}
