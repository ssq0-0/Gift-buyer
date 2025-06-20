package utils

import (
	"crypto/rand"
	"encoding/binary"
	mathRand "math/rand"
	"strings"
	"time"
)

func cryptoSeed() int64 {
	var seed int64
	if err := binary.Read(rand.Reader, binary.BigEndian, &seed); err != nil {
		return time.Now().UnixNano()
	}
	return seed
}

var fastRand = mathRand.New(mathRand.NewSource(cryptoSeed()))

// selectRandomElementFast - максимально быстрый выбор случайного элемента
func SelectRandomElementFast[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	return slice[fastRand.Intn(len(slice))]
}

func RandString5(lenght int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var sb strings.Builder
	sb.Grow(lenght)

	for i := 0; i < lenght; i++ {
		sb.WriteByte(letters[fastRand.Intn(len(letters))])
	}
	return sb.String()
}
