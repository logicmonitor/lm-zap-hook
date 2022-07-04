package lmzaphook

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewLMCore(t *testing.T) {
	resourceTags := map[string]string{"system.displayname": "test-device"}
	lmCore, err := NewLMCore(context.Background(), Params{ResourceMapperTags: resourceTags}, WithNopLogIngesterClient())
	assert.NoError(t, err)
	assert.Equal(t, resourceTags, lmCore.logNotifier.logIngesterSetting.resourceMapperTags)
	assert.Equal(t, false, lmCore.logNotifier.logIngesterSetting.clientBatchingEnabled)
	assert.Equal(t, true, lmCore.logNotifier.async)
}

func TestLMWithOptions(t *testing.T) {
	resourceTags := map[string]string{"system.displayname": "test-device"}
	metadataTags := map[string]string{"env": "staging"}

	lmCore, err := NewLMCore(context.Background(), Params{ResourceMapperTags: resourceTags},
		WithLogLevel(zapcore.InfoLevel),
		WithClientBatchingEnabled(1*time.Minute),
		WithMetadata(metadataTags),
		WithNopLogIngesterClient(),
	)
	assert.NoError(t, err)
	assert.Equal(t, resourceTags, lmCore.logNotifier.logIngesterSetting.resourceMapperTags)
	assert.Equal(t, metadataTags, lmCore.metadata)
	assert.Equal(t, true, lmCore.logNotifier.logIngesterSetting.clientBatchingEnabled)
	assert.Equal(t, 1*time.Minute, lmCore.logNotifier.logIngesterSetting.clientBatchingInterval)
	assert.Equal(t, true, lmCore.logNotifier.async)
}

func TestNewLMCoreWithError(t *testing.T) {
	resourceTags := map[string]string{}
	lmCore, err := NewLMCore(context.Background(), Params{ResourceMapperTags: resourceTags}, WithNopLogIngesterClient())
	assert.Error(t, err)
	assert.Nil(t, lmCore)
}

func TestCheck(t *testing.T) {
	entry := zapcore.Entry{
		Message:    "test",
		Level:      zapcore.InfoLevel,
		Time:       time.Now(),
		LoggerName: "main",
		Stack:      "fake-stack",
	}
	ce := &zapcore.CheckedEntry{}
	resourceTags := map[string]string{"system.displayname": "test-device"}
	lmCore, _ := NewLMCore(context.Background(), Params{ResourceMapperTags: resourceTags}, WithLogLevel(zapcore.InfoLevel), WithNopLogIngesterClient())
	assert.Equal(t, ce, lmCore.Check(entry, ce))
}

func TestCheckWithSkipLog(t *testing.T) {
	entry := zapcore.Entry{
		Message:    "test",
		Level:      zapcore.DebugLevel,
		Time:       time.Now(),
		LoggerName: "main",
		Stack:      "fake-stack",
	}
	ce := &zapcore.CheckedEntry{}
	resourceTags := map[string]string{"system.displayname": "test-device"}
	lmCore, _ := NewLMCore(context.Background(), Params{ResourceMapperTags: resourceTags}, WithLogLevel(zapcore.InfoLevel), WithNopLogIngesterClient())
	assert.Equal(t, ce, lmCore.Check(entry, ce))
}

func TestWith(t *testing.T) {
	resourceTags := map[string]string{"system.displayname": "test-device"}
	fields := []zapcore.Field{makeInt64Field("k", 42)}
	lmCore, _ := NewLMCore(context.Background(), Params{ResourceMapperTags: resourceTags}, WithLogLevel(zapcore.InfoLevel), WithNopLogIngesterClient())
	clone := lmCore.clone()
	fields[0].AddTo(clone.enc)
	assert.Equal(t, clone, lmCore.With(fields))
}

func TestWrite(t *testing.T) {
	entry := zapcore.Entry{
		Message:    "test",
		Level:      zapcore.InfoLevel,
		Time:       time.Now(),
		LoggerName: "main",
		Stack:      "fake-stack",
	}
	fields := []zapcore.Field{makeInt64Field("k", 42)}
	resourceTags := map[string]string{"system.displayname": "test-device"}
	metadataTags := map[string]string{"env": "staging"}
	lmCore, _ := NewLMCore(context.Background(), Params{ResourceMapperTags: resourceTags}, WithLogLevel(zapcore.InfoLevel), WithMetadata(metadataTags), WithNopLogIngesterClient())
	assert.NoError(t, lmCore.Write(entry, fields))
}

func TestWriteWithAsyncDisabled(t *testing.T) {
	entry := zapcore.Entry{
		Message:    "test",
		Level:      zapcore.InfoLevel,
		Time:       time.Now(),
		LoggerName: "main",
		Stack:      "fake-stack",
	}
	fields := []zapcore.Field{makeInt64Field("k", 42)}
	resourceTags := map[string]string{"system.displayname": "test-device"}
	metadataTags := map[string]string{"env": "staging"}
	lmCore, _ := NewLMCore(context.Background(), Params{ResourceMapperTags: resourceTags}, WithLogLevel(zapcore.InfoLevel), WithBlocking(), WithMetadata(metadataTags), WithNopLogIngesterClient())
	assert.NoError(t, lmCore.Write(entry, fields))
}

func BenchmarkLMCore(b *testing.B) {
	os.Setenv("LM_ACCOUNT", "localdev")
	defer os.Unsetenv("LM_ACCOUNT")
	resourceTags := map[string]string{"system.displayname": "test-device"}
	metadataTags := map[string]string{"env": "staging"}
	logger := zap.NewNop()

	lmCore, _ := NewLMCore(context.Background(), Params{ResourceMapperTags: resourceTags},
		WithLogLevel(zapcore.InfoLevel),
		WithClientBatchingEnabled(1*time.Minute),
		WithMetadata(metadataTags),
		WithNopLogIngesterClient(),
	)
	opt := zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, lmCore)
	})
	logger = logger.WithOptions(opt)
	for i := 0; i < b.N; i++ {
		logger.Info("test log", zap.String("DemoKey", "DemoValue"))
	}
}

func makeInt64Field(key string, val int) zapcore.Field {
	return zapcore.Field{Type: zapcore.Int64Type, Integer: int64(val), Key: key}
}
