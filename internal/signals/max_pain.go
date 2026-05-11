package signals

import (
	"math"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

func MaxPain(chain []core.OptionData) int {
	if len(chain) == 0 {
		return 0
	}

	strikeMap := make(map[int]float64)
	for _, opt := range chain {
		if core.DTE(opt.Expiry) == 0 {
			continue
		}
		loss := float64(opt.OI) * math.Abs(float64(opt.Strike))
		strikeMap[opt.Strike] += loss
	}

	maxPainStrike := 0
	maxPainValue := 0.0
	for strike, pain := range strikeMap {
		if pain > maxPainValue {
			maxPainValue = pain
			maxPainStrike = strike
		}
	}

	return maxPainStrike
}

func calcSkew(chain []core.OptionData, atmStrike int) float64 {
	var atmPutIV, atmCallIV float64
	for _, opt := range chain {
		if opt.Strike != atmStrike {
			continue
		}
		if opt.OptionType == core.CE && opt.IV > 0 {
			atmCallIV = opt.IV
		}
		if opt.OptionType == core.PE && opt.IV > 0 {
			atmPutIV = opt.IV
		}
	}
	if atmCallIV <= 0 {
		return 0
	}
	return (atmPutIV - atmCallIV) * 100
}
