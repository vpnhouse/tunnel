package eventlog

import "time"

const (
	reportOffsetTimeout = 5 * time.Second
	lockTtl             = time.Minute

	waitOutputWriteTimeout = time.Second
)
