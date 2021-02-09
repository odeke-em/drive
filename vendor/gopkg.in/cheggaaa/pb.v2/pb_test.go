package pb

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"gopkg.in/fatih/color.v1"
)

func TestPBBasic(t *testing.T) {
	bar := new(ProgressBar)
	var a, e int64
	if a, e = bar.Total(), 0; a != e {
		t.Errorf("Unexpected total: actual: %v; expected: %v", a, e)
	}
	if a, e = bar.Current(), 0; a != e {
		t.Errorf("Unexpected current: actual: %v; expected: %v", a, e)
	}
	bar.SetCurrent(10).SetTotal(20)
	if a, e = bar.Total(), 20; a != e {
		t.Errorf("Unexpected total: actual: %v; expected: %v", a, e)
	}
	if a, e = bar.Current(), 10; a != e {
		t.Errorf("Unexpected current: actual: %v; expected: %v", a, e)
	}
	bar.Add(5)
	if a, e = bar.Current(), 15; a != e {
		t.Errorf("Unexpected current: actual: %v; expected: %v", a, e)
	}
	bar.Increment()
	if a, e = bar.Current(), 16; a != e {
		t.Errorf("Unexpected current: actual: %v; expected: %v", a, e)
	}
}

func TestPBWidth(t *testing.T) {
	terminalWidth = func() (int, error) {
		return 50, nil
	}
	// terminal width
	bar := new(ProgressBar)
	if a, e := bar.Width(), 50; a != e {
		t.Errorf("Unexpected width: actual: %v; expected: %v", a, e)
	}
	// terminal width error
	terminalWidth = func() (int, error) {
		return 0, errors.New("test error")
	}
	if a, e := bar.Width(), defaultBarWidth; a != e {
		t.Errorf("Unexpected width: actual: %v; expected: %v", a, e)
	}
	// terminal width panic
	terminalWidth = func() (int, error) {
		panic("test")
		return 0, nil
	}
	if a, e := bar.Width(), defaultBarWidth; a != e {
		t.Errorf("Unexpected width: actual: %v; expected: %v", a, e)
	}
	// set negative terminal width
	bar.SetWidth(-42)
	if a, e := bar.Width(), defaultBarWidth; a != e {
		t.Errorf("Unexpected width: actual: %v; expected: %v", a, e)
	}
	// set terminal width
	bar.SetWidth(42)
	if a, e := bar.Width(), 42; a != e {
		t.Errorf("Unexpected width: actual: %v; expected: %v", a, e)
	}
}

func TestPBTemplate(t *testing.T) {
	bar := new(ProgressBar)
	result := bar.SetTotal(100).SetCurrent(50).SetWidth(40).String()
	expected := "50 / 100 [------->________] 50.00% ? p/s"
	if result != expected {
		t.Errorf("Unexpected result: (actual/expected)\n%s\n%s", result, expected)
	}

	// check strip
	result = bar.SetWidth(8).String()
	expected = "50 / 100"
	if result != expected {
		t.Errorf("Unexpected result: (actual/expected)\n%s\n%s", result, expected)
	}

	// invalid template
	for _, invalidTemplate := range []string{
		`{{invalid template`, `{{speed}}`,
	} {
		bar.SetTemplateString(invalidTemplate)
		result = bar.String()
		expected = ""
		if result != expected {
			t.Errorf("Unexpected result: (actual/expected)\n%s\n%s", result, expected)
		}
		if err := bar.Err(); err == nil {
			t.Errorf("Must be error")
		}
	}

	// simple template without adaptive elemnts
	bar.SetTemplateString(`{{counters . }}`)
	result = bar.String()
	expected = "50 / 100"
	if result != expected {
		t.Errorf("Unexpected result: (actual/expected)\n%s\n%s", result, expected)
	}
}

func TestPBStartFinish(t *testing.T) {
	bar := ProgressBarTemplate(`{{counters . }}`).New(0)
	for i := int64(0); i < 2; i++ {
		if bar.IsStarted() {
			t.Error("Must be false")
		}
		var buf = bytes.NewBuffer(nil)
		bar.SetTotal(100).
			SetCurrent(int64(i)).
			SetWidth(7).
			Set(Terminal, true).
			SetWriter(buf).
			SetRefreshRate(time.Millisecond * 20).
			Start()
		if !bar.IsStarted() {
			t.Error("Must be true")
		}
		time.Sleep(time.Millisecond * 100)
		bar.Finish()
		if buf.Len() == 0 {
			t.Error("no writes")
		}
		var resultsString = strings.TrimPrefix(buf.String(), "\r")
		if !strings.HasSuffix(resultsString, "\n") {
			t.Error("No end \\n symb")
		} else {
			resultsString = resultsString[:len(resultsString)-1]
		}
		var results = strings.Split(resultsString, "\r")
		if len(results) < 3 {
			t.Errorf("Unexpected writes count: %v", len(results))
		}
		exp := fmt.Sprintf("%d / 100", i)
		for i, res := range results {
			if res != exp {
				t.Errorf("Unexpected result[%d]: '%v'", i, res)
			}
		}
		// test second finish call
		bar.Finish()
	}
}

func TestPBFlags(t *testing.T) {
	// Static
	color.NoColor = false
	buf := bytes.NewBuffer(nil)
	bar := ProgressBarTemplate(`{{counters . | red}}`).New(100)
	bar.Set(Static, true).SetCurrent(50).SetWidth(10).SetWriter(buf).Start()
	if bar.IsStarted() {
		t.Error("Must be false")
	}
	bar.Write()
	result := buf.String()
	expected := "50 / 100"
	if result != expected {
		t.Errorf("Unexpected result: (actual/expected)\n'%s'\n'%s'", result, expected)
	}
	if !bar.state.IsFirst() {
		t.Error("must be true")
	}
	// Color
	bar.Set(Color, true)
	buf.Reset()
	bar.Write()
	result = buf.String()
	expected = color.RedString("50 / 100")
	if result != expected {
		t.Errorf("Unexpected result: (actual/expected)\n'%s'\n'%s'", result, expected)
	}
	if bar.state.IsFirst() {
		t.Error("must be false")
	}
	// Terminal
	bar.Set(Terminal, true).SetWriter(buf)
	buf.Reset()
	bar.Write()
	result = buf.String()
	expected = "\r" + color.RedString("50 / 100") + "  "
	if result != expected {
		t.Errorf("Unexpected result: (actual/expected)\n'%s'\n'%s'", result, expected)
	}
}

func BenchmarkRender(b *testing.B) {
	var formats = []string{
		string(Simple),
		string(Default),
		string(Full),
		`{{string . "prefix" | red}}{{counters . | green}} {{bar . | yellow}} {{percent . | cyan}} {{speed . | cyan}}{{string . "suffix" | cyan}}`,
	}
	var names = []string{
		"Simple", "Default", "Full", "Color",
	}
	for i, tmpl := range formats {
		bar := new(ProgressBar)
		bar.SetTemplateString(tmpl).SetWidth(100)
		b.Run(names[i], func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				bar.String()
			}
		})
	}
}
