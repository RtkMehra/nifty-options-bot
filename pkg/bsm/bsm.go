package bsm

import (
	"math"
)

func OptionPrice(S, K, r, sigma float64, T float64, optType string) float64 {
	d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / (sigma * math.Sqrt(T))
	d2 := d1 - sigma*math.Sqrt(T)

	if optType == "CE" {
		return S*normCDF(d1) - K*math.Exp(-r*T)*normCDF(d2)
	}
	return K*math.Exp(-r*T)*normCDF(-d2) - S*normCDF(-d1)
}

func Delta(S, K, r, sigma float64, T float64, optType string) float64 {
	d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / (sigma * math.Sqrt(T))
	if optType == "CE" {
		return normCDF(d1)
	}
	return normCDF(d1) - 1
}

func ImpliedVol(S, K, r float64, T float64, marketPrice float64, optType string) float64 {
	sigma := 0.15
	for i := 0; i < 100; i++ {
		price := OptionPrice(S, K, r, sigma, T, optType)
		diff := price - marketPrice
		if math.Abs(diff) < 1e-6 {
			return sigma
		}
		vega := vega(S, K, r, sigma, T)
		if vega < 1e-12 {
			break
		}
		sigma = sigma - diff/vega
		if sigma <= 0 {
			sigma = 0.01
		}
	}
	return sigma
}

func vega(S, K, r, sigma float64, T float64) float64 {
	d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / (sigma * math.Sqrt(T))
	return S * normPDF(d1) * math.Sqrt(T)
}

func normPDF(x float64) float64 {
	return math.Exp(-0.5*x*x) / math.Sqrt(2*math.Pi)
}

func normCDF(x float64) float64 {
	return 0.5 * math.Erfc(-x/math.Sqrt2)
}
