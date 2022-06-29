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

// WithClientBatchingEnabled enables the client side batching and configures the batching interval
func WithClientBatchingEnabled(batchingInterval time.Duration) Option {
	return func(lmCore *lmCore) error {
		lmCore.logIngester.logIngesterSetting.clientBatchingEnabled = true
		lmCore.logIngester.logIngesterSetting.clientBatchingInterval = batchingInterval
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
// will be sent
func WithLogLevel(level zapcore.Level) Option {
	return func(lmCore *lmCore) error {
		lmCore.LevelEnabler = level
		return nil
	}
}

// WithAsync allows to send the log requests in async manner.
// Note: It does not take any effect when batching is enabled
func WithAsync() Option {
	return func(lmCore *lmCore) error {
		lmCore.logIngester.async = true
		return nil
	}
}

// WithAuthProvider is used by the lmotel collector for passing the auth provider
func WithAuthProvider(authProvider AuthProvider) Option {
	return func(lmCore *lmCore) error {
		lmCore.logIngester.logIngesterSetting.authProvider = authProvider
		return nil
	}
}

// WithNopLogIngesterClient will use no-op log ingester which will not send any logs. Can be used for testing
func WithNopLogIngesterClient() Option {
	return func(lmCore *lmCore) error {
		lmCore.logIngester.LogIngesterClient = newNopLogIngesterClient()
		return nil
	}
}
