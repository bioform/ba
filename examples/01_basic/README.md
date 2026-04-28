# 01 — Basic action

The "hello world" of `ba`. One action, one DB write, one after-commit callback.

**Lesson:** how the four core lifecycle methods compose.

| Hook        | What this example does                                      |
| ----------- | ----------------------------------------------------------- |
| `IsAllowed` | inherited default (`true, nil`) — no authorization checks   |
| `IsEnabled` | inherited default — no preconditions                        |
| `IsValid`   | govalidator: `Name` and `Email` are required                |
| `Perform`   | inserts a `User`, registers a successful `AfterCommit` log  |

The `Perform` body runs inside a transaction provided by `api.BaseAction.TransactionProvider()`. The `AfterCommit` callback runs *only* after that transaction commits — if `Perform` returns an error, the callback never fires.

*Incidentally also showcases `attr.Required[T]` — a govalidator rule for "set and non-empty" that distinguishes an omitted field from an empty string. `attr` is a convention; not required to use `ba`.*

Run it:

```sh
mise exec -- go run ./examples/01_basic
```
