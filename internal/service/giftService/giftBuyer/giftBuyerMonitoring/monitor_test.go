package giftBuyerMonitoring

import (
	"context"
	"testing"
	"time"

	"gift-buyer/internal/service/giftService/giftTypes"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNotificationService для тестирования
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

func createTestGift(id int64, stars int64) *tg.StarGift {
	return &tg.StarGift{
		ID:    id,
		Stars: stars,
	}
}

func TestNewGiftBuyerMonitoring(t *testing.T) {
	mockNotification := &MockNotificationService{}

	monitor := NewGiftBuyerMonitoring(nil, mockNotification)

	assert.NotNil(t, monitor)
	assert.Nil(t, monitor.api)
	assert.Equal(t, mockNotification, monitor.notification)
}

func TestGiftBuyerMonitoringImpl_MonitorProcess(t *testing.T) {
	t.Run("успешный мониторинг с результатами", func(t *testing.T) {
		mockNotification := &MockNotificationService{}
		monitor := NewGiftBuyerMonitoring(nil, mockNotification)

		gifts := map[*tg.StarGift]int64{
			createTestGift(1, 100): 2,
			createTestGift(2, 200): 1,
		}

		resultsCh := make(chan giftTypes.GiftResult, 10)
		doneChan := make(chan struct{})

		// Настраиваем мок
		mockNotification.On("SetBot").Return(true)
		mockNotification.On("SendBuyStatus", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Запускаем монитор в отдельной горутине
		go monitor.MonitorProcess(ctx, resultsCh, doneChan, gifts)

		// Отправляем успешные результаты
		resultsCh <- giftTypes.GiftResult{GiftID: 1, Success: true, Err: nil}
		resultsCh <- giftTypes.GiftResult{GiftID: 1, Success: true, Err: nil}
		resultsCh <- giftTypes.GiftResult{GiftID: 2, Success: true, Err: nil}

		// Сигнализируем о завершении
		close(doneChan)

		// Даем время для обработки
		time.Sleep(100 * time.Millisecond)

		mockNotification.AssertCalled(t, "SetBot")
		mockNotification.AssertCalled(t, "SendBuyStatus", mock.Anything, mock.AnythingOfType("string"), mock.Anything)
	})

	t.Run("частичный успех", func(t *testing.T) {
		mockNotification := &MockNotificationService{}
		monitor := NewGiftBuyerMonitoring(nil, mockNotification)

		gifts := map[*tg.StarGift]int64{
			createTestGift(1, 100): 2,
		}

		resultsCh := make(chan giftTypes.GiftResult, 10)
		doneChan := make(chan struct{})

		// Настраиваем мок
		mockNotification.On("SetBot").Return(true)
		mockNotification.On("SendBuyStatus", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)

		ctx := context.Background()

		// Запускаем монитор в отдельной горутине
		go monitor.MonitorProcess(ctx, resultsCh, doneChan, gifts)

		// Отправляем смешанные результаты
		resultsCh <- giftTypes.GiftResult{GiftID: 1, Success: true, Err: nil}
		resultsCh <- giftTypes.GiftResult{GiftID: 1, Success: false, Err: assert.AnError}

		// Сигнализируем о завершении
		close(doneChan)

		// Даем время для обработки
		time.Sleep(100 * time.Millisecond)

		mockNotification.AssertCalled(t, "SetBot")
		mockNotification.AssertCalled(t, "SendBuyStatus", mock.Anything, mock.AnythingOfType("string"), mock.Anything)
	})

	t.Run("полный провал", func(t *testing.T) {
		mockNotification := &MockNotificationService{}
		monitor := NewGiftBuyerMonitoring(nil, mockNotification)

		gifts := map[*tg.StarGift]int64{
			createTestGift(1, 100): 1,
		}

		resultsCh := make(chan giftTypes.GiftResult, 10)
		doneChan := make(chan struct{})

		// Настраиваем мок
		mockNotification.On("SetBot").Return(true)
		mockNotification.On("SendBuyStatus", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)

		ctx := context.Background()

		// Запускаем монитор в отдельной горутине
		go monitor.MonitorProcess(ctx, resultsCh, doneChan, gifts)

		// Отправляем только неудачные результаты
		resultsCh <- giftTypes.GiftResult{GiftID: 1, Success: false, Err: assert.AnError}

		// Сигнализируем о завершении
		close(doneChan)

		// Даем время для обработки
		time.Sleep(100 * time.Millisecond)

		mockNotification.AssertCalled(t, "SetBot")
		mockNotification.AssertCalled(t, "SendBuyStatus", mock.Anything, mock.AnythingOfType("string"), mock.Anything)
	})

	t.Run("использование логгера вместо бота", func(t *testing.T) {
		mockNotification := &MockNotificationService{}
		monitor := NewGiftBuyerMonitoring(nil, mockNotification)

		gifts := map[*tg.StarGift]int64{
			createTestGift(1, 100): 1,
		}

		resultsCh := make(chan giftTypes.GiftResult, 10)
		doneChan := make(chan struct{})

		// Настраиваем мок для использования логгера
		mockNotification.On("SetBot").Return(false)

		ctx := context.Background()

		// Запускаем монитор в отдельной горутине
		go monitor.MonitorProcess(ctx, resultsCh, doneChan, gifts)

		// Отправляем успешный результат
		resultsCh <- giftTypes.GiftResult{GiftID: 1, Success: true, Err: nil}

		// Сигнализируем о завершении
		close(doneChan)

		// Даем время для обработки
		time.Sleep(100 * time.Millisecond)

		mockNotification.AssertCalled(t, "SetBot")
		// SendBuyStatus не должен быть вызван при использовании логгера
		mockNotification.AssertNotCalled(t, "SendBuyStatus")
	})

	t.Run("отмена контекста", func(t *testing.T) {
		mockNotification := &MockNotificationService{}
		monitor := NewGiftBuyerMonitoring(nil, mockNotification)

		gifts := map[*tg.StarGift]int64{
			createTestGift(1, 100): 1,
		}

		resultsCh := make(chan giftTypes.GiftResult, 10)
		doneChan := make(chan struct{})

		ctx, cancel := context.WithCancel(context.Background())

		// Запускаем монитор в отдельной горутине
		done := make(chan struct{})
		go func() {
			monitor.MonitorProcess(ctx, resultsCh, doneChan, gifts)
			close(done)
		}()

		// Отменяем контекст
		cancel()

		// Ждем завершения
		select {
		case <-done:
			// Монитор должен завершиться
		case <-time.After(500 * time.Millisecond):
			t.Fatal("MonitorProcess не завершился после отмены контекста")
		}

		// Никакие уведомления не должны быть отправлены
		mockNotification.AssertNotCalled(t, "SetBot")
		mockNotification.AssertNotCalled(t, "SendBuyStatus")
	})

	t.Run("закрытие канала результатов", func(t *testing.T) {
		mockNotification := &MockNotificationService{}
		monitor := NewGiftBuyerMonitoring(nil, mockNotification)

		gifts := map[*tg.StarGift]int64{
			createTestGift(1, 100): 1,
		}

		resultsCh := make(chan giftTypes.GiftResult, 10)
		doneChan := make(chan struct{})

		ctx := context.Background()

		// Запускаем монитор в отдельной горутине
		done := make(chan struct{})
		go func() {
			monitor.MonitorProcess(ctx, resultsCh, doneChan, gifts)
			close(done)
		}()

		// Закрываем канал результатов
		close(resultsCh)

		// Ждем завершения
		select {
		case <-done:
			// Монитор должен завершиться
		case <-time.After(500 * time.Millisecond):
			t.Fatal("MonitorProcess не завершился после закрытия канала результатов")
		}

		// Никакие уведомления не должны быть отправлены
		mockNotification.AssertNotCalled(t, "SetBot")
		mockNotification.AssertNotCalled(t, "SendBuyStatus")
	})
}

func TestGiftBuyerMonitoringImpl_GetMostFrequentError(t *testing.T) {
	t.Run("получение самой частой ошибки", func(t *testing.T) {
		monitor := &GiftBuyerMonitoringImpl{}

		errorCounts := map[string]int64{
			"error1": 5,
			"error2": 10,
			"error3": 3,
		}

		err := monitor.getMostFrequentError(errorCounts)

		assert.NotNil(t, err)
		assert.Equal(t, "error2", err.Error())
	})

	t.Run("пустой список ошибок", func(t *testing.T) {
		monitor := &GiftBuyerMonitoringImpl{}

		errorCounts := map[string]int64{}

		err := monitor.getMostFrequentError(errorCounts)

		assert.Nil(t, err)
	})

	t.Run("одинаковое количество ошибок", func(t *testing.T) {
		monitor := &GiftBuyerMonitoringImpl{}

		errorCounts := map[string]int64{
			"error1": 5,
			"error2": 5,
		}

		err := monitor.getMostFrequentError(errorCounts)

		assert.NotNil(t, err)
		// Должна вернуться одна из ошибок
		assert.True(t, err.Error() == "error1" || err.Error() == "error2")
	})

	t.Run("единственная ошибка", func(t *testing.T) {
		monitor := &GiftBuyerMonitoringImpl{}

		errorCounts := map[string]int64{
			"single error": 1,
		}

		err := monitor.getMostFrequentError(errorCounts)

		assert.NotNil(t, err)
		assert.Equal(t, "single error", err.Error())
	})
}
