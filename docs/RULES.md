# Development Rules & Conventions

## General Rules

### Code Quality
- All code must compile without errors (`go build ./...`)
- All code must pass `gofmt` formatting
- All code must pass `go vet` with zero warnings
- No unused imports, variables, or functions
- No exported symbols without doc comments
- No `panic()` in production code — use error return values
- No `log.Fatal()` outside of `main.go`

### Testing
- Every package must have tests (`*_test.go`)
- Minimum coverage: 70% for packages under `internal/`, 50% for `pkg/`
- Tests must not depend on external APIs (NSE, Dhan) — mock all external calls
- Table-driven tests using Go's `testing` package (no third-party test frameworks)
- Race condition tests for all goroutine-based code: `go test -race`
- Backtest tests must reproduce from stored data, not fetch live

### Git Conventions
- **Branch naming**: `feat/<description>`, `fix/<description>`, `chore/<description>`
- **Commit style**: Imperative mood, present tense. e.g. "add IV rank calculator" not "added IV rank"
- Keep commits atomic — one logical change per commit
- No commits to `main` directly — always use PRs

### Dependencies
- Minimize third-party dependencies — prefer stdlib
- No dependency with a license incompatible with commercial use
- Pin exact versions in `go.mod` — no `latest` tags
- All new dependencies must be approved

## Go-Specific Rules

### Error Handling
- Always check errors. Never use `_` to discard an error
- Wrap errors with context: `fmt.Errorf("fetch option chain: %w", err)`
- Use `errors.Is()` and `errors.As()` for error inspection — never string matching
- Define sentinel errors for expected failure modes: `var ErrNotTradable = errors.New("not tradable")`
- Log errors at the boundary (goroutine entry/exit), not at every intermediate step

### Goroutines & Channels
- All goroutines must accept `context.Context` and respect cancellation
- All goroutines must be started with a deferred cleanup or `defer` in the launch function
- Channel ownership: the goroutine that creates a channel is responsible for closing it
- Buffered channels preferred — size must be documented in a comment
- No global channel variables — pass channels as function parameters
- Use `sync.WaitGroup` or `errgroup.Group` to track goroutine completion during shutdown
- **Never** use `sync.Mutex` — use channels for communication instead

### Concurrency Patterns
- Pipeline pattern: goroutines connected by channels (datafetcher → signal engine → strategy → execution)
- Fan-out for independent parallel work (e.g., multiple signal calculations concurrently)
- Select statement for non-blocking multi-channel waits
- Always include `case <-ctx.Done()` in select statements for clean shutdown

### Struct Design
- Zero-value structs must be usable or have a `New*` constructor
- Use value receivers for structs that are small and immutable
- Use pointer receivers for structs with mutable state or large size
- Validate struct fields at construction time, not at method call time

### Configuration
- Never hardcode API keys, endpoints, or thresholds
- Use `config/config.yaml` for all tunable parameters
- Use environment variables for secrets (access tokens, client IDs)
- Config fields must have YAML tags and sensible defaults

## Code Review Checklist

Before merging any code, verify:
- [ ] Compiles and passes `go vet`
- [ ] All tests pass including `-race`
- [ ] No leaked goroutines — shutdown is testable
- [ ] All external API calls are mocked in tests
- [ ] Error paths are tested (not just happy path)
- [ ] No magic numbers — all constants have names
- [ ] Channels are properly owned and closed
- [ ] Context cancellation is respected
- [ ] Config changes are backward-compatible
- [ ] Logs contain enough info to debug in production

## Workflow Rules

### Before Writing Code
1. Read the relevant existing files to understand patterns
2. Check if the task is in the build roadmap
3. Write the test file first (TDD where practical)

### After Writing Code
1. Run `go build ./...`
2. Run `go vet ./...`
3. Run `go test ./... -race`
4. Run `gofmt -w .`
5. Commit with a descriptive message

### AI Agent Rules
- Never generate or guess API keys, secrets, or tokens
- Never use `os/exec` or shell commands to work around Go code
- Never add comments that explain what the code does — code should be self-documenting
- Never modify `go.mod` without running `go mod tidy`
