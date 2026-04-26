# 03 — Subject pattern

A naming convention carried over from Granite: when an action mutates a single primary entity, hold it as a struct field — the **subject** — distinct from input parameters.

```go
type UpdateUserName struct {
    api.BaseAction

    User    *model.User       // subject — entity being mutated
    NewName attr.Type[string] // input
}
```

Compare to *creator-style* actions like `CreateUser` (see [`01_basic`](../01_basic)) where there is no pre-existing subject — the action *produces* the entity from input.

The convention's payoff is in `IsAllowed`, which reads naturally as "may this `Performer()` modify `User`?". This example uses a tiny `Admin{ID}` struct as the performer and rejects everything else:

```go
func (a *UpdateUserName) IsAllowed() (bool, error) {
    switch a.Performer().(type) {
    case Admin: return true, nil
    default:    return false, nil
    }
}
```

The library does not enforce the pattern. It's a convention worth adopting because it makes authorization and tests express intent.

Run it:

```sh
mise exec -- go run ./examples/03_subject_pattern
```
