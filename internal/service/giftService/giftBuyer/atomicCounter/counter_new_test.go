package atomicCounter

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAtomicCounterExtended(t *testing.T) {
	t.Run("создание нового счетчика", func(t *testing.T) {
		maxCount := int64(100)
		counter := NewAtomicCounter(maxCount)

		assert.NotNil(t, counter)
		assert.Equal(t, int64(0), counter.Get())
		assert.Equal(t, maxCount, counter.GetMax())
	})

	t.Run("создание счетчика с нулевым максимумом", func(t *testing.T) {
		counter := NewAtomicCounter(0)

		assert.NotNil(t, counter)
		assert.Equal(t, int64(0), counter.Get())
		assert.Equal(t, int64(0), counter.GetMax())
		assert.False(t, counter.TryIncrement()) // Нельзя увеличить
	})
}

func TestAtomicCounter_Operations(t *testing.T) {
	t.Run("базовые операции с счетчиком", func(t *testing.T) {
		counter := NewAtomicCounter(5)

		// Начальное значение
		assert.Equal(t, int64(0), counter.Get())

		// Успешное увеличение
		assert.True(t, counter.TryIncrement())
		assert.Equal(t, int64(1), counter.Get())

		// Еще несколько увеличений
		assert.True(t, counter.TryIncrement())
		assert.True(t, counter.TryIncrement())
		assert.Equal(t, int64(3), counter.Get())

		// Уменьшение
		counter.Decrement()
		assert.Equal(t, int64(2), counter.Get())

		// Уменьшение до нуля и ниже
		counter.Decrement()
		counter.Decrement()
		assert.Equal(t, int64(0), counter.Get())

		counter.Decrement()
		assert.Equal(t, int64(-1), counter.Get())
	})

	t.Run("достижение максимума", func(t *testing.T) {
		counter := NewAtomicCounter(2)

		// Увеличиваем до максимума
		assert.True(t, counter.TryIncrement())
		assert.Equal(t, int64(1), counter.Get())

		assert.True(t, counter.TryIncrement())
		assert.Equal(t, int64(2), counter.Get())

		// Попытка превысить максимум
		assert.False(t, counter.TryIncrement())
		assert.Equal(t, int64(2), counter.Get())

		// Еще одна попытка
		assert.False(t, counter.TryIncrement())
		assert.Equal(t, int64(2), counter.Get())
	})

	t.Run("уменьшение после достижения максимума", func(t *testing.T) {
		counter := NewAtomicCounter(2)

		// Достигаем максимума
		counter.TryIncrement()
		counter.TryIncrement()
		assert.Equal(t, int64(2), counter.Get())
		assert.False(t, counter.TryIncrement())

		// Уменьшаем и снова увеличиваем
		counter.Decrement()
		assert.Equal(t, int64(1), counter.Get())
		assert.True(t, counter.TryIncrement())
		assert.Equal(t, int64(2), counter.Get())
	})
}

func TestAtomicCounter_Concurrent(t *testing.T) {
	t.Run("конкурентные увеличения", func(t *testing.T) {
		maxCount := int64(100)
		counter := NewAtomicCounter(maxCount)
		numGoroutines := 50
		incrementsPerGoroutine := 10

		var wg sync.WaitGroup
		successCount := int64(0)
		var successMutex sync.Mutex

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				localSuccess := 0
				for j := 0; j < incrementsPerGoroutine; j++ {
					if counter.TryIncrement() {
						localSuccess++
					}
				}
				successMutex.Lock()
				successCount += int64(localSuccess)
				successMutex.Unlock()
			}()
		}

		wg.Wait()

		// Проверяем что количество успешных увеличений равно значению счетчика
		assert.Equal(t, successCount, counter.Get())
		// И не превышает максимум
		assert.True(t, counter.Get() <= maxCount)
		// Фактически должно быть ровно максимум (так как пытаемся больше)
		assert.Equal(t, maxCount, counter.Get())
	})

	t.Run("конкурентные увеличения и уменьшения", func(t *testing.T) {
		counter := NewAtomicCounter(50)
		numGoroutines := 20

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
		for i := 0; i < numGoroutines/2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 5; j++ {
					counter.Decrement()
				}
			}()
		}

		wg.Wait()

		// Проверяем что счетчик в разумных пределах
		finalValue := counter.Get()
		assert.True(t, finalValue >= -50) // Может быть отрицательным из-за Decrement
		assert.True(t, finalValue <= counter.GetMax())
	})

	t.Run("конкурентное чтение", func(t *testing.T) {
		counter := NewAtomicCounter(10)
		numReaders := 100

		// Устанавливаем начальное значение
		counter.TryIncrement()
		counter.TryIncrement()
		counter.TryIncrement()
		expectedValue := counter.Get()

		var wg sync.WaitGroup
		results := make([]int64, numReaders)

		for i := 0; i < numReaders; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				results[index] = counter.Get()
			}(i)
		}

		wg.Wait()

		// Все чтения должны вернуть одинаковое значение
		for i, value := range results {
			assert.Equal(t, expectedValue, value, "Reader %d got different value", i)
		}
	})
}

func TestAtomicCounter_AdvancedEdgeCases(t *testing.T) {
	t.Run("максимум равен единице", func(t *testing.T) {
		counter := NewAtomicCounter(1)

		assert.True(t, counter.TryIncrement())
		assert.Equal(t, int64(1), counter.Get())
		assert.False(t, counter.TryIncrement())
		assert.Equal(t, int64(1), counter.Get())

		counter.Decrement()
		assert.Equal(t, int64(0), counter.Get())
		assert.True(t, counter.TryIncrement())
	})

	t.Run("отрицательный максимум", func(t *testing.T) {
		counter := NewAtomicCounter(-5)

		// С отрицательным максимумом нельзя увеличивать
		assert.False(t, counter.TryIncrement())
		assert.Equal(t, int64(0), counter.Get())

		// Но можно уменьшать
		counter.Decrement()
		assert.Equal(t, int64(-1), counter.Get())
	})

	t.Run("большие значения", func(t *testing.T) {
		maxValue := int64(1000000)
		counter := NewAtomicCounter(maxValue)

		// Увеличиваем на большое количество
		successCount := 0
		for i := 0; i < int(maxValue)+100; i++ {
			if counter.TryIncrement() {
				successCount++
			}
		}

		assert.Equal(t, int(maxValue), successCount)
		assert.Equal(t, maxValue, counter.Get())
		assert.Equal(t, maxValue, counter.GetMax())
	})
}
