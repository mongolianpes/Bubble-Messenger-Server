package handlers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCheckRequestLatencyOld(t *testing.T) {
	t.Run("time is less than maxRequestLatency", func(t *testing.T) {
		start := time.Now().Add(-29 * time.Second).Format(time.RFC3339)
		isOld, err := checkRequestLatencyOld(start)
		require.NoError(t, err)
		require.False(t, isOld)
	})

	t.Run("time is greater than maxRequestLatency", func(t *testing.T) {
		start := time.Now().Add(-31 * time.Second).Format(time.RFC3339)
		isOld, err := checkRequestLatencyOld(start)
		require.NoError(t, err)
		require.True(t, isOld)
	})

	t.Run("invalid time format", func(t *testing.T) {
		start := "invalid time"
		_, err := checkRequestLatencyOld(start)
		require.Error(t, err)
	})
}
