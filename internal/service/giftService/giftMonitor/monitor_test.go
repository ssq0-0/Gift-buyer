package giftMonitor

import (
	"context"
	"testing"
	"time"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGiftCache is a mock implementation of GiftCache interface
type MockGiftCache struct {
	mock.Mock
}

func (m *MockGiftCache) SetGift(id int64, gift *tg.StarGift) {
	m.Called(id, gift)
}

func (m *MockGiftCache) GetGift(id int64) (*tg.StarGift, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tg.StarGift), args.Error(1)
}

func (m *MockGiftCache) GetAllGifts() map[int64]*tg.StarGift {
	args := m.Called()
	return args.Get(0).(map[int64]*tg.StarGift)
}

func (m *MockGiftCache) HasGift(id int64) bool {
	args := m.Called(id)
	return args.Bool(0)
}

func (m *MockGiftCache) DeleteGift(id int64) {
	m.Called(id)
}

func (m *MockGiftCache) Clear() {
	m.Called()
}

// MockGiftManager is a mock implementation of Giftmanager interface
type MockGiftManager struct {
	mock.Mock
}

func (m *MockGiftManager) GetAvailableGifts(ctx context.Context) ([]*tg.StarGift, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tg.StarGift), args.Error(1)
}

// MockGiftValidator is a mock implementation of GiftValidator interface
type MockGiftValidator struct {
	mock.Mock
}

func (m *MockGiftValidator) IsEligible(gift *tg.StarGift) (int64, bool) {
	args := m.Called(gift)
	return args.Get(0).(int64), args.Bool(1)
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

func (m *MockNotificationService) SendErrorNotification(ctx context.Context, err error) error {
	args := m.Called(ctx, err)
	return args.Error(0)
}

func (m *MockNotificationService) SetBot() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockNotificationService) SendUpdateNotification(ctx context.Context, version, message string) error {
	args := m.Called(ctx, version, message)
	return args.Error(0)
}

func TestNewGiftMonitor(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)
	tickTime := time.Second

	monitor := NewGiftMonitor(mockCache, mockManager, mockValidator, mockNotification, tickTime)

	assert.NotNil(t, monitor)

	// Cast to concrete type to verify internal structure
	gm, ok := monitor.(*GiftMonitorImpl)
	assert.True(t, ok)
	assert.Equal(t, mockCache, gm.cache)
	assert.Equal(t, mockManager, gm.manager)
	assert.Equal(t, mockValidator, gm.validator)
	assert.Equal(t, mockNotification, gm.notification)
	assert.NotNil(t, gm.ticker)
}

func TestGiftMonitor_Start_FirstRun(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)

	monitor := NewGiftMonitor(mockCache, mockManager, mockValidator, mockNotification, time.Millisecond*10)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	// Create test gifts
	gift1 := &tg.StarGift{ID: 1, Stars: 100}
	gift2 := &tg.StarGift{ID: 2, Stars: 200}
	currentGifts := []*tg.StarGift{gift1, gift2}

	// First run - should return "touch grass" error and continue monitoring
	mockManager.On("GetAvailableGifts", mock.AnythingOfType("*context.timerCtx")).Return(currentGifts, nil).Times(1)
	mockCache.On("HasGift", int64(1)).Return(false).Times(1)
	mockCache.On("HasGift", int64(2)).Return(false).Times(1)
	mockValidator.On("IsEligible", gift1).Return(int64(10), true).Times(1)
	mockValidator.On("IsEligible", gift2).Return(int64(20), true).Times(1)
	mockCache.On("SetGift", int64(1), gift1).Return().Times(1)
	mockCache.On("SetGift", int64(2), gift2).Return().Times(1)
	mockNotification.On("SendErrorNotification", mock.Anything, mock.MatchedBy(func(err error) bool {
		return err.Error() == "touch grass: first run"
	})).Return(nil).Times(1)

	// Second run - no new gifts since they're in cache
	mockManager.On("GetAvailableGifts", mock.AnythingOfType("*context.timerCtx")).Return(currentGifts, nil)
	mockCache.On("HasGift", int64(1)).Return(true)
	mockCache.On("HasGift", int64(2)).Return(true)
	mockCache.On("SetGift", int64(1), gift1).Return()
	mockCache.On("SetGift", int64(2), gift2).Return()

	newGifts, err := monitor.Start(ctx)

	// Should timeout since no new gifts are found after first run
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.Nil(t, newGifts)

	mockCache.AssertExpectations(t)
	mockManager.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
	mockNotification.AssertExpectations(t)
}

