package log

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	IsLevelEnabled(level logrus.Level) bool
	Trace(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
	WithFields(map[string]interface{}) LogEntry

	TraceWithContext(ctx context.Context, args ...interface{})
	DebugWithContext(ctx context.Context, args ...interface{})
	InfoWithContext(ctx context.Context, args ...interface{})
	WarnWithContext(ctx context.Context, args ...interface{})
	ErrorWithContext(ctx context.Context, args ...interface{})
	FatalWithContext(ctx context.Context, args ...interface{})
	PanicWithContext(ctx context.Context, args ...interface{})
	TracefWithContext(ctx context.Context, format string, args ...interface{})
	DebugfWithContext(ctx context.Context, format string, args ...interface{})
	InfofWithContext(ctx context.Context, format string, args ...interface{})
	WarnfWithContext(ctx context.Context, format string, args ...interface{})
	ErrorfWithContext(ctx context.Context, format string, args ...interface{})
	FatalfWithContext(ctx context.Context, format string, args ...interface{})
	PanicfWithContext(ctx context.Context, format string, args ...interface{})
}

type LogEntry interface {
	Trace(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
}
