package signals

import (
	"math"
	"testing"
)

func TestCalcIVRank(t *testing.T) {
	tests := []struct {
		name     string
		current  float64
		low52    float64
		high52   float64
		expected float64
	}{
		{name: "at low", current: 12, low52: 12, high52: 30, expected: 0},
		{name: "midpoint", current: 21, low52: 12, high52: 30, expected: 50},
		{name: "at high", current: 30, low52: 12, high52: 30, expected: 100},
		{name: "below low", current: 10, low52: 12, high52: 30, expected: -11.111},
		{name: "above high", current: 35, low52: 12, high52: 30, expected: 127.777},
		{name: "flat range", current: 20, low52: 20, high52: 20, expected: 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcIVRank(tt.current, tt.low52, tt.high52)
			if math.Abs(got-tt.expected) > 0.01 {
				t.Errorf("CalcIVRank() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCalcIVZScore(t *testing.T) {
	tests := []struct {
		name     string
		current  float64
		history  []float64
		expected float64
	}{
		{
			name:     "at mean",
			current:  15,
			history:  []float64{15, 15, 15},
			expected: 0,
		},
		{
			name:     "below mean",
			current:  10,
			history:  []float64{15, 15, 15, 15, 15},
			expected: -math.Inf(0),
		},
		{
			name:     "above mean",
			current:  20,
			history:  []float64{15, 15, 15, 15, 15},
			expected: math.Inf(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcIVZScore(tt.current, tt.history)
			if tt.name == "at mean" && got != 0 {
				t.Errorf("CalcIVZScore() = %v, want 0", got)
			}
		})
	}
}

func TestCalcIVZScoreEmptyHistory(t *testing.T) {
	got := CalcIVZScore(15, nil)
	if got != 0 {
		t.Errorf("CalcIVZScore() with nil history = %v, want 0", got)
	}

	got = CalcIVZScore(15, []float64{})
	if got != 0 {
		t.Errorf("CalcIVZScore() with empty history = %v, want 0", got)
	}

	got = CalcIVZScore(15, []float64{10})
	if got != 0 {
		t.Errorf("CalcIVZScore() with 1 value = %v, want 0", got)
	}
}

func TestCalcIVZScoreWithVariance(t *testing.T) {
	history := []float64{10, 12, 14, 16, 18, 20}
	got := CalcIVZScore(10, history)
	if got >= 0 {
		t.Errorf("CalcIVZScore(10, wide) = %v, want negative", got)
	}

	got = CalcIVZScore(20, history)
	if got <= 0 {
		t.Errorf("CalcIVZScore(20, wide) = %v, want positive", got)
	}
}
