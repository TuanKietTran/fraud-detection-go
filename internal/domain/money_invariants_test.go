package domain

import (
	"errors"
	"fmt"
	"math"
	"testing"
)

func TestMoneyRejectsMalformedTextWithoutPartialValue(t *testing.T) {
	for _, text := range []string{
		"", " 1", "1 ", "\t1", "1\n", "+1", "-0", "-0.01",
		".01", "1.2.3", "1..0", "00", "01.00", "1.", "1.001",
		"one.00", "1e2", "1_0", "NaN", "１.00", "1.a", "1.0a", "1.１",
		"9223372036854775808", "92233720368547758.08", "92233720368547759.00",
	} {
		t.Run(fmt.Sprintf("%q", text), func(t *testing.T) {
			value, err := ParseMinorUnits(text, "USD")
			if value != 0 || !errors.Is(err, ErrInvalidMoney) {
				t.Fatalf("accepted/partially parsed %q: %d %v", text, value, err)
			}
		})
	}
	for _, currency := range []string{"", "EUR", "usd", "US", "USDD"} {
		value, err := ParseMinorUnits("1.00", currency)
		if value != 0 || !errors.Is(err, ErrUnsupportedCurrency) {
			t.Fatalf("unsupported %q: %d %v", currency, value, err)
		}
	}
}

func TestMoneyExactScaleAndInt64Boundaries(t *testing.T) {
	for _, tc := range []struct {
		text  string
		cents int64
	}{
		{"0", 0}, {"0.00", 0}, {"0.01", 1}, {"0.1", 10}, {"0.10", 10}, {"1", 100}, {"1.2", 120}, {"1.20", 120}, {"10.09", 1009},
		{"92233720368547758", 9223372036854775800}, {"92233720368547758.06", math.MaxInt64 - 1}, {"92233720368547758.07", math.MaxInt64},
	} {
		got, err := ParseMinorUnits(tc.text, "USD")
		if err != nil || got != tc.cents {
			t.Fatalf("%q = %d %v, want %d", tc.text, got, err, tc.cents)
		}
	}
	// Integer-only canonical round trips across boundaries and deterministic broad samples.
	check := func(cents int64) {
		t.Helper()
		text := fmt.Sprintf("%d.%02d", cents/100, cents%100)
		got, err := ParseMinorUnits(text, "USD")
		if err != nil || got != cents {
			t.Fatalf("round trip %d -> %q -> %d %v", cents, text, got, err)
		}
	}
	for _, cents := range []int64{0, 1, 9, 10, 99, 100, 101, 500000, 1000001, math.MaxInt64 - 1, math.MaxInt64} {
		check(cents)
	}
	seed := uint64(17)
	for i := 0; i < 1000; i++ {
		seed = seed*6364136223846793005 + 1
		check(int64(seed >> 1))
	}
}
