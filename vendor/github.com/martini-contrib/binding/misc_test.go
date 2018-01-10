package binding

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-martini/martini"
)

// When binding from Form data, testing the type of data to bind
// and converting a string into that type is tedious, so these tests
// cover all those cases.
func TestSetWithProperType(t *testing.T) {
	testInputs := map[string]string{
		"successful": `integer=-1&integer8=-8&integer16=-16&integer32=-32&integer64=-64&uinteger=1&uinteger8=8&uinteger16=16&uinteger32=32&uinteger64=64&boolean_1=true&fl32_1=32.3232&fl64_1=-64.6464646464&str=string`,
		"errorful":   `integer=&integer8=asdf&integer16=--&integer32=&integer64=dsf&uinteger=&uinteger8=asdf&uinteger16=+&uinteger32= 32 &uinteger64=+%20+&boolean_1=&boolean_2=asdf&fl32_1=asdf&fl32_2=&fl64_1=&fl64_2=asdfstr`,
	}

	expectedOutputs := map[string]Everything{
		"successful": Everything{
			Integer:    -1,
			Integer8:   -8,
			Integer16:  -16,
			Integer32:  -32,
			Integer64:  -64,
			Uinteger:   1,
			Uinteger8:  8,
			Uinteger16: 16,
			Uinteger32: 32,
			Uinteger64: 64,
			Boolean_1:  true,
			Fl32_1:     32.3232,
			Fl64_1:     -64.6464646464,
			Str:        "string",
		},
		"errorful": Everything{},
	}

	for key, testCase := range testInputs {
		httpRecorder := httptest.NewRecorder()
		m := martini.Classic()

		m.Post(testRoute, Form(Everything{}), func(actual Everything, errs Errors) {
			actualStr := fmt.Sprintf("%+v", actual)
			expectedStr := fmt.Sprintf("%+v", expectedOutputs[key])
			if actualStr != expectedStr {
				t.Errorf("For '%s' expected\n%s but got\n%s", key, expectedStr, actualStr)
			}
		})
		req, err := http.NewRequest("POST", testRoute, strings.NewReader(testCase))
		if err != nil {
			panic(err)
		}
		req.Header.Set("Content-Type", formContentType)
		m.ServeHTTP(httpRecorder, req)
	}
}

// Each binder middleware should assert that the struct passed in is not
// a pointer (to avoid race conditions)
func TestEnsureNotPointer(t *testing.T) {
	shouldPanic := func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("Should have panicked because argument is a pointer, but did NOT panic")
			}
		}()
		ensureNotPointer(&Post{})
	}

	shouldNotPanic := func() {
		defer func() {
			r := recover()
			if r != nil {
				t.Errorf("Should NOT have panicked because argument is not a pointer, but did panic")
			}
		}()
		ensureNotPointer(Post{})
	}

	shouldPanic()
	shouldNotPanic()
}

// Used in testing setWithProperType; kind of clunky...
type Everything struct {
	Integer    int     `form:"integer"`
	Integer8   int8    `form:"integer8"`
	Integer16  int16   `form:"integer16"`
	Integer32  int32   `form:"integer32"`
	Integer64  int64   `form:"integer64"`
	Uinteger   uint    `form:"uinteger"`
	Uinteger8  uint8   `form:"uinteger8"`
	Uinteger16 uint16  `form:"uinteger16"`
	Uinteger32 uint32  `form:"uinteger32"`
	Uinteger64 uint64  `form:"uinteger64"`
	Boolean_1  bool    `form:"boolean_1"`
	Boolean_2  bool    `form:"boolean_2"`
	Fl32_1     float32 `form:"fl32_1"`
	Fl32_2     float32 `form:"fl32_2"`
	Fl64_1     float64 `form:"fl64_1"`
	Fl64_2     float64 `form:"fl64_2"`
	Str        string  `form:"str"`
}
