package util

import (
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Rng returns a new random number generator seeded with the current time.
func Rng() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// RandomInt returns a random integer between min and max inclusive.
func RandomInt(min, max int64) int64 {
	rng := Rng()
	return min + rng.Int63n(max-min+1)
}

// RandomString returns a random string of length n.
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)
	rng := Rng()

	for i := 0; i < n; i++ {
		c := alphabet[rng.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// RandomOwner returns a random owner name.
func RandomOwner() string {
	return RandomString(6)
}

// RandomMoney returns a random amount of money.
func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

// RandomCurrency returns a random currency from the list
func RandomCurrency() string {
	currencies := []string{EUR, USD, ILS}
	n := len(currencies)
	rng := Rng()

	return currencies[rng.Intn(n)]
}
