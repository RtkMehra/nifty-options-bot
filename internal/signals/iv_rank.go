package signals

import "math"

func CalcIVRank(currentVIX, low52Week, high52Week float64) float64 {
	range_ := high52Week - low52Week
	if range_ <= 0 {
		return 50
	}
	return (currentVIX - low52Week) / range_ * 100
}

func CalcIVZScore(currentIV float64, history []float64) float64 {
	n := len(history)
	if n < 2 {
		return 0
	}

	sum := 0.0
	for _, v := range history {
		sum += v
	}
	mean := sum / float64(n)

	variance := 0.0
	for _, v := range history {
		diff := v - mean
		variance += diff * diff
	}
	stddev := math.Sqrt(variance / float64(n))

	if stddev < 1e-10 {
		return 0
	}

	return (currentIV - mean) / stddev
}
