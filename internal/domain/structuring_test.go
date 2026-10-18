package domain

import (
	"math"
	"testing"
)

func TestStructuringExactBandAndCountBoundaries(t *testing.T) {
	for _, tc := range []struct {
		amount, band int64
		prior        int
		want         bool
	}{
		{89999, 100000, 2, false}, {90000, 100000, 2, true}, {100000, 100000, 2, true}, {100001, 100000, 2, false},
		{90000, 100000, 1, false}, {90000, 100000, 3, true}, {90000, 100000, -1, false},
		{0, 100000, 3, false}, {-1, 100000, 3, false}, {1, 0, 3, false}, {1, -1, 3, false},
		{1, 1, 2, true}, {9, 11, 2, false}, {10, 11, 2, true}, {math.MaxInt64, math.MaxInt64, 2, true},
	} {
		if got := InferredStructuring(tc.amount, tc.band, tc.prior); got != tc.want {
			t.Errorf("%+v: %v", tc, got)
		}
	}
}

func FuzzStructuringBand(f *testing.F) {
	f.Add(int64(90000), int64(100000), 2)
	f.Add(int64(math.MaxInt64), int64(math.MaxInt64), 3)
	f.Fuzz(func(t *testing.T, amount, band int64, prior int) {
		hit := InferredStructuring(amount, band, prior)
		if hit && (amount <= 0 || band <= 0 || amount > band || prior < 2 || amount < band-band/10) {
			t.Fatal("invalid pattern classified")
		}
		if band > 0 && !NearStructuringBand(band, band) {
			t.Fatal("upper boundary excluded")
		}
	})
}