func TestGiftMonitor_Start_SecondRunWithNewGifts(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)

	// Create monitor and skip first run manually
	gm := &GiftMonitorImpl{
		cache:        mockCache,
		manager:      mockManager,
		validator:    mockValidator,
		notification: mockNotification,
		ticker:       time.NewTicker(time.Millisecond * 10),
		firstRun:     false, // Skip first run
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	// Create test gifts
	gift1 := &tg.StarGift{ID: 1, Stars: 100}
	gift2 := &tg.StarGift{ID: 2, Stars: 200}
	currentGifts := []*tg.StarGift{gift1, gift2}

	// Setup mocks for new gifts
	mockManager.On("GetAvailableGifts", mock.AnythingOfType("*context.timerCtx")).Return(currentGifts, nil)
	mockCache.On("HasGift", int64(1)).Return(false)
	mockCache.On("HasGift", int64(2)).Return(false)
	mockValidator.On("IsEligible", gift1).Return(int64(10), true)
	mockValidator.On("IsEligible", gift2).Return(int64(20), true)
	mockCache.On("SetGift", int64(1), gift1).Return()
	mockCache.On("SetGift", int64(2), gift2).Return()

	newGifts, err := gm.Start(ctx)

	assert.NoError(t, err)
	assert.Len(t, newGifts, 2)
	assert.Contains(t, newGifts, gift1)
	assert.Contains(t, newGifts, gift2)
	assert.Equal(t, int64(10), newGifts[gift1])
	assert.Equal(t, int64(20), newGifts[gift2])

	mockCache.AssertExpectations(t)
	mockManager.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

func TestGiftMonitor_Start_ContextCancelled(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)

	monitor := NewGiftMonitor(mockCache, mockManager, mockValidator, mockNotification, time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	newGifts, err := monitor.Start(ctx)

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Nil(t, newGifts)
}

func TestGiftMonitor_Start_NoNewGifts(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)

	monitor := NewGiftMonitor(mockCache, mockManager, mockValidator, mockNotification, time.Millisecond*10)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	// Setup mocks to return no new gifts on first run (triggers SendErrorNotification)
	mockManager.On("GetAvailableGifts", mock.AnythingOfType("*context.timerCtx")).Return([]*tg.StarGift{}, nil)
	mockNotification.On("SendErrorNotification", mock.Anything, mock.MatchedBy(func(err error) bool {
		return err.Error() == "touch grass: first run"
	})).Return(nil).Times(1)

	// Start monitoring - it should continue until context timeout
	newGifts, err := monitor.Start(ctx)

	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.Nil(t, newGifts)

	mockCache.AssertExpectations(t)
	mockManager.AssertExpectations(t)
	mockNotification.AssertExpectations(t)
}

func TestGiftMonitor_CheckForNewGifts_Success(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)

	monitor := &GiftMonitorImpl{
		cache:        mockCache,
		manager:      mockManager,
		validator:    mockValidator,
		notification: mockNotification,
		ticker:       time.NewTicker(time.Second),
	}

	ctx := context.Background()

	// Create test gifts
	gift1 := &tg.StarGift{ID: 1, Stars: 100}
	gift2 := &tg.StarGift{ID: 2, Stars: 200}
	currentGifts := []*tg.StarGift{gift1, gift2}

	// Setup mocks
	mockManager.On("GetAvailableGifts", ctx).Return(currentGifts, nil)
	mockCache.On("HasGift", int64(1)).Return(false)
	mockCache.On("HasGift", int64(2)).Return(false)
	mockValidator.On("IsEligible", gift1).Return(int64(10), true)
	mockValidator.On("IsEligible", gift2).Return(int64(20), true)
	mockCache.On("SetGift", int64(1), gift1).Return()
	mockCache.On("SetGift", int64(2), gift2).Return()

	newGifts, err := monitor.checkForNewGifts(ctx)

	assert.NoError(t, err)
	assert.Len(t, newGifts, 2) // Both gifts should be eligible
	assert.Contains(t, newGifts, gift1)
	assert.Contains(t, newGifts, gift2)

	mockCache.AssertExpectations(t)
	mockManager.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

func TestGiftMonitor_CheckForNewGifts_ManagerError(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)

	monitor := &GiftMonitorImpl{
		cache:        mockCache,
		manager:      mockManager,
		validator:    mockValidator,
		notification: mockNotification,
		ticker:       time.NewTicker(time.Second),
	}

	ctx := context.Background()
	expectedError := assert.AnError

	mockManager.On("GetAvailableGifts", ctx).Return(([]*tg.StarGift)(nil), expectedError)

	newGifts, err := monitor.checkForNewGifts(ctx)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, newGifts)

	mockManager.AssertExpectations(t)
}

func TestGiftMonitor_CheckForNewGifts_ExistingGifts(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)

	monitor := &GiftMonitorImpl{
		cache:        mockCache,
		manager:      mockManager,
		validator:    mockValidator,
		notification: mockNotification,
		ticker:       time.NewTicker(time.Second),
	}

	ctx := context.Background()

	// Create test gifts
	gift1 := &tg.StarGift{ID: 1, Stars: 100}
	currentGifts := []*tg.StarGift{gift1}

	// Setup mocks - gift already exists in cache
	mockManager.On("GetAvailableGifts", ctx).Return(currentGifts, nil)
	mockCache.On("HasGift", int64(1)).Return(true)

	newGifts, err := monitor.checkForNewGifts(ctx)

	assert.NoError(t, err)
	assert.Empty(t, newGifts) // No new gifts since it already exists

	mockCache.AssertExpectations(t)
	mockManager.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

func TestGiftMonitor_CheckForNewGifts_NotEligible(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)

	monitor := &GiftMonitorImpl{
		cache:        mockCache,
		manager:      mockManager,
		validator:    mockValidator,
		notification: mockNotification,
		ticker:       time.NewTicker(time.Second),
	}

	ctx := context.Background()

	// Create test gifts
	gift1 := &tg.StarGift{ID: 1, Stars: 100}
	currentGifts := []*tg.StarGift{gift1}

	// Setup mocks - gift is not eligible
	mockManager.On("GetAvailableGifts", ctx).Return(currentGifts, nil)
	mockCache.On("HasGift", int64(1)).Return(false)
	mockValidator.On("IsEligible", gift1).Return(int64(0), false)
	mockCache.On("SetGift", int64(1), gift1).Return()

	newGifts, err := monitor.checkForNewGifts(ctx)

	assert.NoError(t, err)
	assert.Empty(t, newGifts) // No eligible gifts

	mockCache.AssertExpectations(t)
	mockManager.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}
