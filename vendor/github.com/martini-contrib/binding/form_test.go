package binding

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/go-martini/martini"
)

var addressableTrue bool = true

var formTestCases = []formTestCase{
	{
		description:   "Happy path",
		shouldSucceed: true,
		payload:       `title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet`,
		contentType:   formContentType,
		expected:      Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:   "Happy path with interface",
		shouldSucceed: true,
		withInterface: true,
		payload:       `title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet`,
		contentType:   formContentType,
		expected:      Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:   "Empty payload",
		shouldSucceed: false,
		payload:       ``,
		contentType:   formContentType,
		expected:      Post{},
	},
	{
		description:   "Empty content type",
		shouldSucceed: false,
		payload:       `title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet`,
		contentType:   ``,
		expected:      Post{},
	},
	{
		description:   "Malformed form body",
		shouldSucceed: false,
		payload:       `title=%2`,
		contentType:   formContentType,
		expected:      Post{},
	},
	{
		description:   "With nested and embedded structs",
		shouldSucceed: true,
		payload:       `title=Glorious+Post+Title&id=1&name=Matt+Holt`,
		contentType:   formContentType,
		expected:      BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:   "Required embedded struct field not specified",
		shouldSucceed: false,
		payload:       `id=1&name=Matt+Holt`,
		contentType:   formContentType,
		expected:      BlogPost{Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:   "Required nested struct field not specified",
		shouldSucceed: false,
		payload:       `title=Glorious+Post+Title&id=1`,
		contentType:   formContentType,
		expected:      BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1},
	},
	{
		description:   "Multiple values into slice",
		shouldSucceed: true,
		payload:       `title=Glorious+Post+Title&id=1&name=Matt+Holt&rating=4&rating=3&rating=5`,
		contentType:   formContentType,
		expected:      BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}, Ratings: []int{4, 3, 5}},
	},
	{
		description:   "Unexported field",
		shouldSucceed: true,
		payload:       `title=Glorious+Post+Title&id=1&name=Matt+Holt&unexported=foo`,
		contentType:   formContentType,
		expected:      BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:   "Bool pointer",
		shouldSucceed: true,
		payload:       `title=Glorious+Post+Title&id=1&name=Matt+Holt&is_active=true`,
		contentType:   formContentType,
		expected:      BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}, IsActive: &addressableTrue},
		deepEqual:     true,
	},
	{
		description:   "Query string POST",
		shouldSucceed: true,
		payload:       `title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet`,
		contentType:   formContentType,
		expected:      Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:   "Query string with Content-Type (POST request)",
		shouldSucceed: true,
		queryString:   "?title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet",
		payload:       ``,
		contentType:   formContentType,
		expected:      Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:   "Query string without Content-Type (GET request)",
		shouldSucceed: true,
		method:        "GET",
		queryString:   "?title=Glorious+Post+Title&content=Lorem+ipsum+dolor+sit+amet",
		payload:       ``,
		expected:      Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:   "Embed struct pointer",
		shouldSucceed: true,
		deepEqual:     true,
		method:        "GET",
		queryString:   "?name=Glorious+Post+Title&email=Lorem+ipsum+dolor+sit+amet",
		payload:       ``,
		expected:      EmbedPerson{&Person{Name: "Glorious Post Title", Email: "Lorem ipsum dolor sit amet"}},
	},
	{
		description:   "Embed struct pointer remain nil if not binded",
		shouldSucceed: true,
		deepEqual:     true,
		method:        "GET",
		queryString:   "?",
		payload:       ``,
		expected:      EmbedPerson{nil},
	},
}

func TestForm(t *testing.T) {
	for _, testCase := range formTestCases {
		performFormTest(t, Form, testCase)
	}
}

func performFormTest(t *testing.T, binder handlerFunc, testCase formTestCase) {
	httpRecorder := httptest.NewRecorder()
	m := martini.Classic()

	formTestHandler := func(actual interface{}, errs Errors) {
		if testCase.shouldSucceed && len(errs) > 0 {
			t.Errorf("'%s' should have succeeded, but there were errors (%d):\n%+v",
				testCase.description, len(errs), errs)
		} else if !testCase.shouldSucceed && len(errs) == 0 {
			t.Errorf("'%s' should have had errors, but there were none", testCase.description)
		}
		expString := fmt.Sprintf("%+v", testCase.expected)
		actString := fmt.Sprintf("%+v", actual)
		if actString != expString && !(testCase.deepEqual && reflect.DeepEqual(testCase.expected, actual)) {
			t.Errorf("'%s': expected\n'%s'\nbut got\n'%s'",
				testCase.description, expString, actString)
		}
	}

	switch testCase.expected.(type) {
	case Post:
		if testCase.withInterface {
			m.Post(testRoute, binder(Post{}, (*modeler)(nil)), func(actual Post, iface modeler, errs Errors) {
				if actual.Title != iface.Model() {
					t.Errorf("For '%s': expected the struct to be mapped to the context as an interface",
						testCase.description)
				}
				formTestHandler(actual, errs)
			})
		} else {
			m.Post(testRoute, binder(Post{}), func(actual Post, errs Errors) {
				formTestHandler(actual, errs)
			})
			m.Get(testRoute, binder(Post{}), func(actual Post, errs Errors) {
				formTestHandler(actual, errs)
			})
		}

	case BlogPost:
		if testCase.withInterface {
			m.Post(testRoute, binder(BlogPost{}, (*modeler)(nil)), func(actual BlogPost, iface modeler, errs Errors) {
				if actual.Title != iface.Model() {
					t.Errorf("For '%s': expected the struct to be mapped to the context as an interface",
						testCase.description)
				}
				formTestHandler(actual, errs)
			})
		} else {
			m.Post(testRoute, binder(BlogPost{}), func(actual BlogPost, errs Errors) {
				formTestHandler(actual, errs)
			})
		}

	case EmbedPerson:
		m.Post(testRoute, binder(EmbedPerson{}), func(actual EmbedPerson, errs Errors) {
			formTestHandler(actual, errs)
		})
		m.Get(testRoute, binder(EmbedPerson{}), func(actual EmbedPerson, errs Errors) {
			formTestHandler(actual, errs)
		})
	}

	if testCase.method == "" {
		testCase.method = "POST"
	}

	req, err := http.NewRequest(testCase.method, testRoute+testCase.queryString, strings.NewReader(testCase.payload))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", testCase.contentType)

	m.ServeHTTP(httpRecorder, req)

	switch httpRecorder.Code {
	case http.StatusNotFound:
		panic("Routing is messed up in test fixture (got 404): check methods and paths")
	case http.StatusInternalServerError:
		panic("Something bad happened on '" + testCase.description + "'")
	}
}

type (
	formTestCase struct {
		description   string
		shouldSucceed bool
		deepEqual     bool
		withInterface bool
		queryString   string
		payload       string
		contentType   string
		expected      interface{}
		method        string
	}
)
