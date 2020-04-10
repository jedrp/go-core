package log

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jedrp/go-core/util"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

var (
	// ErrCannotCreateIndex Fired if the index is not created
	ErrCannotCreateIndex = fmt.Errorf("cannot create index")
)

// IndexNameFunc get index name
type IndexNameFunc func() string

type fireFunc func(entry *logrus.Entry, hook *ElasticHook) error

// ElasticHook is a logrus
// hook for ElasticSearch
type ElasticHook struct {
	client    *elastic.Client
	host      string
	index     IndexNameFunc
	levels    []logrus.Level
	ctx       context.Context
	ctxCancel context.CancelFunc
	fireFunc  fireFunc
}

type message struct {
	Host      string        `json:"host"`
	Timestamp string        `json:"@timestamp"`
	Message   string        `json:"message"`
	Fields    logrus.Fields `json:"fields"`
	Level     string        `json:"level"`
}

const (
	Host        = "host"
	IndexPrefix = "index-prefix"
	Sniff       = "sniff"
	Mode        = "mode"
	UserName    = "username"
	Password    = "password"
)

func NewElasticHookFromStr(configStr string, level logrus.Level) (*ElasticHook, error) {
	config, err := util.GetConfig(configStr)
	if err != nil {
		log.Panic(err)
	}
	host, found := config[Host]
	if !found || host == "" {
		log.Panic("host empty is not valid")
	}
	indexPrefix, found := config[IndexPrefix]
	if !found || indexPrefix == "" {
		log.Panic("indexPrefix empty is not valid")
	}
	indexFunc := func() string {
		dt := time.Now()
		return fmt.Sprintf("%s-%s", indexPrefix, dt.Format("2006-01-02"))
	}
	sniffEnableStr, found := config[Sniff]
	sniffEnable := false
	if found && sniffEnableStr == "true" {
		sniffEnable = true
	}
	var clientOptionFuncs []elastic.ClientOptionFunc
	clientOptionFuncs = append(clientOptionFuncs, elastic.SetSniff(sniffEnable))
	clientOptionFuncs = append(clientOptionFuncs, elastic.SetURL(strings.Split(host, ",")...))

	if un, f := config[UserName]; f {
		if un != "" {
			pass, f := config[Password]
			if f && pass != "" {
				clientOptionFuncs = append(clientOptionFuncs, elastic.SetBasicAuth(un, pass))
			}
		}
	}

	client, err := elastic.NewClient(clientOptionFuncs...)
	if err != nil {
		log.Panic(err)
	}

	mode, found := config[Mode]
	if found && mode == "async" {
		return NewAsyncElasticHookWithFunc(client, host, level, indexFunc)
	}
	return NewElasticHookWithFunc(client, host, level, indexFunc)
}

// NewElasticHook creates new hook.
// client - ElasticSearch client with specific es version (v5/v6/v7/...)
// host - host of system
// level - log level
// index - name of the index in ElasticSearch
func NewElasticHook(client *elastic.Client, host string, level logrus.Level, index string) (*ElasticHook, error) {
	return NewElasticHookWithFunc(client, host, level, func() string { return index })
}

// NewAsyncElasticHook creates new  hook with asynchronous log.
// client - ElasticSearch client with specific es version (v5/v6/v7/...)
// host - host of system
// level - log level
// index - name of the index in ElasticSearch
func NewAsyncElasticHook(client *elastic.Client, host string, level logrus.Level, index string) (*ElasticHook, error) {
	return NewAsyncElasticHookWithFunc(client, host, level, func() string { return index })
}

// NewBulkProcessorElasticHook creates new hook that uses a bulk processor for indexing.
// client - ElasticSearch client with specific es version (v5/v6/v7/...)
// host - host of system
// level - log level
// index - name of the index in ElasticSearch
func NewBulkProcessorElasticHook(client *elastic.Client, host string, level logrus.Level, index string) (*ElasticHook, error) {
	return NewBulkProcessorElasticHookWithFunc(client, host, level, func() string { return index })
}

// NewElasticHookWithFunc creates new hook with
// function that provides the index name. This is useful if the index name is
// somehow dynamic especially based on time.
// client - ElasticSearch client with specific es version (v5/v6/v7/...)
// host - host of system
// level - log level
// indexFunc - function providing the name of index
func NewElasticHookWithFunc(client *elastic.Client, host string, level logrus.Level, indexFunc IndexNameFunc) (*ElasticHook, error) {
	return newHookFuncAndFireFunc(client, host, level, indexFunc, syncFireFunc)
}

