package signals

import "math"

func ExpectedMove(spot, iv float64, dte int) float64 {
	T := float64(dte) / 365.0
	if T <= 0 {
		return 0
	}
	return spot * iv * math.Sqrt(T)
}

func ExpectedMoveRange(spot, iv float64, dte int) (lower, upper float64) {
	em := ExpectedMove(spot, iv, dte)
	return spot - em, spot + em
}

func TargetStrike(spot float64, em float64, isBullish bool, strikeStep int) int {
	if strikeStep <= 0 {
		strikeStep = 50
	}
	var target float64
	if isBullish {
		target = spot + em*0.5
	} else {
		target = spot - em*0.5
	}
	rounded := math.Round(target/float64(strikeStep)) * float64(strikeStep)
	return int(rounded)
}
