package signals

import (
	"math"
	"testing"
)

func TestExpectedMove(t *testing.T) {
	// Nifty=22500, IV=13%, DTE=10
	// EM = 22500 * 0.13 * sqrt(10/365) = ~383
	spot := 22500.0
	iv := 0.13
	dte := 10

	em := ExpectedMove(spot, iv, dte)
	expected := 22500 * 0.13 * math.Sqrt(10.0/365.0)

	if math.Abs(em-expected) > 1 {
		t.Errorf("ExpectedMove() = %v, want %v", em, expected)
	}
}

func TestExpectedMoveZeroDTE(t *testing.T) {
	em := ExpectedMove(22500, 0.13, 0)
	if em != 0 {
		t.Errorf("ExpectedMove(0 DTE) = %v, want 0", em)
	}
}

func TestExpectedMoveRange(t *testing.T) {
	lower, upper := ExpectedMoveRange(22500, 0.13, 10)
	if lower >= upper {
		t.Errorf("lower %v should be < upper %v", lower, upper)
	}
	if lower >= 22500 || upper <= 22500 {
		t.Errorf("range should straddle spot: lower=%v upper=%v spot=22500", lower, upper)
	}
}

func TestTargetStrike(t *testing.T) {
	em := ExpectedMove(22500, 0.13, 10)
	strike := TargetStrike(22500, em, true, 50)
	if strike <= 22500 {
		t.Errorf("bullish target strike %v should be above spot", strike)
	}

	strike = TargetStrike(22500, em, false, 50)
	if strike >= 22500 {
		t.Errorf("bearish target strike %v should be below spot", strike)
	}
}
