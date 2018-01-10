package sexagesimal_test

import (
	"testing"

	"github.com/odeke-em/go-utils/sexagesimal"
)

func TestSexagesimal(t *testing.T) {

	testCases := [...]struct {
		value int
		want  string
	}{
		0: {49, "00:00:49"},
		1: {60, "00:01:00"},
		2: {120, "00:02:00"},
		3: {1200, "00:20:00"},
		4: {61, "00:01:01"},
		5: {69, "00:01:09"},
	}

	for i, tt := range testCases {
		got, want := sexagesimal.Sexag(tt.value), tt.want
		if got != want {
			t.Errorf("%d: got=%v want=%v", i, got, want)
		}
	}
}
