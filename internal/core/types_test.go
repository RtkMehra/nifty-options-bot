package core

import (
	"testing"
	"time"
)

func TestMidPrice(t *testing.T) {
	tests := []struct {
		name     string
		bid      float64
		ask      float64
		expected float64
	}{
		{name: "normal", bid: 180, ask: 190, expected: 185},
		{name: "zero bid", bid: 0, ask: 190, expected: 0},
		{name: "zero ask", bid: 180, ask: 0, expected: 0},
		{name: "both zero", bid: 0, ask: 0, expected: 0},
		{name: "same", bid: 185, ask: 185, expected: 185},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MidPrice(tt.bid, tt.ask)
			if got != tt.expected {
				t.Errorf("MidPrice(%v, %v) = %v, want %v", tt.bid, tt.ask, got, tt.expected)
			}
		})
	}
}

func TestDTE(t *testing.T) {
	tests := []struct {
		name     string
		expiry   time.Time
		expected int
	}{
		{
			name:     "future",
			expiry:   time.Now().Add(24 * 10 * time.Hour),
			expected: 10,
		},
		{
			name:     "today",
			expiry:   time.Now().Add(time.Hour),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DTE(tt.expiry)
			if got != tt.expected {
				t.Errorf("DTE() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRegimeValues(t *testing.T) {
	if RegimeBuyIV != 0 {
		t.Errorf("RegimeBuyIV should be 0, got %v", RegimeBuyIV)
	}
	if RegimeSellIV != 1 {
		t.Errorf("RegimeSellIV should be 1, got %v", RegimeSellIV)
	}
	if RegimeNeutral != 2 {
		t.Errorf("RegimeNeutral should be 2, got %v", RegimeNeutral)
	}
}

func TestOptionTypeStrings(t *testing.T) {
	if string(CE) != "CE" {
		t.Errorf("CE string = %v, want CE", string(CE))
	}
	if string(PE) != "PE" {
		t.Errorf("PE string = %v, want PE", string(PE))
	}
}
