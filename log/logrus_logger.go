package log

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/jedrp/go-core/util"
	"github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

const (
	// HookType type of hook es (elastic search)
	HookType = "type"
	Loglevel = "level"
)

// LogrusLogger logrus wrapper implementation
// LogHook# fired when log fired
// Elasticsearch hook format "type=[es];host=host_url;index-prefix=prefix;sniff=true|false;mode=sync|async"
type LogrusLogger struct {
	LogHook1 string `long:"log-hook-1" description:"the hook connection string" env:"LOG_HOOK_1" json:"hook1,omitempty"`
	LogHook2 string `long:"log-hook-2" description:"the hook connection string" env:"LOG_HOOK_2" json:"hook2,omitempty"`
	LogHook3 string `long:"log-hook-3" description:"the hook connection string" env:"LOG_HOOK_3" json:"hook3,omitempty"`
	LogHook4 string `long:"log-hook-4" description:"the hook connection string" env:"LOG_HOOK_4" json:"hook4,omitempty"`

	LogConfigStr   string `long:"log-config" description:"the hook connection string" env:"LOG_CONFIG" json:"logConfigStr,omitempty"`
	logLevel       string
	*logrus.Logger `json:"-"`
}

// New logger
func New() Logger {
	logrusLogger := &LogrusLogger{
		logLevel: "debug",
	}
	parser := flags.NewParser(logrusLogger, flags.IgnoreUnknown)
	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}
	return newWith(logrusLogger)
}

func newWith(logrusLogger *LogrusLogger) Logger {
	config, err := util.GetConfig(logrusLogger.LogConfigStr)
	if err != nil {
		log.Panic(err)
	}
	log := &logrus.Logger{
		Out:          os.Stdout,
		Formatter:    new(LoggerTextFormatter),
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.InfoLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
	logrusLogger.Logger = log
	if config == nil {
		return logrusLogger
	}

	if v, f := config[Loglevel]; f {
		logrusLogger.logLevel = v
	}

	level, err := logrus.ParseLevel(logrusLogger.logLevel)

	if err != nil {
		log.Panic(err)
	}
	log.Level = level
	addHook(log, logrusLogger.LogHook1, level)
	addHook(log, logrusLogger.LogHook2, level)
	addHook(log, logrusLogger.LogHook3, level)
	addHook(log, logrusLogger.LogHook4, level)

	logSetting, err := json.Marshal(logrusLogger)
	if err != nil {
		log.Infof("serialize log setting fail with message %s", err.Error())
	}
	log.Infof("log initialized with setting: %s", string(logSetting))
	return logrusLogger
}

func addHook(log *logrus.Logger, hookStr string, level logrus.Level) {
	if hookStr != "" {
		hookType := getHookType(hookStr)
		switch hookType {
		case "es":
			hook, err := NewElasticHookFromStr(hookStr, level)
			if err != nil {
				log.Panic(err)
			}
			log.Hooks.Add(hook)
			break
		}
	}
}
func getHookType(hookStr string) string {
	config, err := util.GetConfig(hookStr)
	if err != nil {
		panic(err)
	}
	return config[HookType]
}

// WithFields
func (logrusLogger *LogrusLogger) WithFields(fields map[string]interface{}) LogEntry {
	return logrusLogger.Logger.WithFields(fields)
}

func (logrusLogger *LogrusLogger) IsLevelEnabled(level logrus.Level) bool {
	return logrusLogger.Logger.IsLevelEnabled(level)
}

// NewEntry
func NewEntry(logger *LogrusLogger) *logrus.Entry {
	return &logrus.Entry{
		Logger: logger.Logger,
		// Default is three fields, plus one optional.  Give a little extra room.
		Data: make(logrus.Fields, 6),
	}
}

func (logrusLogger *LogrusLogger) TraceWithContext(ctx context.Context, args ...interface{}) {
	CreateRequestLogEntryFromContext(ctx, logrusLogger).Trace(args)
}
func (logrusLogger *LogrusLogger) DebugWithContext(ctx context.Context, args ...interface{}) {
	CreateRequestLogEntryFromContext(ctx, logrusLogger).Debug(args)
}
func (logrusLogger *LogrusLogger) InfoWithContext(ctx context.Context, args ...interface{}) {
	CreateRequestLogEntryFromContext(ctx, logrusLogger).Info(args)
}
func (logrusLogger *LogrusLogger) WarnWithContext(ctx context.Context, args ...interface{}) {
	CreateRequestLogEntryFromContext(ctx, logrusLogger).Warn(args)
}
func (logrusLogger *LogrusLogger) ErrorWithContext(ctx context.Context, args ...interface{}) {
	CreateRequestLogEntryFromContext(ctx, logrusLogger).Error(args)
}
func (logrusLogger *LogrusLogger) FatalWithContext(ctx context.Context, args ...interface{}) {
	CreateRequestLogEntryFromContext(ctx, logrusLogger).Error(args)
}
func (logrusLogger *LogrusLogger) PanicWithContext(ctx context.Context, args ...interface{}) {
	CreateRequestLogEntryFromContext(ctx, logrusLogger).Error(args)
}
func (logrusLogger *LogrusLogger) TracefWithContext(ctx context.Context, format string, args ...interface{}) {
	CreateRequestLogEntryFromContext(ctx, logrusLogger).Tracef(format, args)
}
func (logrusLogger *LogrusLogger) DebugfWithContext(ctx context.Context, format string, args ...interface{}) {
	CreateRequestLogEntryFromContext(ctx, logrusLogger).Debugf(format, args)
}
func (logrusLogger *LogrusLogger) InfofWithContext(ctx context.Context, format string, args ...interface{}) {
	CreateRequestLogEntryFromContext(ctx, logrusLogger).Infof(format, args)
}
func (logrusLogger *LogrusLogger) WarnfWithContext(ctx context.Context, format string, args ...interface{}) {
	CreateRequestLogEntryFromContext(ctx, logrusLogger).Warnf(format, args)
}
func (logrusLogger *LogrusLogger) ErrorfWithContext(ctx context.Context, format string, args ...interface{}) {
	CreateRequestLogEntryFromContext(ctx, logrusLogger).Errorf(format, args)
}
func (logrusLogger *LogrusLogger) FatalfWithContext(ctx context.Context, format string, args ...interface{}) {
	CreateRequestLogEntryFromContext(ctx, logrusLogger).Fatalf(format, args)
}
func (logrusLogger *LogrusLogger) PanicfWithContext(ctx context.Context, format string, args ...interface{}) {
	CreateRequestLogEntryFromContext(ctx, logrusLogger).Panicf(format, args)
}
