package grpc

import (
	"context"

	strfmt "github.com/go-openapi/strfmt"
	"github.com/jedrp/go-core/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type validator interface {
	Validate(formats strfmt.Registry) error
}

type receiverWrapper struct {
	grpc.ServerStream
	formats strfmt.Registry
	ctx     context.Context
	logger  log.Logger
}

// UnaryValidatorServerInterceptor returns a new unary server interceptor that validates incoming messages.
func UnaryValidatorServerInterceptor(formats strfmt.Registry, logger log.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if v, ok := req.(validator); ok {
			if err := v.Validate(formats); err != nil {
				log.CreateRequestLogEntryFromContext(ctx, logger).Errorf("InvalidArgument %s", err.Error())
				return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
			}
		}
		return handler(ctx, req)
	}
}

// StreamValidatorServerInterceptor returns a new streaming server interceptor that validates incoming messages.
func StreamValidatorServerInterceptor(formats strfmt.Registry, logger log.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapper := &receiverWrapper{stream, formats, stream.Context(), logger}
		return handler(srv, wrapper)
	}
}

func (s *receiverWrapper) RecvMsg(m interface{}) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	if v, ok := m.(validator); ok {
		if err := v.Validate(s.formats); err != nil {
			log.CreateRequestLogEntryFromContext(s.ctx, s.logger).Errorf("InvalidArgument %s", err.Error())
			return grpc.Errorf(codes.InvalidArgument, err.Error())
		}
	}
	return nil
}
