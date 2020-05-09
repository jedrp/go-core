package result

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetRPCError(err *Error) error {
	if err == nil {
		return nil
	}
	var grpcCode codes.Code
	switch err.Code {
	case Aborted:
		grpcCode = codes.Aborted
	case ResourceExhausted:
		grpcCode = codes.ResourceExhausted
	case AlreadyExists:
		grpcCode = codes.AlreadyExists
	case Canceled:
		grpcCode = codes.Canceled
	case DataLoss:
		grpcCode = codes.DataLoss
	case DeadlineExceeded:
		grpcCode = codes.DeadlineExceeded
	case FailedPrecondition:
		grpcCode = codes.FailedPrecondition
	case Internal:
		grpcCode = codes.Internal
	case InvalidArgument:
		grpcCode = codes.InvalidArgument
	case NotFound:
		grpcCode = codes.NotFound
	case OutOfRange:
		grpcCode = codes.OutOfRange
	case PermissionDenied:
		grpcCode = codes.PermissionDenied
	case Unauthenticated:
		grpcCode = codes.Unauthenticated
	case Unavailable:
		grpcCode = codes.Unavailable
	case Unimplemented:
		grpcCode = codes.Unimplemented
	default:
		grpcCode = codes.Unknown
	}

	return status.Errorf(grpcCode, err.Message)
}
