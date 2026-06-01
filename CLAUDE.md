# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`github.com/bioform/ba` is a Go library implementing a structured **Business Action** pattern (BA = "business action"). It provides a generic, type-safe framework for executing business operations with built-in transaction management, authorization, validation, feature-flag/precondition checks, after-commit callbacks, and first-class testing support.

Inspired by Toptal's [`granite`](https://github.com/toptal/granite) Ruby gem.

## Commands

The project uses [mise](https://mise.jdx.dev/) to pin the Go version (`mise.toml`). Prefer `mise exec -- go ...` over the system `go` so the pinned version is honored.

```bash
# Run all tests
mise exec -- go test ./...

# Run with race detector
mise exec -- go test -race ./...

# Run with coverage
mise exec -- go test -cover ./...

# Run a single Ginkgo suite (the library uses package ba_test for its own suite)
mise exec -- go test -run "TestAction" .

# Run an example end-to-end
mise exec -- go run ./examples/01_basic
mise exec -- go run ./examples/02_lifecycle_hooks
mise exec -- go run ./examples/03_subject_pattern
mise exec -- go run ./examples/04_nested_transactions
```

CI runs build + race tests + `go mod tidy` cleanliness check on push/PR (see `.github/workflows/ci.yml`).

## Architecture

### Core execution flow

`ba.New[A](ctx, action).Perform()` runs this sequence:

1. **IsAllowed()** — authorization check (cached)
2. **IsEnabled()** — feature flag *or* state-dependent precondition (cached). Granite's "precondition" blocks live here — return `(false, ba.ErrorMap{...})` for any state-based reason the action shouldn't run (e.g. subject already in target state, system in maintenance, quota exceeded). The framework surfaces it as a `DisabledError`. See `examples/02_lifecycle_hooks` for a precondition use.
3. **IsValid()** — validation check
4. **Perform()** — business logic, wrapped in a transaction from `TransactionProvider`
5. **AfterCommit callbacks** — run after the transaction commits

`Try()` is like `Perform()` but returns `(false, nil)` when the action is disabled (via `NopIfDisabled` option) instead of an error.

A BA calling `ba.New(ctx, &Other{}).Perform()` from inside its own `Perform` produces a **nested transaction** — typically a savepoint, depending on what your `TransactionProvider` supports. `examples/04_nested_transactions` demonstrates this with a deliberately-failing inner action whose savepoint rolls back without aborting the outer.

### Key types

**`ba.Action` interface** (`action.go`) — contract every action must satisfy. Users typically only implement `Perform()` and override `IsAllowed`, `IsEnabled`, or `IsValid` as needed.

**`ba.BaseAction`** (`base_action.go`) — embed this in every action struct to get default no-op implementations of all interface methods except `Perform()`.

**`ba.ActionPerformerImpl[A]`** (`action_performer.go`) — generic execution engine. Constructed with `ba.New[A](ctx, action)`. Exposes `Perform()`, `Try()`, `IsPerformable()`, `As(performer)`, `AsSystem()`.

**`ba.TransactionProvider`** (`transaction_provider.go`) — interface with a single `Transaction(ctx, fn)` method. The action's `TransactionProvider()` method returns this; the examples wrap a GORM DB. The core library has no DB dependency — keep it that way.

**`ba.ErrorMap`** (`map[string]string`) — field-level error map returned by `IsEnabled()` and `IsValid()`. The performer wraps it into `DisabledError` or `ValidationError`.

**`attr.Type[T]`** (`attr/value.go`) — generic attribute wrapper that distinguishes unset from zero-value. Use `attr.Required[T](v)` as a validator with govalidator.

### Error hierarchy

All errors wrap `ActionError` (carries the action and performer for context). Specific subtypes: `ValidationError` (from `IsValid`), `DisabledError` (from `IsEnabled`), `AuthorizationError` (from `IsAllowed`), `CallbackError` (from after-commit callbacks).

### Testing support

**`ba.CallTracker`** — package-level variable. When non-nil, every `ba.New()` call registers with it. Set this in tests to capture executions.

**`matcher.CallActionMatcher`** (`matcher/`) — Gomega matcher for asserting that a specific action was called, with specific fields and performer. See `examples/04_nested_transactions/main_test.go` for usage:

```go
Expect(func() {
    ba.New(ctx, &ActionA{}).Perform()
}).To(
    matcher.CallAction(&ActionB{}).
        AsSystem().
        With(Fields{"SomeAttr": Equal(expectedVal)}).
        AndCallOriginal(), // omit to stub the inner action's body
)
```

`AndCallOriginal()` invokes the real inner action and asserts it succeeded (rejects if it returned an error). Omit it to stub the body — useful when the inner action intentionally fails.

**`tracker.TestTracker`** (`matcher/tracker/`) — implementation of `CallTracker` that wraps an inner tracker and can stub or spy on specific action types.

### Application layer pattern (from examples)

The `examples/pkg/api/` directory shows the recommended wiring:

- `examples/pkg/api/api.go` — implements `TransactionProvider` by wrapping a GORM `*gorm.DB`. When `Transaction(ctx, fn)` is invoked, it creates a new `api` instance bound to the GORM tx and re-publishes it on the lambda's context — that's how nested BAs see the transaction-scoped DB.
- `examples/pkg/api/context.go` — context plumbing: `(*api).AddTo(ctx)` stores the instance under the package-private `apiKey`; `api.From(ctx)` retrieves it (returning `ErrNoAPI` / `ErrInvalidAPI`).
- `examples/pkg/api/base_action.go` — application-specific `BaseAction` that embeds `ba.BaseAction` and adds `API()` and `DB()` helpers that pull from context via `api.From(ctx)`.
- Actions embed the app-level `BaseAction`, not the library `BaseAction`, to get access to DB/API in `Perform()`.

Context setup:
```go
ctx := api.New(database.Default()).AddTo(context.Background())
ap := ba.New(ctx, &MyAction{Attr: value})
ok, err := ap.Perform()
```

### Package layout

```
.                       # Core library — package ba
├── action.go           # Action interface, Performer, SystemPerformer
├── action_performer.go # Generic execution engine: ActionPerformerImpl[A]
├── base_action.go      # Default no-op implementation users embed
├── after_commit.go     # AfterCommitCallback wiring
├── cache.go            # Authorization/feature-flag result cache
├── errors.go           # ActionError, ValidationError, DisabledError, etc.
├── options.go          # Perform options (NopIfDisabled, SkipCache, ...)
├── tracker.go          # CallTracker hook
├── transaction_provider.go
│
├── attr/               # Generic attribute wrapper (attr.Type[T])
├── dummy/              # DummyAction for unit tests
├── matcher/            # Gomega matchers for action assertions
│   ├── option/         # Matcher options (Perform vs Try, CallOriginal)
│   └── tracker/        # TestTracker implementation
├── mocks/              # testify/mock generated mocks
│
└── examples/           # Reference implementations (separate from the library)
    ├── 01_basic/               # Four core lifecycle methods, happy path
    ├── 02_lifecycle_hooks/     # Init, IsEnabled (precondition), ErrorHandler, AfterCommit failure
    ├── 03_subject_pattern/     # Subject-pattern action with Performer-aware IsAllowed
    ├── 04_nested_transactions/ # Nested actions → savepoint demo + matcher test
    └── pkg/                    # Shared example infrastructure
        ├── api/                # GORM-backed TransactionProvider, app BaseAction
        ├── database/           # In-memory SQLite + AutoMigrate at init
        └── model/              # Example User model

Each example directory has its own README.md explaining the single lesson it teaches. When adding a new lesson, prefer a new numbered directory over piling onto an existing one — examples are scoped to one concept by convention.
```

## Testing framework

Tests use **Ginkgo v2** + **Gomega** (`github.com/onsi/ginkgo/v2`, `github.com/onsi/gomega`) with **testify/mock** for mocks. The library's own suite lives in `*_test.go` files at the module root using `package ba_test`. Run with standard `go test` (or `mise exec -- go test`).
