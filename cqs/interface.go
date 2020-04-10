package cqs

import (
	"context"

	"github.com/jedrp/go-core/result"
)

// Executor interface, runable for Dispatcher
type Executor interface {
	Execute(context.Context) *result.Result
	SetDependences(context.Context, interface{})
}

// Command interface, change state action
type Command interface {
	Executor
}

// Query interface, return result, don't change state
type Query interface {
	Executor
}

// Dispatcher execute command or query, log when command or query return fail status
type Dispatcher interface {
	Dispatch(ctx context.Context, e Executor) *result.Result
	Register(ctx context.Context, deps interface{}, v ...Executor)
}
