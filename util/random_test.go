package util

import (
	"testing"
)

func TestRandomIntRange(t *testing.T) {
	minVal, maxVal := int64(5), int64(15)
	for i := 0; i < 100; i++ {
		v := RandomInt(minVal, maxVal)
		if v < minVal || v > maxVal {
			t.Fatalf("value %d out of range [%d,%d]", v, minVal, maxVal)
		}
	}
	// Equal bounds
	if v := RandomInt(7, 7); v != 7 {
		t.Fatalf("expected 7 got %d", v)
	}
}

func TestRandomIntPanicInvalidRange(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for invalid range")
		}
	}()
	_ = RandomInt(10, 5) // max < min should panic
}

func TestRandomStringLength(t *testing.T) {
	n := 12
	s := RandomString(n)
	if len(s) != n {
		t.Fatalf("expected length %d got %d", n, len(s))
	}
	if s == RandomString(n) {
		// Extremely unlikely; treat as failure indicating non-randomness
		t.Fatalf("two consecutive RandomString calls produced identical result: %s", s)
	}
}

func TestRandomOwner(t *testing.T) {
	o := RandomOwner()
	if len(o) != 6 {
		t.Fatalf("expected owner length 6 got %d", len(o))
	}
}

func TestRandomMoneyRange(t *testing.T) {
	for i := 0; i < 100; i++ {
		m := RandomMoney()
		if m < 0 || m > 1000 {
			t.Fatalf("money %d out of range [0,1000]", m)
		}
	}
}

func TestRandomCurrencyMembership(t *testing.T) {
	allowed := map[string]bool{"USD": true, "EUR": true, "ILS": true}
	for i := 0; i < 50; i++ {
		c := RandomCurrency()
		if !allowed[c] {
			t.Fatalf("unexpected currency %s", c)
		}
	}
}
