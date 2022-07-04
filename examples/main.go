package main

import (
	"context"

	lmzaphook "github.com/logicmonitor/lm-zap-hook"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// create a new Zap logger
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any

	// create resource tags for mapping the log messages to a unique LogicMonitor resource
	resourceTags := map[string]string{"system.displayname": "test-device"}

	// create a new core that sends zapcore.WarnLevel and above messages to Logicmonitor Platform
	lmCore, err := lmzaphook.NewLMCore(context.Background(),
		lmzaphook.Params{ResourceMapperTags: resourceTags},
		lmzaphook.WithLogLevel(zapcore.WarnLevel),
	)
	if err != nil {
		logger.Fatal(err.Error())
	}

	// Wrap a NewTee to send log messages to both your main logger and to Logicmonitor
	logger = logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, lmCore)
	}))

	// This message will only go to the main logger
	logger.Info("Test log message for main logger", zap.String("foo", "bar"))

	// This warning will go to both the main logger and to Logicmonitor.
	logger.Warn("Warning message with fields", zap.String("foo", "bar"))
}
