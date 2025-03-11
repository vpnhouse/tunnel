package eventlog

import (
	"time"
)

const (
	defaultLockTtl                = time.Minute
	defaultReportPositionInterval = 5 * time.Second
	defaultLockProlongateTimeout  = 30 * time.Second
	defaultWaitOutputWriteTimeout = 5 * time.Second
)
