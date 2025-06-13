package giftBuyer

import (
	"math/rand"
	"time"
)

var fastRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// selectRandomElementFast - максимально быстрый выбор случайного элемента
func selectRandomElementFast[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	return slice[fastRand.Intn(len(slice))]
}