// NewAsyncElasticHookWithFunc creates new asynchronous hook with
// function that provides the index name. This is useful if the index name is
// somehow dynamic especially based on time.
// client - ElasticSearch client with specific es version (v5/v6/v7/...)
// host - host of system
// level - log level
// indexFunc - function providing the name of index
func NewAsyncElasticHookWithFunc(client *elastic.Client, host string, level logrus.Level, indexFunc IndexNameFunc) (*ElasticHook, error) {
	return newHookFuncAndFireFunc(client, host, level, indexFunc, asyncFireFunc)
}

// NewBulkProcessorElasticHookWithFunc creates new hook with
// function that provides the index name. This is useful if the index name is
// somehow dynamic especially based on time that uses a bulk processor for
// indexing.
// client - ElasticSearch client with specific es version (v5/v6/v7/...)
// host - host of system
// level - log level
// indexFunc - function providing the name of index
func NewBulkProcessorElasticHookWithFunc(client *elastic.Client, host string, level logrus.Level, indexFunc IndexNameFunc) (*ElasticHook, error) {
	fireFunc, err := makeBulkFireFunc(client)
	if err != nil {
		return nil, err
	}
	return newHookFuncAndFireFunc(client, host, level, indexFunc, fireFunc)
}

func newHookFuncAndFireFunc(client *elastic.Client, host string, level logrus.Level, indexFunc IndexNameFunc, fireFunc fireFunc) (*ElasticHook, error) {
	var levels []logrus.Level
	for _, l := range []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	} {
		if l <= level {
			levels = append(levels, l)
		}
	}

	ctx, cancel := context.WithCancel(context.TODO())

	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists(indexFunc()).Do(ctx)
	if err != nil {
		// Handle error
		cancel()
		return nil, err
	}
	if !exists {
		createIndex, err := client.CreateIndex(indexFunc()).Do(ctx)
		if err != nil {
			cancel()
			return nil, err
		}
		if !createIndex.Acknowledged {
			cancel()
			return nil, ErrCannotCreateIndex
		}
	}

	return &ElasticHook{
		client:    client,
		host:      host,
		index:     indexFunc,
		levels:    levels,
		ctx:       ctx,
		ctxCancel: cancel,
		fireFunc:  fireFunc,
	}, nil
}

// Fire is required to implement
// Logrus hook
func (hook *ElasticHook) Fire(entry *logrus.Entry) error {
	return hook.fireFunc(entry, hook)
}

func asyncFireFunc(entry *logrus.Entry, hook *ElasticHook) error {
	go syncFireFunc(entry, hook)
	return nil
}

func createMessage(entry *logrus.Entry, hook *ElasticHook) *message {
	level := entry.Level.String()

	if e, ok := entry.Data[logrus.ErrorKey]; ok && e != nil {
		if err, ok := e.(error); ok {
			entry.Data[logrus.ErrorKey] = err.Error()
		}
	}

	return &message{
		hook.host,
		entry.Time.UTC().Format(time.RFC3339Nano),
		entry.Message,
		entry.Data,
		strings.ToUpper(level),
	}
}

func syncFireFunc(entry *logrus.Entry, hook *ElasticHook) error {
	_, err := hook.client.
		Index().
		Index(hook.index()).
		Type("log").
		BodyJson(*createMessage(entry, hook)).
		Do(hook.ctx)

	return err
}

// Create closure with bulk processor tied to fireFunc.
func makeBulkFireFunc(client *elastic.Client) (fireFunc, error) {
	processor, err := client.BulkProcessor().
		Name("elogrus.v3.bulk.processor").
		Workers(2).
		FlushInterval(time.Second).
		Do(context.Background())

	return func(entry *logrus.Entry, hook *ElasticHook) error {
		r := elastic.NewBulkIndexRequest().
			Index(hook.index()).
			Type("log").
			Doc(*createMessage(entry, hook))
		processor.Add(r)
		return nil
	}, err
}

// Levels Required for logrus hook implementation
func (hook *ElasticHook) Levels() []logrus.Level {
	return hook.levels
}

// Cancel all calls to elastic
func (hook *ElasticHook) Cancel() {
	hook.ctxCancel()
}
