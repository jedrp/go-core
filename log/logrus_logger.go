package log

import (
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

	LogConfigStr string `long:"log-config" description:"the hook connection string" env:"LOG_CONFIG" json:"logConfigStr,omitempty"`
	logLevel     string
	*logrus.Logger
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
	log := logrus.New()
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

	logSetting, _ := json.Marshal(logrusLogger)
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

// NewEntry
func NewEntry(logger *LogrusLogger) *logrus.Entry {
	return &logrus.Entry{
		Logger: logger.Logger,
		// Default is three fields, plus one optional.  Give a little extra room.
		Data: make(logrus.Fields, 6),
	}
}
