package lmzaphook

import (
	"context"
	"time"

	"github.com/logicmonitor/lm-data-sdk-go/api/logs"
)

// LogNotifier holds the log ingester config
type LogNotifier struct {
	LogIngesterClient
	logIngesterSetting *LogIngesterSetting
	async              bool
}

// LogIngesterClient represents interface for Log Ingest client
type LogIngesterClient interface {
	SendLogs(ctx context.Context, logMessage string, resourceidMap, metadata map[string]string) error
}

// nopLogIngesterClient for testing
type nopLogIngesterClient struct {
}

// LogIngesterSetting holds the properties required for configuring the LMLogIngest instance
type LogIngesterSetting struct {
	clientBatchingEnabled  bool
	clientBatchingInterval time.Duration
	resourceMapperTags     map[string]string
	authProvider           AuthProvider
}

// newLogIngesterClient returns the LM LogIngest client instance from the lm-data-sdk-go
func newLogIngesterClient(ctx context.Context, logIngesterSetting LogIngesterSetting) (*logs.LMLogIngest, error) {
	var opts []logs.Option
	// if batching is enabled, then check for interval
	if logIngesterSetting.clientBatchingEnabled {
		if logIngesterSetting.clientBatchingInterval > 0 {
			opts = append(opts, logs.WithLogBatchingInterval(logIngesterSetting.clientBatchingInterval))
		}
	} else {
		// disable batching
		opts = append(opts, logs.WithLogBatchingDisabled())
	}
	if logIngesterSetting.authProvider != nil {
		opts = append(opts, logs.WithAuthentication(logIngesterSetting.authProvider))
	}
	return logs.NewLMLogIngest(ctx, opts...)
}

func newNopLogIngesterClient() nopLogIngesterClient {
	return nopLogIngesterClient{}
}

// nopLogIngesterClient SendLog implementation for testing
func (nopIngesterClient nopLogIngesterClient) SendLogs(ctx context.Context, logMessage string, resourceidMap, metadata map[string]string) error {
	return nil
}

func (logNotifier *LogNotifier) Notify(ctx context.Context, data []byte, metadata map[string]string) error {
	var err error
	// Sending logs in async mode will make sense only if the batching is disabled
	if !logNotifier.logIngesterSetting.clientBatchingEnabled && logNotifier.async {
		go func() {
			_ = logNotifier.LogIngesterClient.SendLogs(ctx, string(data), logNotifier.logIngesterSetting.resourceMapperTags, metadata)
		}()
		return nil
	}
	// Sending logs in synchronus mode
	err = logNotifier.LogIngesterClient.SendLogs(ctx, string(data), logNotifier.logIngesterSetting.resourceMapperTags, metadata)
	if err != nil {
		return err
	}
	return nil
}
