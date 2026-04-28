# ba — Business Action

A small, generic Go library for structuring business operations as **actions**: a unit of work with built-in authorization, validation, transaction handling, and after-commit callbacks. Inspired by Toptal's [`granite`](https://github.com/toptal/granite) Ruby gem and ported to idiomatic Go with generics. Originally extracted from [`go-web-app-template`](https://github.com/bioform/go-web-app-template).

## Design tenets

1. **Each BA runs in its own transaction.** The `TransactionProvider` is invoked per-action; lifecycle is bound to the action (commit triggers `AfterCommit` callbacks; abort skips them).
2. **Nested BAs use nested transactions.** A BA calling `ba.New(ctx, &Other{}).Perform()` from inside its own `Perform` produces a nested transaction — typically a savepoint, depending on what your `TransactionProvider` supports. The library assumes the convention that nested transactions are always real savepoints (Granite enforces this with `requires_new: true`); GORM's `db.Transaction(fn)` already behaves this way. If you implement `TransactionProvider` against `database/sql` directly, issue a `SAVEPOINT` for nested calls to preserve the contract.
3. **ORM-independent.** The library defines a `TransactionProvider` interface; the `examples/` happen to use GORM, but the core has no DB dependency. Plug in `database/sql`, ent, sqlx, or any custom store.
4. **First-class `Performer` concept.** Every action has a Performer (typically a user; `ba.SystemPerformer` for system calls). `IsAllowed()` decides whether this performer may run the action; `As(performer)` and `AsSystem()` set it explicitly. Authorization is not bolted on — it's a primitive.
5. **Built-in tracing simplifies testing.** `ba.CallTracker` is a package-level hook every `ba.New(...)` registers with; combined with `matcher.CallAction`, tests can assert "this BA called BA-X with these fields, as system, via Try" without mocking the full call graph.
6. **Optional attributes via `attr.Type[T]`.** Distinguishes "unset" from "zero-value" (e.g., `int(0)` vs. unset) — useful for conditional defaults in `Init()` and for "set and non-empty" validation via `attr.Required[T]` (plays nicely with [govalidator](https://github.com/rezakhademix/govalidator)). Using `attr` is itself optional — the core library has no dependency on it, and action fields can be plain Go types; the examples adopt it as a convention.

## Install

```sh
go get github.com/bioform/ba
```

Requires Go 1.26+. The core library has no required runtime dependencies — you bring your own ORM (or none) by implementing `TransactionProvider`. The examples pull in [GORM](https://gorm.io/), [SQLite](https://gorm.io/docs/connecting_to_the_database.html#SQLite), and [govalidator](https://github.com/rezakhademix/govalidator) — none of which are required to consume `ba` itself.

The library's own test suite uses [Ginkgo v2](https://github.com/onsi/ginkgo), [Gomega](https://github.com/onsi/gomega), and [testify/mock](https://github.com/stretchr/testify) — these are picked up only when running `go test`, not when importing `ba`.

## Execution flow

`ba.New(ctx, action).Perform()` runs this sequence:

1. `IsAllowed()` — authorization (cached)
2. `IsEnabled()` — feature flag *or* subject-state precondition (cached). Granite's `precondition` blocks live here: return `(false, ba.ErrorMap{...})` for any state-dependent reason the action shouldn't run (subject already in target state, system in maintenance, quota exceeded). The framework surfaces it as a `DisabledError`.
3. `IsValid()` — validation
4. `Perform()` — your business logic, wrapped in a transaction from `TransactionProvider`
5. `AfterCommit` callbacks — executed only if the transaction commits

`Try()` is the same as `Perform()` but returns `(false, nil)` when the action is disabled, instead of an error.

## Lifecycle hooks

Beyond the four core checks (`IsAllowed`, `IsEnabled`, `IsValid`, `Perform`), `BaseAction` provides two additional hooks you can override:

- **`Init()`** — called once by `ba.New(...)`, before any of the lifecycle checks. Override to set default attribute values:

  ```go
  func (a *CreateUser) Init() {
      if !a.Role.IsSet() {
          a.Role = attr.Value("member")
      }
  }
  ```

- **`ErrorHandler(err error) error`** — called when `Perform` (or any lifecycle method) returns an error, before the framework wraps it in `ActionError`. Override to translate domain errors into typed `ba` errors, swallow benign failures, or attach context. Returning the same `err` (the default) leaves it untouched:

  ```go
  func (a *CreateUser) ErrorHandler(err error) error {
      if errors.Is(err, gorm.ErrDuplicatedKey) {
          return &EmailDuplicateError{Email: a.Email}
      }
      return err
  }
  ```

## Minimal usage

```go
type CreateUser struct {
    api.BaseAction       // your app-specific BaseAction (provides DB / TransactionProvider)
    Name  string
    Email string
}

func (a *CreateUser) Perform() error {
    return a.DB().Create(&model.User{Name: a.Name, Email: a.Email}).Error
}

func (a *CreateUser) IsValid() (bool, error) {
    v := validator.New()
    v.RequiredString(a.Email, "Email", "required")
    return v.IsPassed(), ba.ErrorMap(v.Errors())
}

// Run it:
ok, err := ba.New(ctx, &CreateUser{Name: "Ada", Email: "ada@example.com"}).Perform()
```

## Examples

The [`examples/`](./examples) directory contains four focused, runnable examples — each with its own `README.md` explaining the lesson:

| Directory                                                                 | Lesson                                                                            |
| ------------------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| [`01_basic`](./examples/01_basic)                                         | The four core lifecycle methods composed into a single happy path                 |
| [`02_lifecycle_hooks`](./examples/02_lifecycle_hooks)                     | `Init`, `IsEnabled` precondition, `ErrorHandler`, after-commit failure semantics  |
| [`03_subject_pattern`](./examples/03_subject_pattern)                     | Subject-pattern action with `Performer`-aware `IsAllowed`                         |
| [`04_nested_transactions`](./examples/04_nested_transactions)             | Nested actions → nested transactions / savepoints, plus a matcher integration test |

Run any example with `go run`:

```sh
go run ./examples/01_basic
go run ./examples/02_lifecycle_hooks
go run ./examples/03_subject_pattern
go run ./examples/04_nested_transactions
```

Shared infrastructure (`api`, `database`, `model`) lives in [`examples/pkg/`](./examples/pkg) — a GORM-backed `TransactionProvider`, an app-level `BaseAction`, and an in-memory SQLite seeded via `AutoMigrate`.

## Conventions

### Subject pattern

When an action mutates a single primary entity, hold it as a struct field — the **subject** — distinct from input parameters:

```go
type UpdateUserName struct {
    api.BaseAction

    User    *model.User       // subject — what the action operates on
    NewName attr.Type[string] // input
}
```

Compare to creator-style actions (`CreateUser` in [`01_basic`](./examples/01_basic)) where there is no pre-existing subject — the action *produces* the entity from input. The distinction makes authorization read naturally — `IsAllowed` asks "may this `Performer()` modify `User`?" — and lets matcher assertions express intent clearly:

```go
matcher.CallAction(&UpdateUserName{User: targetUser}).
    With(Fields{"NewName": Equal(attr.Value("Ada Lovelace"))})
```

The convention isn't enforced by the library — it's a Granite carry-over worth adopting. See [`examples/03_subject_pattern`](./examples/03_subject_pattern) for a runnable demo with a `Performer`-aware `IsAllowed`.

## Errors

All errors wrap `ba.ActionError`. Specific subtypes:

| Type                   | Source                  |
| ---------------------- | ----------------------- |
| `AuthorizationError`   | `IsAllowed` returns `false` |
| `DisabledError`        | `IsEnabled` returns `false` |
| `ValidationError`      | `IsValid` returns `false`   |
| `CallbackError`        | An after-commit callback fails |

## Testing

The library ships with first-class testing support:

- **`ba.CallTracker`** — package-level hook; every `ba.New(...)` registers with it. Set in tests to capture or stub action calls.
- **`matcher.CallAction`** — Gomega matcher for asserting that a specific action was invoked with specific fields and performer:

```go
Expect(func() {
    ba.New(ctx, &ParentAction{}).Perform()
}).To(
    matcher.CallAction(&ChildAction{}).
        AsSystem().
        With(Fields{"SomeAttr": Equal(expected)}).
        AndCallOriginal(),
)
```

- **`mocks/`** — testify/mock-generated mocks for `Action`, `TransactionProvider`, etc.
- **`dummy/`** — a `DummyAction` for unit tests.

A complete integration test is in [`examples/04_nested_transactions/main_test.go`](./examples/04_nested_transactions/main_test.go).

## Running locally

Clone the repo and use any Go 1.26+. The project pins its Go version via [mise](https://mise.jdx.dev/) (`mise.toml`); if you have mise installed, `mise install` provisions the matching toolchain. Otherwise install Go 1.26+ yourself.

```sh
git clone https://github.com/bioform/ba && cd ba
mise install              # optional: provisions Go 1.26.2 per mise.toml
go test -race ./...       # library + integration test
go run ./examples/01_basic
```

CI runs `go vet`, `go mod tidy` cleanliness, build, and `go test -race -cover ./...` on every push and pull request — see [`.github/workflows/ci.yml`](./.github/workflows/ci.yml).

## License

MIT. See [LICENSE](./LICENSE).
