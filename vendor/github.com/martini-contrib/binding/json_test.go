package binding

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-martini/martini"
)

var jsonTestCases = []jsonTestCase{
	{
		description:         "Happy path",
		shouldSucceedOnJson: true,
		payload:             `{"title": "Glorious Post Title", "content": "Lorem ipsum dolor sit amet"}`,
		contentType:         jsonContentType,
		expected:            Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:         "Happy path with interface",
		shouldSucceedOnJson: true,
		withInterface:       true,
		payload:             `{"title": "Glorious Post Title", "content": "Lorem ipsum dolor sit amet"}`,
		contentType:         jsonContentType,
		expected:            Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:         "Nil payload",
		shouldSucceedOnJson: false,
		payload:             `-nil-`,
		contentType:         jsonContentType,
		expected:            Post{},
	},
	{
		description:         "Empty payload",
		shouldSucceedOnJson: false,
		payload:             ``,
		contentType:         jsonContentType,
		expected:            Post{},
	},
	{
		description:         "Empty content type",
		shouldSucceedOnJson: true,
		shouldFailOnBind:    true,
		payload:             `{"title": "Glorious Post Title", "content": "Lorem ipsum dolor sit amet"}`,
		contentType:         ``,
		expected:            Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:         "Unsupported content type",
		shouldSucceedOnJson: true,
		shouldFailOnBind:    true,
		payload:             `{"title": "Glorious Post Title", "content": "Lorem ipsum dolor sit amet"}`,
		contentType:         `BoGuS`,
		expected:            Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:         "Malformed JSON",
		shouldSucceedOnJson: false,
		payload:             `{"title":"foo"`,
		contentType:         jsonContentType,
		expected:            Post{},
	},
	{
		description:         "Deserialization with nested and embedded struct",
		shouldSucceedOnJson: true,
		payload:             `{"title":"Glorious Post Title", "id":1, "author":{"name":"Matt Holt"}}`,
		contentType:         jsonContentType,
		expected:            BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:         "Deserialization with nested and embedded struct with interface",
		shouldSucceedOnJson: true,
		withInterface:       true,
		payload:             `{"title":"Glorious Post Title", "id":1, "author":{"name":"Matt Holt"}}`,
		contentType:         jsonContentType,
		expected:            BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:         "Required nested struct field not specified",
		shouldSucceedOnJson: false,
		payload:             `{"title":"Glorious Post Title", "id":1, "author":{}}`,
		contentType:         jsonContentType,
		expected:            BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1},
	},
	{
		description:         "Required embedded struct field not specified",
		shouldSucceedOnJson: false,
		payload:             `{"id":1, "author":{"name":"Matt Holt"}}`,
		contentType:         jsonContentType,
		expected:            BlogPost{Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:         "Slice of Posts",
		shouldSucceedOnJson: true,
		payload:             `[{"title": "First Post"}, {"title": "Second Post"}]`,
		contentType:         jsonContentType,
		expected:            []Post{Post{Title: "First Post"}, Post{Title: "Second Post"}},
	},
}

func TestJson(t *testing.T) {
	for _, testCase := range jsonTestCases {
		performJsonTest(t, Json, testCase)
	}
}

func performJsonTest(t *testing.T, binder handlerFunc, testCase jsonTestCase) {
	var payload io.Reader
	httpRecorder := httptest.NewRecorder()
	m := martini.Classic()

	jsonTestHandler := func(actual interface{}, errs Errors) {
		if testCase.shouldSucceedOnJson && len(errs) > 0 {
			t.Errorf("'%s' should have succeeded, but there were errors (%d):\n%+v",
				testCase.description, len(errs), errs)
		} else if !testCase.shouldSucceedOnJson && len(errs) == 0 {
			t.Errorf("'%s' should NOT have succeeded, but there were NO errors", testCase.description)
		}
		expString := fmt.Sprintf("%+v", testCase.expected)
		actString := fmt.Sprintf("%+v", actual)
		if actString != expString {
			t.Errorf("'%s': expected\n'%s'\nbut got\n'%s'",
				testCase.description, expString, actString)
		}
	}

	switch testCase.expected.(type) {
	case []Post:
		if testCase.withInterface {
			m.Post(testRoute, binder([]Post{}, (*modeler)(nil)), func(actual []Post, iface modeler, errs Errors) {

				for _, a := range actual {
					if a.Title != iface.Model() {
						t.Errorf("For '%s': expected the struct to be mapped to the context as an interface",
							testCase.description)
					}
					jsonTestHandler(a, errs)
				}
			})
		} else {
			m.Post(testRoute, binder([]Post{}), func(actual []Post, errs Errors) {
				jsonTestHandler(actual, errs)
			})
		}

	case Post:
		if testCase.withInterface {
			m.Post(testRoute, binder(Post{}, (*modeler)(nil)), func(actual Post, iface modeler, errs Errors) {
				if actual.Title != iface.Model() {
					t.Errorf("For '%s': expected the struct to be mapped to the context as an interface",
						testCase.description)
				}
				jsonTestHandler(actual, errs)
			})
		} else {
			m.Post(testRoute, binder(Post{}), func(actual Post, errs Errors) {
				jsonTestHandler(actual, errs)
			})
		}

	case BlogPost:
		if testCase.withInterface {
			m.Post(testRoute, binder(BlogPost{}, (*modeler)(nil)), func(actual BlogPost, iface modeler, errs Errors) {
				if actual.Title != iface.Model() {
					t.Errorf("For '%s': expected the struct to be mapped to the context as an interface",
						testCase.description)
				}
				jsonTestHandler(actual, errs)
			})
		} else {
			m.Post(testRoute, binder(BlogPost{}), func(actual BlogPost, errs Errors) {
				jsonTestHandler(actual, errs)
			})
		}
	}

	if testCase.payload == "-nil-" {
		payload = nil
	} else {
		payload = strings.NewReader(testCase.payload)
	}

	req, err := http.NewRequest("POST", testRoute, payload)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", testCase.contentType)

	m.ServeHTTP(httpRecorder, req)

	switch httpRecorder.Code {
	case http.StatusNotFound:
		panic("Routing is messed up in test fixture (got 404): check method and path")
	case http.StatusInternalServerError:
		panic("Something bad happened on '" + testCase.description + "'")
	default:
		if testCase.shouldSucceedOnJson &&
			httpRecorder.Code != http.StatusOK &&
			!testCase.shouldFailOnBind {
			t.Errorf("'%s' should have succeeded (except when using Bind, where it should fail), but returned HTTP status %d with body '%s'",
				testCase.description, httpRecorder.Code, httpRecorder.Body.String())
		}
	}
}

type (
	jsonTestCase struct {
		description         string
		withInterface       bool
		shouldSucceedOnJson bool
		shouldFailOnBind    bool
		payload             string
		contentType         string
		expected            interface{}
	}
)
