package eventlog

import (
	"errors"
)

var ErrOutputEventStucked = errors.New("output event stucked")
var ErrLockNotAcquired = errors.New("lock not acquired")
