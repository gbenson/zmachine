package util

import (
	"time"

	"golang.org/x/sys/unix"
)

// Uptime returns a CLOCK_MONOTONIC timestamp à la dmesg.
func Uptime() (time.Duration, error) {
	var ts unix.Timespec
	if err := unix.ClockGettime(unix.CLOCK_MONOTONIC, &ts); err != nil {
		return 0, err
	}

	return time.Duration(ts.Nano()) * time.Nanosecond, nil
}
