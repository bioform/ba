package ctxstore

import (
	"context"
	"log/slog"

	"github.com/bioform/ba/examples/pkg/model"
)

// ContextKey is used for context.Context value. The value requires a key that is not primitive type.
type contextKey string // can be unexported

type userKey string // can be unexported
// UserKey is the ContextKey for User
const UserKey userKey = "user"

func AssignUser(ctx context.Context, user *model.User) context.Context {

	return context.WithValue(ctx, UserKey, user)
}

func GetUser(ctx context.Context) *model.User {
	if ctx == nil {
		slog.Error("get user from context", "error", "context is nil")
		return nil
	}
	u := ctx.Value(UserKey)
	if u == nil {
		return nil
	}

	user, ok := u.(*model.User)
	if !ok {
		slog.Error("get user from context", "user", u)
		return nil
	}

	return user
}
