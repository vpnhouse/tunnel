package manager

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPeerSessions(t *testing.T) {
	ts := time.Date(2023, 03, 01, 10, 0, 0, 0, time.UTC)
	updateInterval := 5 * time.Minute
	stat := newRuntimePeerStat(ts.Unix(), 0, 0, "")

	ts = ts.Add(time.Minute)
	stat.Update(ts, 1, 1, "", updateInterval)
	require.Equal(t, 1, len(stat.Sessions()))

	ts = ts.Add(time.Second)
	stat.Update(ts, 2, 2, "", updateInterval)
	require.Equal(t, 1, len(stat.Sessions()))

	ts = ts.Add(time.Hour)
	stat.Update(ts, 3, 3, "", updateInterval)
	require.Equal(t, 2, len(stat.Sessions()))
}
