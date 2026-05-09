# Go Coding Standards

## Formatting & Style

- **Formatting**: `gofmt` enforced. No exceptions.
- **Line length**: Soft limit 100 chars, hard limit 120 chars
- **Imports**: Grouped as stdlib → third-party → internal, separated by blank lines
- **Naming**: `camelCase` for unexported, `PascalCase` for exported. No snake_case. No Hungarian notation.
- **Acronyms**: `IV`, `API`, `HTTP`, `DB`, `URL` — all caps or all lower. Not `Ivr`, `Api`, `Http`

## Naming Conventions

### Files
- `snake_case.go` (Go convention)
- One primary type per file, named after the file
- Test files: `*_test.go` (same directory, `package xxx_test` for black-box tests)

### Identifiers
| Scope | Convention | Example |
|-------|------------|---------|
| Exported type | PascalCase | `MarketSnapshot` |
| Unexported type | camelCase | `marketSnapshot` |
| Exported function | PascalCase | `CalcIVRank()` |
| Unexported function | camelCase | `fetchOptionChain()` |
| Interface | PascalCase + "er" suffix | `DataFetcher`, `SignalProvider` |
| Receiver | 1–2 letter abbreviation | `d *DhanClient`, `s *SignalEngine` |
| Constants | PascalCase (Go style) | `MaxRetryAttempts`, `DefaultDTE` |

### Packages
- Lowercase, one word, no underscores (`datafeed`, `signals`, not `data_feed`)
- No stuttering: `signals.NewEngine()` not `signals.NewSignalsEngine()`
- Packages should have a single responsibility

## Documentation Comments

- Every exported symbol must have a doc comment
- Doc comments should be complete sentences ending in period
- Commentary should state **what** and **why**, not **how**
- Example of good doc comment:
  ```go
  // CalcIVRank returns the current VIX percentile within a 52-week window.
  // Values below 35 indicate cheap options suitable for buying.
  func CalcIVRank(currentVIX float64, low52, high52 float64) float64 {
  ```

## Type Design

### Struct Tags
```go
type OrderRequest struct {
    ClientID    string  `json:"dhanClientId"`
    OrderType   string  `json:"orderType" validate:"required,oneof=LIMIT MARKET"`
    Quantity    int     `json:"quantity" validate:"min=75"`
    Price       float64 `json:"price,omitempty"`
}
```

### Constructor Pattern
```go
// NewDhanClient creates a client with the given credentials.
// Returns error if clientID or accessToken are empty.
func NewDhanClient(clientID, accessToken string) (*DhanClient, error) {
    if clientID == "" || accessToken == "" {
        return nil, errors.New("clientID and accessToken are required")
    }
    return &DhanClient{
        ClientID:    clientID,
        AccessToken: accessToken,
        HttpClient:  &http.Client{Timeout: 5 * time.Second},
        BaseURL:     "https://api.dhan.co",
    }, nil
}
```

## Error Handling

- Return errors, don't swallow them
- Wrap with `fmt.Errorf("context: %w", err)` for propagation
- Define sentinel errors at package level:
  ```go
  var (
      ErrNotTradable   = errors.New("not tradable")
      ErrIVTooHigh     = errors.New("IV exceeds maximum threshold")
      ErrNoConviction  = errors.New("directional conviction below minimum")
  )
  ```
- Use `errors.Is()` for sentinel checks, `errors.As()` for type checks

## Concurrency

### Goroutine Lifecycle
```go
func RunDataFetcher(ctx context.Context, out chan<- MarketSnapshot) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            snap, err := fetchMarketSnapshot()
            if err != nil {
                log.Warn("snapshot fetch failed", err)
                continue
            }
            select {
            case out <- snap:
            case <-ctx.Done():
                return
            }
        case <-ctx.Done():
            return
        }
    }
}
```

### Channel Naming
- `snapshots chan<- MarketSnapshot` — send-only in parameter
- `signals <-chan Signal` — receive-only in parameter
- Full channel: `decisions chan TradeDecision`

## Testing Standards

### Table-Driven Tests
```go
func TestCalcIVRank(t *testing.T) {
    tests := []struct {
        name      string
        current   float64
        low52     float64
        high52    float64
        expected  float64
    }{
        {name: "at low", current: 12, low52: 12, high52: 30, expected: 0},
        {name: "midpoint", current: 21, low52: 12, high52: 30, expected: 50},
        {name: "at high", current: 30, low52: 12, high52: 30, expected: 100},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := CalcIVRank(tt.current, tt.low52, tt.high52)
            if got != tt.expected {
                t.Errorf("CalcIVRank() = %v, want %v", got, tt.expected)
            }
        })
    }
}
```

### Mocking External APIs
- Define interfaces for external dependencies
- Implement mock structs in `package_test` or `internal/testutil`
- Never hit real APIs in unit tests

```go
type OptionChainFetcher interface {
    Fetch() (*MarketSnapshot, error)
}

type mockFetcher struct {
    snap *MarketSnapshot
    err  error
}

func (m *mockFetcher) Fetch() (*MarketSnapshot, error) {
    return m.snap, m.err
}
```

### Race Detection
All concurrent code must be tested with `go test -race`. Any data race is a bug.

### Coverage Requirements
| Package | Minimum |
|---------|---------|
| `internal/signals/` | 80% |
| `internal/strategy/` | 80% |
| `internal/risk/` | 80% |
| `pkg/bsm/` | 90% |
| `internal/datafeed/` | 60% |
| `internal/execution/` | 60% |
| `internal/store/` | 50% |

## Performance

- No allocations in hot paths (fetch loops, signal calculation)
- Pre-allocate slices: `chain := make([]OptionData, 0, 200)`
- Use `float64` for all financial calculations (never `float32`)
- Use `int64` for quantities and OI
- Avoid `fmt.Sprintf` in loops — use `strconv` or direct byte writes

## Config Structure

```go
type Config struct {
    Dhan     DhanConfig     `yaml:"dhan"`
    NSE      NSEConfig      `yaml:"nse"`
    Trading  TradingConfig  `yaml:"trading"`
    Risk     RiskConfig     `yaml:"risk"`
    Notify   NotifyConfig   `yaml:"notify"`
}
```
