package mlog

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTimeStamp(t *testing.T) {
	ts := GetTimeStamp()
	matched, err := regexp.MatchString(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`, ts)
	require.NoError(t, err)
	require.True(t, matched, "unexpected timestamp format: %s", ts)
}

func TestSetLogLevel(t *testing.T) {
	SetLogLevel("test-level", 3)
	logger := getLogger("test-level")
	require.Equal(t, 3, logger.Level)
}

func TestSetStoreDays(t *testing.T) {
	SetStoreDays("test-store", 14)
	logger := getLogger("test-store")
	require.Equal(t, 14, logger.StoreDays)
}

func TestSetLogUseOwnDir(t *testing.T) {
	SetLogUseOwnDir("test-dir", true)
	logger := getLogger("test-dir")
	require.True(t, logger.UseOwnDir)
}
