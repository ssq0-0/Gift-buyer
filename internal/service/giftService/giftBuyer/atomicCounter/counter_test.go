package atomicCounter

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAtomicCounter(t *testing.T) {
	t.Run("создание нового счетчика", func(t *testing.T) {
		max := int64(100)
		counter := NewAtomicCounter(max)

		assert.NotNil(t, counter)
		assert.Equal(t, int64(0), counter.Get())
		assert.Equal(t, max, counter.GetMax())
	})

	t.Run("создание счетчика с нулевым максимумом", func(t *testing.T) {
		counter := NewAtomicCounter(0)

		assert.NotNil(t, counter)
		assert.Equal(t, int64(0), counter.Get())
		assert.Equal(t, int64(0), counter.GetMax())
	})

	t.Run("создание счетчика с отрицательным максимумом", func(t *testing.T) {
		counter := NewAtomicCounter(-5)

		assert.NotNil(t, counter)
		assert.Equal(t, int64(0), counter.Get())
		assert.Equal(t, int64(-5), counter.GetMax())
	})
}

func TestAtomicCounter_TryIncrement(t *testing.T) {
	t.Run("успешное увеличение счетчика", func(t *testing.T) {
		counter := NewAtomicCounter(5)

		// Первое увеличение должно быть успешным
		result := counter.TryIncrement()
		assert.True(t, result)
		assert.Equal(t, int64(1), counter.Get())

		// Второе увеличение тоже должно быть успешным
		result = counter.TryIncrement()
		assert.True(t, result)
		assert.Equal(t, int64(2), counter.Get())
	})

	t.Run("достижение максимального значения", func(t *testing.T) {
		counter := NewAtomicCounter(2)

		// Увеличиваем до максимума
		assert.True(t, counter.TryIncrement())
		assert.True(t, counter.TryIncrement())
		assert.Equal(t, int64(2), counter.Get())

		// Попытка увеличить сверх максимума должна провалиться
		result := counter.TryIncrement()
		assert.False(t, result)
		assert.Equal(t, int64(2), counter.Get())
	})

	t.Run("счетчик с максимумом 0", func(t *testing.T) {
		counter := NewAtomicCounter(0)

		// Любая попытка увеличения должна провалиться
		result := counter.TryIncrement()
		assert.False(t, result)
		assert.Equal(t, int64(0), counter.Get())
	})

	t.Run("счетчик с отрицательным максимумом", func(t *testing.T) {
		counter := NewAtomicCounter(-1)

		// Любая попытка увеличения должна провалиться
		result := counter.TryIncrement()
		assert.False(t, result)
		assert.Equal(t, int64(0), counter.Get())
	})
}

func TestAtomicCounter_Decrement(t *testing.T) {
	t.Run("уменьшение счетчика", func(t *testing.T) {
		counter := NewAtomicCounter(10)

		// Увеличиваем счетчик
		counter.TryIncrement()
		counter.TryIncrement()
		assert.Equal(t, int64(2), counter.Get())

		// Уменьшаем счетчик
		counter.Decrement()
		assert.Equal(t, int64(1), counter.Get())

		counter.Decrement()
		assert.Equal(t, int64(0), counter.Get())
	})

	t.Run("уменьшение ниже нуля", func(t *testing.T) {
		counter := NewAtomicCounter(10)

		// Уменьшаем счетчик ниже нуля
		counter.Decrement()
		assert.Equal(t, int64(-1), counter.Get())

		counter.Decrement()
		assert.Equal(t, int64(-2), counter.Get())
	})
}

