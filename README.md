# lm-zap-hook

**lm-zap-hook** is a `zapcore.Core` implementation to integrate with the Logicmonitor Platform. It sends the log messages generated by the application, to the Logicmonitor platform.
## Installation

`go get -u github.com/logicmonitor/lm-zap-hook`

## Quick Start

### Authentication:

Set the `LM_ACCESS_ID` and `LM_ACCESS_KEY` for using the LMv1 authentication. The company name or account name must be set to `LM_ACCOUNT` property. All properties can be set using environment variable.

| Environment variable |	Description                                        |
| -------------------- | ------------------------------------------------------|
|   LM_ACCOUNT         | Account name (Company Name) is your organization name |
|   LM_ACCESS_ID       | Access id while using LMv1 authentication.|
|   LM_ACCESS_KEY      | Access key while using LMv1 authentication.|

### Getting Started

Here's an example code snippet for configuring the `lm-zap-hook` with the application code.

```go
package main

import (
	"context"
	"time"

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

	// create a new core that sends zapcore.InfoLevel and above messages to Logicmonitor Platform
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

	// By default, log send operations happens async way, so blocking the execution
	time.Sleep(15 * time.Second)
}

```
### Options

Following are the options that can be passed to `NewLMCore()` to configure the `lmCore`.

| Option                                     |   Description                                                                    |             
|--------------------------------------------|----------------------------------------------------------------------------------|
|   WithLogLevel(`logLevel zapcore.Level`)                   | Configures `lmCore` to send the logs having level equal or above the level specified by `logLevel`. Default logLevel is `Warning`. |
|   WithClientBatchingInterval(`batchInterval time.Duration`) | Configures interval for batching of the log messages. |
|   WithClientBatchingDisabled() | Disables the batching of log messages. By default, batching is enabled. |
|   WithMetadata(`metadata map[string]string`)                   | Metadata to be sent with the every log message.                                    |
|   WithNopLogIngesterClient()               | Configures `lmCore` to use the nopLogIngesterClient which discards the log messages. It can be used for testing.                          |
|   WithBlocking()      | It makes the call to the send log operation blocking. Default value of Async Mode is `true`. |
---

Copyright, 2022, LogicMonitor, Inc.

This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0. If a copy of the MPL was not distributed with this file, You can obtain one at https://mozilla.org/MPL/2.0/.