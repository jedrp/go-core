package cqs

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/jedrp/go-core/log"
)

type memoryDispatcher struct {
	maxLatencyInMillisecond time.Duration
	handlersMap             map[string]interface{}
}

var (
	ErrHandlerNotFound       = fmt.Errorf("handler not found error")
	ErrHandlerTypeNotSupport = fmt.Errorf("handler type not supported")
	defaultDispatcher        = &memoryDispatcher{
		maxLatencyInMillisecond: 0,
		handlersMap:             make(map[string]interface{}),
	}
)

func ConfigureTimeOut(timeoutInMillisecond int) {
	defaultDispatcher.maxLatencyInMillisecond = time.Duration(timeoutInMillisecond) * time.Millisecond
}

func RegisterRequestHandlerFactory[TRequest Request, TResponse Response](ctx context.Context, factory HandlerFactory[TRequest, TResponse]) error {
	return registerRequestHandler[TRequest, TResponse](factory)
}

func RegisterHandler[TRequest Request, TResponse Response](ctx context.Context, handler Handler[TRequest, TResponse]) error {
	return registerRequestHandler[TRequest, TResponse](handler)
}

func registerRequestHandler[TRequest Request, TResponse Response](handler any) error {
	r := *new(TRequest)
	_, exist := defaultDispatcher.handlersMap[r.HandlerID()]
	if exist {
		typeName := reflect.TypeOf(r).String()
		// each request in request/response strategy should have just one handler
		return fmt.Errorf("duplicated executer registration detected of type: %s handlerID: %s", typeName, r.HandlerID())
	}

	defaultDispatcher.handlersMap[r.HandlerID()] = handler

	return nil
}

func Send[TRequest Request, TResponse Response](ctx context.Context, request TRequest) (TResponse, error) {
	var (
		cancel context.CancelFunc
	)
	if defaultDispatcher.maxLatencyInMillisecond > 0 {
		ctx, cancel = context.WithTimeout(ctx, defaultDispatcher.maxLatencyInMillisecond*time.Millisecond)
		defer cancel()
	}
	handlerID := request.HandlerID()
	if log.DefaultLogger.IsLevelEnabled(logrus.DebugLevel) {
		defer elapsed(ctx, "dispatching "+request.HandlerID(), log.DefaultLogger)()
	}
	if hv, ok := defaultDispatcher.handlersMap[handlerID]; ok {
		h, err := buildHandler[TRequest, TResponse](hv)
		if err != nil {
			return *new(TResponse), err
		}
		response, err := h.Handle(ctx, request)
		if err != nil {
			log.CreateRequestLogEntryFromContext(ctx, log.DefaultLogger).Error(err)
		}
		return response, err
	}

	msg := fmt.Sprintf("MemoryDispatcher can't find handler for type: %s handlerID: %s", reflect.TypeOf(request).String(), handlerID)
	log.CreateRequestLogEntryFromContext(ctx, log.DefaultLogger).Error(msg)
	return *new(TResponse), ErrHandlerNotFound
}

func buildHandler[TRequest Request, TResponse Response](handler any) (Handler[TRequest, TResponse], error) {
	handlerValue, ok := handler.(Handler[TRequest, TResponse])
	if !ok {
		factory, ok := handler.(HandlerFactory[TRequest, TResponse])
		if !ok {
			return nil, ErrHandlerTypeNotSupport
		}

		return factory(), nil
	}

	return handlerValue, nil
}

func elapsed(ctx context.Context, what string, logger log.Logger) func() {
	start := time.Now()
	return func() {
		logger.DebugfWithContext(ctx, "%s took %v", what, time.Since(start))
	}
}

func ResetDispatcherSetting() {
	defaultDispatcher.handlersMap = make(map[string]interface{})
}
