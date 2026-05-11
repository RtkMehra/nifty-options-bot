package signals

import (
	"math"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

type DirectionalSignals struct {
	BOSConfirmed   bool
	OBRetest       bool
	FVGFilled      bool
	EMAPosition    int
	PCR            float64
	NearKeySupport bool
	SpotPrice      float64
	ATMStrike      int
}

func ConvictionScore(d DirectionalSignals) float64 {
	weights := map[string]float64{
		"BOS":     0.30,
		"OB":      0.25,
		"FVG":     0.15,
		"EMA":     0.15,
		"PCR":     0.10,
		"Support": 0.05,
	}

	score := 0.0
	if d.BOSConfirmed {
		score += weights["BOS"]
	}
	if d.OBRetest {
		score += weights["OB"]
	}
	if d.FVGFilled {
		score += weights["FVG"]
	}
	if d.EMAPosition != 0 {
		score += weights["EMA"]
	}

	pcrScore := calcPCRScore(d.PCR)
	score += pcrScore * weights["PCR"]

	if d.NearKeySupport {
		score += weights["Support"]
	}

	return math.Min(score, 1.0)
}

func ResolveDirection(d DirectionalSignals) core.Direction {
	bullish := 0
	bearish := 0

	if d.BOSConfirmed && d.EMAPosition > 0 {
		bullish += 2
	}
	if d.BOSConfirmed && d.EMAPosition < 0 {
		bearish += 2
	}
	if d.OBRetest {
		if d.EMAPosition > 0 {
			bullish++
		} else {
			bearish++
		}
	}
	if d.FVGFilled {
		if d.EMAPosition > 0 {
			bullish++
		} else {
			bearish++
		}
	}
	if d.PCR > 1.3 {
		bullish++
	} else if d.PCR < 0.7 {
		bearish++
	}

	if bullish > bearish {
		return core.DirectionBullish
	} else if bearish > bullish {
		return core.DirectionBearish
	}
	return core.DirectionNeutral
}

func calcPCRScore(pcr float64) float64 {
	if pcr <= 0 {
		return 0
	}
	if pcr >= 1.3 {
		return 1.0
	} else if pcr >= 1.1 {
		return 0.5
	} else if pcr <= 0.7 {
		return 1.0
	} else if pcr <= 0.85 {
		return 0.5
	}
	return 0
}
