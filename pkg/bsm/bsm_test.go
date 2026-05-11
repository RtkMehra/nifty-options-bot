package bsm

import (
	"math"
	"testing"
)

func TestVega(t *testing.T) {
	S := 22500.0
	K := 22500.0
	r := 0.07
	sigma := 0.13
	T := 10.0 / 365.0

	v := vega(S, K, r, sigma, T)

	if v <= 0 {
		t.Errorf("vega should be positive, got %v", v)
	}
}

func TestNormPDF(t *testing.T) {
	tests := []struct {
		x        float64
		expected float64
	}{
		{x: 0, expected: 0.3989},
		{x: 1, expected: 0.2420},
		{x: -1, expected: 0.2420},
		{x: 2, expected: 0.0540},
	}

	for _, tt := range tests {
		got := normPDF(tt.x)
		if math.Abs(got-tt.expected) > 0.001 {
			t.Errorf("normPDF(%v) = %v, want %v", tt.x, got, tt.expected)
		}
	}
}

func TestImpliedVolZeroVega(t *testing.T) {
	S := 22500.0
	K := 22500.0
	r := 0.07
	T := 0.0001

	price := OptionPrice(S, K, r, 0.15, T, "CE")
	iv := ImpliedVol(S, K, r, T, price, "CE")

	if iv <= 0 {
		t.Errorf("IV should be positive even near expiry, got %v", iv)
	}
}

func TestOptionPriceExtreme(t *testing.T) {
	S := 22500.0
	K := 22500.0
	r := 0.07
	sigma := 0.0
	T := 10.0 / 365.0

	price := OptionPrice(S, K, r, sigma, T, "CE")
	if price < 0 {
		t.Errorf("price with zero IV should not be negative, got %v", price)
	}
}

func TestDeltaExtremeOTM(t *testing.T) {
	S := 20000.0
	K := 25000.0
	r := 0.07
	sigma := 0.13
	T := 10.0 / 365.0

	delta := Delta(S, K, r, sigma, T, "CE")
	if delta < 0 {
		t.Errorf("OTM call delta should be >= 0, got %v", delta)
	}

	delta = Delta(S, K, r, sigma, T, "PE")
	if delta > 0 {
		t.Errorf("OTM put delta should be <= 0, got %v", delta)
	}
}

func TestOptionPriceITMvsOTM(t *testing.T) {
	r := 0.07
	sigma := 0.13
	T := 10.0 / 365.0

	callITM := OptionPrice(23000.0, 22500.0, r, sigma, T, "CE")
	callOTM := OptionPrice(22000.0, 22500.0, r, sigma, T, "CE")

	if callITM <= callOTM {
		t.Errorf("ITM call (%v) should be more expensive than OTM call (%v)", callITM, callOTM)
	}

	putITM := OptionPrice(22000.0, 22500.0, r, sigma, T, "PE")
	putOTM := OptionPrice(23000.0, 22500.0, r, sigma, T, "PE")

	if putITM <= putOTM {
		t.Errorf("ITM put (%v) should be more expensive than OTM put (%v)", putITM, putOTM)
	}
}
