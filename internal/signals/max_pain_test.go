package signals

import (
	"math"
	"testing"
	"time"

	"github.com/ritikmehra/nifty-options-bot/internal/core"
)

func makeOption(strike int, expiry time.Time, optType core.OptionType, oi int64, iv float64) core.OptionData {
	return core.OptionData{
		Strike:     strike,
		Expiry:     expiry,
		OptionType: optType,
		OI:         oi,
		IV:         iv,
	}
}

func TestMaxPain_Empty(t *testing.T) {
	mp := MaxPain(nil)
	if mp != 0 {
		t.Errorf("MaxPain(nil) = %v, want 0", mp)
	}
}

func TestMaxPain_SingleStrike(t *testing.T) {
	expiry := time.Now().AddDate(0, 0, 14)
	chain := []core.OptionData{
		makeOption(22500, expiry, core.CE, 100000, 0.13),
		makeOption(22500, expiry, core.PE, 120000, 0.14),
	}

	mp := MaxPain(chain)
	if mp == 0 {
		t.Errorf("MaxPain() = 0, expected non-zero")
	}
}

func TestMaxPain_MultipleStrikes(t *testing.T) {
	expiry := time.Now().AddDate(0, 0, 14)
	chain := []core.OptionData{
		makeOption(22400, expiry, core.CE, 50000, 0.13),
		makeOption(22400, expiry, core.PE, 60000, 0.14),
		makeOption(22500, expiry, core.CE, 100000, 0.13),
		makeOption(22500, expiry, core.PE, 120000, 0.14),
		makeOption(22600, expiry, core.CE, 30000, 0.12),
		makeOption(22600, expiry, core.PE, 40000, 0.15),
	}

	mp := MaxPain(chain)
	if mp != 22500 {
		t.Errorf("MaxPain() = %v, want 22500 (highest OI)", mp)
	}
}

func TestCalcSkew(t *testing.T) {
	expiry := time.Now().AddDate(0, 0, 14)
	chain := []core.OptionData{
		makeOption(22500, expiry, core.CE, 0, 0.13),
		makeOption(22500, expiry, core.PE, 0, 0.15),
	}

	skew := calcSkew(chain, 22500)
	expected := (0.15 - 0.13) * 100
	if math.Abs(skew-expected) > 0.01 {
		t.Errorf("calcSkew() = %v, want %v", skew, expected)
	}
}
