package cqs

import (
	"context"
)

// Executor interface, runable for Dispatcher
type Handler[TRequest Request, TResponse Response] interface {
	Handle(context.Context, TRequest) (TResponse, error)
}

type HandlerFactory[TRequest Request, TResponse Response] func() Handler[TRequest, TResponse]
