# 04 — Nested transactions + matcher test

Two lessons in one example:

## Nested actions produce nested transactions

`ActionA.Perform` invokes `ba.New(ctx, &ActionB{}).AsSystem().Perform()`. Because each `Perform` runs through the `TransactionProvider`, the inner call becomes a savepoint inside the outer transaction.

The example arranges for `ActionB` to deliberately fail:

1. `ActionA` inserts an `outer` user
2. `ActionA` invokes `ActionB` — `ActionB` inserts an `inner` user, then returns an error
3. The framework rolls back the inner savepoint — the `inner` row is gone
4. `ActionA` recovers (logs the inner failure, returns `nil`)
5. The outer transaction commits — only the `outer` user survives

Run it and check the final user list:

```sh
mise exec -- go run ./examples/04_nested_transactions
```

## Matcher integration test

[`main_test.go`](./main_test.go) demonstrates `matcher.CallAction` — a Gomega matcher that asserts a specific BA was invoked with specific fields, as a specific performer:

```go
Expect(func() {
    ba.New(ctx, &main.ActionA{...}).Perform()
}).To(CallAction(&main.ActionB{}).
    AsSystem().
    With(Fields{"AttrB": Equal(attr.Value(123)), "AttrB2": Equal("some string")}).
    ViaPerform())
```

Without `.AndCallOriginal()`, the matcher *stubs* `ActionB`'s body — useful when the inner action would fail (as it does in the savepoint demo). Adding `.AndCallOriginal()` runs the real body and additionally asserts it succeeded.

Run the test:

```sh
mise exec -- go test ./examples/04_nested_transactions
```
