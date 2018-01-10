package sexagesimal

import (
	"fmt"
	"strings"
)

const (
	second = 0
	minute = 1
	hour   = 2
	day    = 3
)

var positions = map[int]int{
	second: 60,
	minute: 60,
	hour:   24,
	day:    7,
}

func segments(t int) (all []int) {
	i := 0
	for i < len(positions) {
		base := positions[i]
		if base == 0 {
			base = 60
		}

		rem := t % base
		t /= base
		all = append(all, rem)

		if t <= 0 {
			break
		}
		i += 1
	}

	return
}

func sexag(t int) []string {
	rev := segments(t)

	properSeg := make([]string, len(rev))
	for i, n := 0, len(rev); i < n; i++ {
		properSeg[i] = fmt.Sprintf("%02d", rev[n-i-1])
	}

	return properSeg
}

func Sexag(t int) string {
	minSegmentLen := 3
	segments := sexag(t)
	for i := len(segments); i < minSegmentLen; i++ {
		segments = append([]string{"00"}, segments...)
	}

	return strings.Join(segments, ":")
}