func TestAtomicCounter_ConcurrentAccess(t *testing.T) {
	t.Run("конкурентное увеличение счетчика", func(t *testing.T) {
		max := int64(1000)
		counter := NewAtomicCounter(max)
		numGoroutines := 100
		incrementsPerGoroutine := 10

		var wg sync.WaitGroup
		successCount := int64(0)
		var successMutex sync.Mutex

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				localSuccess := int64(0)
				for j := 0; j < incrementsPerGoroutine; j++ {
					if counter.TryIncrement() {
						localSuccess++
					}
				}
				successMutex.Lock()
				successCount += localSuccess
				successMutex.Unlock()
			}()
		}

		wg.Wait()

		// Проверяем что счетчик не превысил максимум
		assert.True(t, counter.Get() <= max)
		assert.Equal(t, counter.Get(), successCount)
		assert.True(t, successCount <= max)
	})

	t.Run("конкурентное увеличение и уменьшение", func(t *testing.T) {
		counter := NewAtomicCounter(100)
		numGoroutines := 50

		var wg sync.WaitGroup

		// Горутины для увеличения
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					counter.TryIncrement()
				}
			}()
		}

		// Горутины для уменьшения
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 5; j++ {
					counter.Decrement()
				}
			}()
		}

		wg.Wait()

		// Проверяем что операции выполнились без гонок данных
		finalValue := counter.Get()
		assert.True(t, finalValue >= -250) // минимально возможное значение
		assert.True(t, finalValue <= 100)  // максимально возможное значение
	})

	t.Run("конкурентное достижение лимита", func(t *testing.T) {
		max := int64(10)
		counter := NewAtomicCounter(max)
		numGoroutines := 100

		var wg sync.WaitGroup
		successCount := int64(0)
		var successMutex sync.Mutex

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if counter.TryIncrement() {
					successMutex.Lock()
					successCount++
					successMutex.Unlock()
				}
			}()
		}

		wg.Wait()

		// Должно быть ровно max успешных увеличений
		assert.Equal(t, max, successCount)
		assert.Equal(t, max, counter.Get())
	})
}

func TestAtomicCounter_EdgeCases(t *testing.T) {
	t.Run("большие значения", func(t *testing.T) {
		max := int64(9223372036854775807) // максимальное значение int64
		counter := NewAtomicCounter(max)

		assert.Equal(t, int64(0), counter.Get())
		assert.Equal(t, max, counter.GetMax())

		// Должно успешно увеличиться
		result := counter.TryIncrement()
		assert.True(t, result)
		assert.Equal(t, int64(1), counter.Get())
	})

	t.Run("минимальные значения", func(t *testing.T) {
		max := int64(-9223372036854775808) // минимальное значение int64
		counter := NewAtomicCounter(max)

		assert.Equal(t, int64(0), counter.Get())
		assert.Equal(t, max, counter.GetMax())

		// Увеличение должно провалиться так как 0 > max
		result := counter.TryIncrement()
		assert.False(t, result)
		assert.Equal(t, int64(0), counter.Get())
	})
}

func TestAtomicCounter_Get(t *testing.T) {
	t.Run("получение текущего значения", func(t *testing.T) {
		counter := NewAtomicCounter(10)

		assert.Equal(t, int64(0), counter.Get())

		counter.TryIncrement()
		assert.Equal(t, int64(1), counter.Get())

		counter.TryIncrement()
		counter.TryIncrement()
		assert.Equal(t, int64(3), counter.Get())

		counter.Decrement()
		assert.Equal(t, int64(2), counter.Get())
	})
}

func TestAtomicCounter_GetMax(t *testing.T) {
	t.Run("получение максимального значения", func(t *testing.T) {
		testCases := []int64{0, 1, 10, 100, 1000, -1, -100}

		for _, max := range testCases {
			counter := NewAtomicCounter(max)
			assert.Equal(t, max, counter.GetMax())
		}
	})
}

func BenchmarkAtomicCounter_TryIncrement(b *testing.B) {
	counter := NewAtomicCounter(int64(b.N))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			counter.TryIncrement()
		}
	})
}

func BenchmarkAtomicCounter_Decrement(b *testing.B) {
	counter := NewAtomicCounter(1000000)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			counter.Decrement()
		}
	})
}

func BenchmarkAtomicCounter_Get(b *testing.B) {
	counter := NewAtomicCounter(1000000)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			counter.Get()
		}
	})
}
