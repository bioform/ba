# 02 — Lifecycle hooks

Three behaviors that aren't the four core checks:

| Hook                 | Demonstrated by run                                          |
| -------------------- | ------------------------------------------------------------ |
| `Init`               | Run 1 — fills a default `Name` when omitted                  |
| `ErrorHandler`       | Run 2 — translates GORM's `ErrDuplicatedKey` into a domain error |
| `IsEnabled` (precondition) | Run 3 — rejects via `ba.ErrorMap` → `DisabledError`     |
| `AfterCommit` (failure) | Run 1 — callback errors but the user row survives commit  |

The third run flips a global `SignupsEnabled` flag — flip it back to demonstrate the happy path again.

`IsEnabled` plays the role Granite calls "preconditions": state-dependent reasons an action shouldn't run, distinct from authorization (who) and validation (input shape).

Run it:

```sh
mise exec -- go run ./examples/02_lifecycle_hooks
```
