"""Fetch NSE option chain data using pnsea with Akamai bypass."""
import json
import sys
from pnsea import NSE


def main():
    try:
        nse = NSE()
        df, expiries, spot = nse.options.option_chain("NIFTY")
        chain = []
        for _, row in df.iterrows():
            def to_float(v):
                if v is None or v == "-" or v == "":
                    return 0.0
                try:
                    return float(v)
                except (ValueError, TypeError):
                    return 0.0

            def to_int(v):
                if v is None or v == "-" or v == "":
                    return 0
                try:
                    return int(v)
                except (ValueError, TypeError):
                    return 0

            strike = int(row.get("strikePrice", 0))
            expiry = None
            if expiries and len(expiries) > 0:
                expiry = str(expiries[0])

            ce = {
                "strike": strike,
                "expiry": expiry,
                "type": "CE",
                "ltp": to_float(row.get("CE_lastPrice", 0)),
                "iv": to_float(row.get("CE_impliedVolatility", 0)) / 100.0,
                "oi": to_int(row.get("CE_openInterest", 0)),
                "volume": to_int(row.get("CE_totalTradedVolume", 0)),
                "bid": to_float(row.get("CE_bidprice", 0)),
                "ask": to_float(row.get("CE_askPrice", 0)),
                "bid_qty": to_int(row.get("CE_bidQty", 0)),
                "ask_qty": to_int(row.get("CE_askQty", 0)),
            }
            pe = {
                "strike": strike,
                "expiry": expiry,
                "type": "PE",
                "ltp": to_float(row.get("PE_lastPrice", 0)),
                "iv": to_float(row.get("PE_impliedVolatility", 0)) / 100.0,
                "oi": to_int(row.get("PE_openInterest", 0)),
                "volume": to_int(row.get("PE_totalTradedVolume", 0)),
                "bid": to_float(row.get("PE_bidprice", 0)),
                "ask": to_float(row.get("PE_askPrice", 0)),
                "bid_qty": to_int(row.get("PE_bidQty", 0)),
                "ask_qty": to_int(row.get("PE_askQty", 0)),
            }
            chain.append(ce)
            chain.append(pe)

        vix = 0.0
        try:
            vix_data = nse.equity.find_index("INDIA VIX")
            if vix_data and "last" in vix_data:
                vix = float(vix_data["last"])
        except Exception:
            pass

        output = {
            "spot": float(spot) if spot else 0.0,
            "vix": vix,
            "expiries": [str(e) for e in expiries] if expiries else [],
            "chain": chain,
            "expiry": str(expiries[0]) if expiries and len(expiries) > 0 else "",
        }
        print(json.dumps(output))
        return 0
    except Exception as e:
        print(json.dumps({"error": str(e)}), file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
