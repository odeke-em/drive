package binding

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-martini/martini"
)

var multipartFormTestCases = []multipartFormTestCase{
	{
		description:      "Happy multipart form path",
		shouldSucceed:    true,
		inputAndExpected: BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:         "FormValue called before req.MultipartReader(); see https://github.com/martini-contrib/csrf/issues/6",
		shouldSucceed:       true,
		callFormValueBefore: true,
		inputAndExpected:    BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:      "Empty payload",
		shouldSucceed:    false,
		inputAndExpected: BlogPost{},
	},
	{
		description:      "Missing required field (Id)",
		shouldSucceed:    false,
		inputAndExpected: BlogPost{Post: Post{Title: "Glorious Post Title"}, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:      "Required embedded struct field not specified",
		shouldSucceed:    false,
		inputAndExpected: BlogPost{Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:      "Required nested struct field not specified",
		shouldSucceed:    false,
		inputAndExpected: BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1},
	},
	{
		description:      "Multiple values",
		shouldSucceed:    true,
		inputAndExpected: BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}, Ratings: []int{3, 5, 4}},
	},
	{
		description:     "Bad multipart encoding",
		shouldSucceed:   false,
		malformEncoding: true,
	},
}

func TestMultipartForm(t *testing.T) {
	for _, testCase := range multipartFormTestCases {
		performMultipartFormTest(t, MultipartForm, testCase)
	}
}

func performMultipartFormTest(t *testing.T, binder handlerFunc, testCase multipartFormTestCase) {
	httpRecorder := httptest.NewRecorder()
	m := martini.Classic()

	m.Post(testRoute, binder(BlogPost{}), func(actual BlogPost, errs Errors) {
		if testCase.shouldSucceed && len(errs) > 0 {
			t.Errorf("'%s' should have succeeded, but there were errors (%d):\n%+v",
				testCase.description, len(errs), errs)
		} else if !testCase.shouldSucceed && len(errs) == 0 {
			t.Errorf("'%s' should not have succeeded, but it did (there were no errors)", testCase.description)
		}
		expString := fmt.Sprintf("%+v", testCase.inputAndExpected)
		actString := fmt.Sprintf("%+v", actual)
		if actString != expString {
			t.Errorf("'%s': expected\n'%s'\nbut got\n'%s'",
				testCase.description, expString, actString)
		}
	})

	multipartPayload, mpWriter := makeMultipartPayload(testCase)

	req, err := http.NewRequest("POST", testRoute, multipartPayload)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Content-Type", mpWriter.FormDataContentType())

	err = mpWriter.Close()
	if err != nil {
		panic(err)
	}

	if testCase.callFormValueBefore {
		req.FormValue("foo")
	}

	m.ServeHTTP(httpRecorder, req)

	switch httpRecorder.Code {
	case http.StatusNotFound:
		panic("Routing is messed up in test fixture (got 404): check methods and paths")
	case http.StatusInternalServerError:
		panic("Something bad happened on '" + testCase.description + "'")
	}
}

// Writes the input from a test case into a buffer using the multipart writer.
func makeMultipartPayload(testCase multipartFormTestCase) (*bytes.Buffer, *multipart.Writer) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if testCase.malformEncoding {
		// TODO: Break the multipart form parser which is apparently impervious!!
		// (Get it to return an error. Trying to get 100% test coverage.)
		body.Write([]byte(`--` + writer.Boundary() + `\nContent-Disposition: form-data; name="foo"\n\n--` + writer.Boundary() + `--`))
		return body, writer
	} else {
		writer.WriteField("title", testCase.inputAndExpected.Title)
		writer.WriteField("content", testCase.inputAndExpected.Content)
		writer.WriteField("id", strconv.Itoa(testCase.inputAndExpected.Id))
		writer.WriteField("ignored", testCase.inputAndExpected.Ignored)
		for _, value := range testCase.inputAndExpected.Ratings {
			writer.WriteField("rating", strconv.Itoa(value))
		}
		writer.WriteField("name", testCase.inputAndExpected.Author.Name)
		writer.WriteField("email", testCase.inputAndExpected.Author.Email)
		writer.WriteField("unexported", testCase.inputAndExpected.unexported)
		return body, writer
	}
}

type (
	multipartFormTestCase struct {
		description         string
		shouldSucceed       bool
		inputAndExpected    BlogPost
		malformEncoding     bool
		callFormValueBefore bool
	}
)
