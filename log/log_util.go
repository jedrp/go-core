package log

import "context"

const (
	RequestIDHeaderKey     = "Request-Id"
	CorrelationIDHeaderKey = "Correlation-Id"
	RequestID              = "RequestId"
	CorrelationID          = "CorrelationId"
)

func CreateRequestLogEntryFromContext(ctx context.Context, log Logger) LogEntry {
	return log.WithFields(map[string]interface{}{
		CorrelationID: ctx.Value(CorrelationID),
		RequestID:     ctx.Value(RequestID),
	})
}
