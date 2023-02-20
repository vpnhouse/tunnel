package eventlog

import (
	"time"
)

const (
	reportOffsetTimeout          = 5 * time.Second
	lockTtl                      = time.Minute
	defaultLockProlongateTimeout = 30 * time.Second

	waitOutputWriteTimeout = time.Second
)
