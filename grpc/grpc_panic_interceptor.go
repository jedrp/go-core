package grpc

import (
	"context"
	"runtime/debug"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/jedrp/go-core/log"
	"google.golang.org/grpc"
)

//UnaryServerPanicInterceptor hanle unexpected panic
func UnaryServerPanicInterceptor(logger log.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				if logger != nil {
					log.CreateRequestLogEntryFromContext(ctx, logger).Error(r, string(debug.Stack()))
				}
			}
		}()

		return handler(ctx, req)
	}
}

func getRecoveryHandlerFuncContextHandler(logger log.Logger) grpc_recovery.RecoveryHandlerFuncContext {
	return grpc_recovery.RecoveryHandlerFuncContext(
		func(ctx context.Context, p interface{}) error {
			if logger != nil {
				log.CreateRequestLogEntryFromContext(ctx, logger).Error(p, string(debug.Stack()))
			}
			return nil
		},
	)
}
