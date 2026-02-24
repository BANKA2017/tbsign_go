package share

import "sync/atomic"

var CrontabBypassTimes atomic.Int64

var MockReleasedVersion = false
