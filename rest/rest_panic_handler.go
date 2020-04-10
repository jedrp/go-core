package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime/debug"

	"github.com/jedrp/go-core/log"
	uuid "github.com/satori/go.uuid"
)

func HandlePanicMiddleware(handler http.Handler, logger log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewRequestContext(r)
		defer func() {
			if rErr := recover(); rErr != nil {
				if logger != nil {
					log.CreateRequestLogEntryFromContext(ctx, logger).Error(rErr, string(debug.Stack()))
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(500)
				response, _ := json.Marshal(map[string]string{"message": "Internal server error"})
				w.Write(response)
			}
		}()
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewRequestContext(r *http.Request) context.Context {
	ctx := r.Context()
	reqID := r.Header.Get(log.RequestIDHeaderKey)
	if reqID != "" {
		ctx = context.WithValue(ctx, log.RequestID, reqID)
	} else {
		ctx = context.WithValue(ctx, log.RequestID, uuid.NewV4().String())
	}

	corID := r.Header.Get(log.CorrelationIDHeaderKey)
	if corID != "" {
		ctx = context.WithValue(ctx, log.CorrelationID, corID)
	}
	return ctx
}
