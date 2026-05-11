package signals

import (
	"math"
	"testing"
)

func TestConvictionScore_FullSignal(t *testing.T) {
	d := DirectionalSignals{
		BOSConfirmed:   true,
		OBRetest:       true,
		FVGFilled:      true,
		EMAPosition:    1,
		PCR:            1.5,
		NearKeySupport: true,
	}

	score := ConvictionScore(d)
	if score < 0.8 {
		t.Errorf("full signal score = %v, want > 0.8", score)
	}
	if score > 1.0 {
		t.Errorf("score > 1.0: %v", score)
	}
}

func TestConvictionScore_NoSignal(t *testing.T) {
	d := DirectionalSignals{
		PCR: 1.0,
	}

	score := ConvictionScore(d)
	if score != 0 {
		t.Errorf("no signal score = %v, want 0", score)
	}
}

func TestConvictionScore_Partial(t *testing.T) {
	d := DirectionalSignals{
		BOSConfirmed: true,
		OBRetest:     true,
		PCR:          1.0,
	}

	score := ConvictionScore(d)
	expected := 0.30 + 0.25
	if math.Abs(score-expected) > 0.01 {
		t.Errorf("BOS+OB score = %v, want %v", score, expected)
	}
}

func TestResolveDirection_Bullish(t *testing.T) {
	d := DirectionalSignals{
		BOSConfirmed: true,
		EMAPosition:  1,
		PCR:          1.5,
	}

	dir := ResolveDirection(d)
	if dir != 0 {
		t.Errorf("direction = %v, want 0 (bullish)", dir)
	}
}

func TestResolveDirection_Bearish(t *testing.T) {
	d := DirectionalSignals{
		BOSConfirmed: true,
		EMAPosition:  -1,
		PCR:          0.5,
	}

	dir := ResolveDirection(d)
	if dir != 1 {
		t.Errorf("direction = %v, want 1 (bearish)", dir)
	}
}

func TestResolveDirection_Neutral(t *testing.T) {
	d := DirectionalSignals{
		PCR: 1.0,
	}

	dir := ResolveDirection(d)
	_ = dir
}

func TestPCRScore(t *testing.T) {
	tests := []struct {
		pcr      float64
		expected float64
	}{
		{pcr: 1.5, expected: 1.0},
		{pcr: 1.2, expected: 0.5},
		{pcr: 1.0, expected: 0},
		{pcr: 0.9, expected: 0},
		{pcr: 0.6, expected: 1.0},
		{pcr: 0.75, expected: 0.5},
		{pcr: 0, expected: 0},
	}

	for _, tt := range tests {
		got := calcPCRScore(tt.pcr)
		if got != tt.expected {
			t.Errorf("calcPCRScore(%v) = %v, want %v", tt.pcr, got, tt.expected)
		}
	}
}
