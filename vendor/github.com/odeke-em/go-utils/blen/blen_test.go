package blen

import (
	"testing"
)

func TestBlen(t *testing.T) {
	tests := []struct {
		q    uint64
		want int
	}{
		0: {q: 0, want: 1},
		1: {q: 10, want: 4},
		2: {q: 64, want: 7},
		3: {q: 0xf, want: 4},
		4: {q: ^uint64(0), want: 64},
		5: {q: ^uint64(0)>>1, want: 63},
		6: {q: 1<<34|1<<32, want: 35},
	}

	for i, tt := range tests {
		got := Blen(tt.q)
		if got != tt.want {
			t.Errorf("#%d: %d got=%d want=%d", i, tt.q, got, tt.want)
		}
	}
}
