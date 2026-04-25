package api

import (
	"context"
	"errors"
)

var (
	ErrNoAPI      = errors.New("no API in context")
	ErrInvalidAPI = errors.New("invalid API in context")
)

type contextKey string

const apiKey contextKey = "api"

func (a *api) AddTo(ctx context.Context) context.Context {
	return context.WithValue(ctx, apiKey, a)
}

func From(ctx context.Context) (*api, error) {
	a := ctx.Value(apiKey)
	if a == nil {
		return nil, ErrNoAPI
	}
	if api, ok := a.(*api); ok {
		return api, nil
	}
	return nil, ErrInvalidAPI
}
