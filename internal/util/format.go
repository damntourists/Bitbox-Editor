package util

import (
	"fmt"
	"time"
)

func SecondsLabel(sec float64) string {
	// mm:ss.mmm
	d := time.Duration(sec * float64(time.Second))
	minutes := int(d / time.Minute)
	rem := d % time.Minute
	return fmt.Sprintf("%02d:%06.2f", minutes, rem.Seconds())
}
