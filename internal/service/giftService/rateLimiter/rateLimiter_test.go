package rateLimiter

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimiter(t *testing.T) {
	tests := []struct {
		name string
		rps  int
	}{
		{
			name: "создание rate limiter с 1 RPS",
			rps:  1,
		},
		{
			name: "создание rate limiter с 10 RPS",
			rps:  10,
		},
		{
			name: "создание rate limiter с 100 RPS",
			rps:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := NewRateLimiter(tt.rps)
			assert.NotNil(t, rl)
			assert.Equal(t, tt.rps, rl.maxTokens)
			assert.NotNil(t, rl.tokens)
			assert.NotNil(t, rl.ticker)
			assert.False(t, rl.closed)

			// Проверяем что канал заполнен токенами
			assert.Equal(t, tt.rps, len(rl.tokens))

			rl.Close()
		})
	}
}

func TestRateLimiter_Acquire(t *testing.T) {
	t.Run("успешное получение токена", func(t *testing.T) {
		rl := NewRateLimiter(5)
		defer rl.Close()

		ctx := context.Background()
		err := rl.Acquire(ctx)
		assert.NoError(t, err)
	})

	t.Run("получение токена с отменой контекста", func(t *testing.T) {
		rl := NewRateLimiter(1)
		defer rl.Close()

		// Сначала берем единственный токен
		ctx := context.Background()
		err := rl.Acquire(ctx)
		require.NoError(t, err)

		// Теперь пытаемся взять еще один с отмененным контекстом
		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		err = rl.Acquire(canceledCtx)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("получение токена с таймаутом", func(t *testing.T) {
		rl := NewRateLimiter(1)
		defer rl.Close()

		// Берем единственный токен
		ctx := context.Background()
		err := rl.Acquire(ctx)
		require.NoError(t, err)

		// Пытаемся взять еще один с коротким таймаутом
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err = rl.Acquire(timeoutCtx)
		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
	})
}

func TestRateLimiter_RateLimit(t *testing.T) {
	t.Run("проверка ограничения скорости", func(t *testing.T) {
		rps := 2
		rl := NewRateLimiter(rps)
		defer rl.Close()

		ctx := context.Background()
		start := time.Now()

		// Берем все доступные токены
		for i := 0; i < rps; i++ {
			err := rl.Acquire(ctx)
			require.NoError(t, err)
		}

		// Следующий запрос должен заблокироваться
		err := rl.Acquire(ctx)
		elapsed := time.Since(start)

		assert.NoError(t, err)
		// Должно пройти примерно время для пополнения токенов
		assert.True(t, elapsed >= 400*time.Millisecond, "elapsed time: %v", elapsed)
	})
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	t.Run("конкурентный доступ к rate limiter", func(t *testing.T) {
		rps := 10
		rl := NewRateLimiter(rps)
		defer rl.Close()

		var wg sync.WaitGroup
		var successCount int64
		var mu sync.Mutex

		numGoroutines := 20
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := rl.Acquire(ctx)
				if err == nil {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}()
		}

		wg.Wait()

		mu.Lock()
		finalCount := successCount
		mu.Unlock()

		// Все горутины должны получить токены (возможно с задержкой)
		// Допускаем небольшую погрешность из-за таймингов
		assert.GreaterOrEqual(t, finalCount, int64(numGoroutines-1))
	})
}

func TestRateLimiter_Close(t *testing.T) {
	t.Run("закрытие rate limiter", func(t *testing.T) {
		rl := NewRateLimiter(5)

		assert.False(t, rl.closed)
		assert.NotNil(t, rl.ticker)

		rl.Close()

		assert.True(t, rl.closed)
	})

	t.Run("повторное закрытие rate limiter", func(t *testing.T) {
		rl := NewRateLimiter(5)

		rl.Close()
		assert.True(t, rl.closed)

		// Повторное закрытие не должно вызывать панику
		assert.NotPanics(t, func() {
			rl.Close()
		})
	})
}

func TestRateLimiter_RefillTokens(t *testing.T) {
	t.Run("пополнение токенов", func(t *testing.T) {
		rps := 2
		rl := NewRateLimiter(rps)
		defer rl.Close()

		ctx := context.Background()

		// Берем все токены
		for i := 0; i < rps; i++ {
			err := rl.Acquire(ctx)
			require.NoError(t, err)
		}

		// Ждем пополнения
		time.Sleep(time.Second + 100*time.Millisecond)

		// Должны быть доступны новые токены
		err := rl.Acquire(ctx)
		assert.NoError(t, err)
	})
}

func TestRateLimiter_EdgeCases(t *testing.T) {
	t.Run("rate limiter с 0 RPS", func(t *testing.T) {
		// Хотя это не практичный случай, код должен работать
		rl := NewRateLimiter(0)
		defer rl.Close()

		assert.Equal(t, 0, rl.maxTokens)
		assert.Equal(t, 0, len(rl.tokens))
	})

	t.Run("rate limiter с большим RPS", func(t *testing.T) {
		rps := 1000
		rl := NewRateLimiter(rps)
		defer rl.Close()

		assert.Equal(t, rps, rl.maxTokens)
		// Изначально канал должен быть заполнен
		initialTokens := len(rl.tokens)
		assert.Equal(t, rps, initialTokens)

		// Должны быть доступны все токены
		ctx := context.Background()
		for i := 0; i < rps; i++ {
			err := rl.Acquire(ctx)
			assert.NoError(t, err)
		}

		// После использования всех токенов канал должен быть пустым
		assert.Equal(t, 0, len(rl.tokens))
	})
}

func TestRateLimiter_Performance(t *testing.T) {
	t.Run("производительность rate limiter", func(t *testing.T) {
		rps := 100
		rl := NewRateLimiter(rps)
		defer rl.Close()

		ctx := context.Background()
		start := time.Now()

		// Берем много токенов
		for i := 0; i < rps; i++ {
			err := rl.Acquire(ctx)
			require.NoError(t, err)
		}

		elapsed := time.Since(start)
		// Первые токены должны быть получены быстро
		assert.True(t, elapsed < 100*time.Millisecond, "elapsed time: %v", elapsed)
	})
}
