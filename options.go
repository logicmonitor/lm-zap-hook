package lmzaphook

import (
	"time"

	"go.uber.org/zap/zapcore"
)

type Option func(*lmCore) error

// AuthProvider is used by the lmotel collector for providing auth token from derived in collector
type AuthProvider interface {
	GetCredentials(method string, uri string, body []byte) string
}

// WithClientBatchingInterval configures the batching interval
func WithClientBatchingInterval(batchingInterval time.Duration) Option {
	return func(lmCore *lmCore) error {
		lmCore.logNotifier.logIngesterSetting.clientBatchingInterval = batchingInterval
		return nil
	}
}

// WithClientBatchingDisabled disables the batching of logs
func WithClientBatchingDisabled() Option {
	return func(lmCore *lmCore) error {
		lmCore.logNotifier.logIngesterSetting.clientBatchingEnabled = false
		return nil
	}
}

// WithMetadata configures the metadata tags for the log messages
func WithMetadata(metadataTags map[string]string) Option {
	return func(lmCore *lmCore) error {
		lmCore.metadata = metadataTags
		return nil
	}
}

// WithLogLevel configures the log level such that only log messages with given level or above that level
// will be sent to the Logicmonitor platform
func WithLogLevel(level zapcore.Level) Option {
	return func(lmCore *lmCore) error {
		lmCore.LevelEnabler = level
		return nil
	}
}

// WithBlocking blocks for the calls to the send log operation
func WithBlocking() Option {
	return func(lmCore *lmCore) error {
		lmCore.logNotifier.async = false
		return nil
	}
}

// WithAuthProvider is used by the lmotel collector for passing the auth provider
func WithAuthProvider(authProvider AuthProvider) Option {
	return func(lmCore *lmCore) error {
		lmCore.logNotifier.logIngesterSetting.authProvider = authProvider
		return nil
	}
}

// WithNopLogIngesterClient will use no-op log ingester which will not send any logs. Can be used for testing
func WithNopLogIngesterClient() Option {
	return func(lmCore *lmCore) error {
		lmCore.logNotifier.LogIngesterClient = newNopLogIngesterClient()
		return nil
	}
}
