//go:build darwin

package main

import (
	"fmt"
	"time"
)

func txtForDuration(d time.Duration) string {
	if d > time.Hour {
		return fmt.Sprintf("%.0fh:%.fm:%.fs", d.Hours(), d.Minutes(), d.Seconds())
	} else if d > time.Minute {
		return fmt.Sprintf("%.0fm:%.0fs", d.Minutes(), d.Seconds())
	} else {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
}

func iconForDuration(d time.Duration) []byte {
	return nil
}

func iconInactive() []byte {
	return nil
}
