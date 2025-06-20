// Package giftBuyer provides gift purchasing functionality for the gift buying system.
package atomicCounter

import (
	"sync/atomic"
)

// atomicCounter provides thread-safe counting with configurable maximum limits.
// It uses atomic operations to ensure concurrent safety when tracking purchase counts
// and enforcing maximum purchase limits across multiple goroutines.
type atomicCounter struct {
	// count stores the current count value using atomic operations
	count int64

	// max defines the maximum allowed count value
	max int64
}

// newAtomicCounter creates a new atomic counter with the specified maximum limit.
//
// Parameters:
//   - max: maximum count value allowed
//
// Returns:
//   - *atomicCounter: initialized counter instance
func NewAtomicCounter(max int64) *atomicCounter {
	return &atomicCounter{
		count: 0,
		max:   max,
	}
}

// TryIncrement attempts to increment the counter if it hasn't reached the maximum.
// This operation is atomic and thread-safe, making it suitable for concurrent use.
//
// Returns:
//   - bool: true if increment was successful, false if maximum limit reached
func (ac *atomicCounter) TryIncrement() bool {
	for {
		current := atomic.LoadInt64(&ac.count)
		if current >= ac.max {
			return false
		}
		if atomic.CompareAndSwapInt64(&ac.count, current, current+1) {
			return true
		}
	}
}

// Decrement decreases the counter by one.
// This operation is atomic and thread-safe.
func (ac *atomicCounter) Decrement() {
	atomic.AddInt64(&ac.count, -1)
}

// Get returns the current count value.
// This operation is atomic and thread-safe.
//
// Returns:
//   - int64: current count value
func (ac *atomicCounter) Get() int64 {
	return atomic.LoadInt64(&ac.count)
}

// GetMax returns the maximum allowed count value.
//
// Returns:
//   - int64: maximum count limit
func (ac *atomicCounter) GetMax() int64 {
	return ac.max
}
