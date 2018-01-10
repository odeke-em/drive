package pb

import (
	"bytes"
	"testing"
)

func TestProgressBarTemplate(t *testing.T) {
	// test New
	bar := ProgressBarTemplate(`{{counters . }}`).New(0)
	result := bar.String()
	expected := "0"
	if result != expected {
		t.Errorf("Unexpected result: (actual/expected)\n%s\n%s", result, expected)
	}
	if bar.IsStarted() {
		t.Error("Must be false")
	}

	// test Start
	bar = ProgressBarTemplate(`{{counters . }}`).Start(42).SetWriter(bytes.NewBuffer(nil))
	result = bar.String()
	expected = "0 / 42"
	if result != expected {
		t.Errorf("Unexpected result: (actual/expected)\n%s\n%s", result, expected)
	}
	if !bar.IsStarted() {
		t.Error("Must be true")
	}
}

func TestTemplateFuncs(t *testing.T) {
	var results = make(map[string]int)
	for i := 0; i < 100; i++ {
		r := rndcolor("s")
		results[r] = results[r] + 1
	}
	if len(results) < 6 {
		t.Errorf("Unexpected rndcolor results count: %v", len(results))
	}

	results = make(map[string]int)
	for i := 0; i < 100; i++ {
		r := rnd("1", "2", "3")
		results[r] = results[r] + 1
	}
	if len(results) != 3 {
		t.Errorf("Unexpected rnd results count: %v", len(results))
	}
	if r := rnd(); r != "" {
		t.Errorf("Unexpected rnd result: '%v'", r)
	}
}
