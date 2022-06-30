package lmzaphook

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLogIngesterClient(t *testing.T) {
	os.Setenv("LM_ACCOUNT", "localdev")
	defer os.Unsetenv("LM_ACCOUNT")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logIngesterClient, err := newLogIngesterClient(ctx, LogIngesterSetting{
		clientBatchingEnabled:  true,
		clientBatchingInterval: 1 * time.Second,
		resourceMapperTags:     map[string]string{"system.displayname": "test-device"},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, logIngesterClient)
}

func TestLogIngesterWrite(t *testing.T) {
	logIngesterClient := newNopLogIngesterClient()
	logNotifier := LogNotifier{
		logIngesterSetting: &LogIngesterSetting{
			resourceMapperTags: map[string]string{"system.displayname": "test-device"},
		},
		LogIngesterClient: logIngesterClient,
	}
	logMessage := "test"
	err := logNotifier.Notify([]byte(logMessage), map[string]string{"env": "dev"})
	assert.NoError(t, err)
}

func TestLogIngesterWriteAsync(t *testing.T) {
	logIngesterClient := newNopLogIngesterClient()
	logNotifier := LogNotifier{
		logIngesterSetting: &LogIngesterSetting{
			resourceMapperTags: map[string]string{"system.displayname": "test-device"},
		},
		async:             true,
		LogIngesterClient: logIngesterClient,
	}
	logMessage := "test"
	err := logNotifier.Notify([]byte(logMessage), map[string]string{"env": "dev"})
	assert.NoError(t, err)
}
