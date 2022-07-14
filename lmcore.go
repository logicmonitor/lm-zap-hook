package lmzaphook

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// Default log level to be used for getting the log levels to be considered for reporting the logs
	defaultLogLevel = zapcore.WarnLevel
	// Default value to decide if log send operation to be performed in async mode
	defaultAsync = true
	// Default batching interval
	defaultBatchingInterval = 10 * time.Second
)

// Params holds the required configurations for the hook
type Params struct {
	// Resource tags for mapping the log messages to a unique LogicMonitor resource
	ResourceMapperTags map[string]string
}

type lmCore struct {
	logNotifier LogNotifier
	zapcore.LevelEnabler
	enc      zapcore.Encoder
	metadata map[string]string
}

// NewCore creates a zap core that sends out the logs using logNotifer
func NewLMCore(ctx context.Context, params Params, opts ...Option) (*lmCore, error) {
	var err error

	// validate the config params
	if err := validate(params); err != nil {
		return nil, err
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// create core config
	lmCore := &lmCore{
		logNotifier: LogNotifier{
			logIngesterSetting: &LogIngesterSetting{
				resourceMapperTags:     params.ResourceMapperTags,
				clientBatchingEnabled:  true,
				clientBatchingInterval: defaultBatchingInterval,
			},
			async: defaultAsync,
		},
		LevelEnabler: defaultLogLevel,
		enc:          zapcore.NewConsoleEncoder(encoderConfig),
	}

	// apply options
	for _, opt := range opts {
		if err := opt(lmCore); err != nil {
			return nil, err
		}
	}

	if lmCore.logNotifier.LogIngesterClient == nil {
		lmCore.logNotifier.LogIngesterClient, err = newLogIngesterClient(ctx, *lmCore.logNotifier.logIngesterSetting)
		if err != nil {
			return nil, err
		}
	}
	return lmCore, nil
}

func validate(params Params) error {
	if len(params.ResourceMapperTags) == 0 {
		return errors.New("hook initialization failed: resourceMapperTags are not set")
	}
	return nil
}

func (c *lmCore) Write(entry zapcore.Entry, fs []zapcore.Field) error {
	clone := c.with(fs)
	buf, err := c.enc.EncodeEntry(entry, fs)
	if err != nil {
		return err
	}
	addMetadata(clone.metadata, entry)
	err = c.logNotifier.Notify(context.Background(), buf.Bytes(), clone.metadata)
	buf.Free()
	if err != nil {
		return err
	}
	if entry.Level > zapcore.ErrorLevel {
		// Since we may be crashing the program, sync the output.
		// TODO: Proper implementation of sync is pending
		if err := c.Sync(); err != nil {
			return err
		}
	}
	return nil
}

func (c *lmCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}
func (c *lmCore) Sync() error {
	return nil
}

func (c *lmCore) With(fs []zapcore.Field) zapcore.Core {
	return c.with(fs)
}

func (c *lmCore) with(fs []zapcore.Field) *lmCore {
	clone := c.clone()
	addFields(clone.enc, fs)
	return clone
}

func addMetadata(metadataTags map[string]string, entry zapcore.Entry) {
	if metadataTags != nil {
		metadataTags["level"] = entry.Level.String()
		metadataTags["function"] = entry.Caller.Function
		if entry.LoggerName != "" {
			metadataTags["logger"] = entry.LoggerName
		}
		trimmedCallerPath := entry.Caller.TrimmedPath()
		if trimmedCallerPath != "" {
			metadataTags["caller"] = trimmedCallerPath
		}
	}
}

func (c *lmCore) clone() *lmCore {
	metadata := make(map[string]string, len(c.metadata))
	for key, value := range c.metadata {
		metadata[key] = value
	}
	return &lmCore{
		logNotifier:  c.logNotifier,
		LevelEnabler: c.LevelEnabler,
		enc:          c.enc.Clone(),
		metadata:     metadata,
	}
}

func addFields(enc zapcore.ObjectEncoder, fields []zapcore.Field) {
	for i := range fields {
		fields[i].AddTo(enc)
	}
}
