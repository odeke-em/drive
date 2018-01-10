package binding

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-martini/martini"
)

var validationTestCases = []validationTestCase{
	{
		description: "No errors",
		data: BlogPost{
			Id: 1,
			Post: Post{
				Title:   "Behold The Title!",
				Content: "And some content",
			},
			Author: Person{
				Name: "Matt Holt",
			},
		},
		expectedErrors: Errors{},
	},
	{
		description: "ID required",
		data: BlogPost{
			Post: Post{
				Title:   "Behold The Title!",
				Content: "And some content",
			},
			Author: Person{
				Name: "Matt Holt",
			},
		},
		expectedErrors: Errors{
			Error{
				FieldNames:     []string{"id"},
				Classification: RequiredError,
				Message:        "Required",
			},
		},
	},
	{
		description: "Embedded struct field required",
		data: BlogPost{
			Id: 1,
			Post: Post{
				Content: "Content given, but title is required",
			},
			Author: Person{
				Name: "Matt Holt",
			},
		},
		expectedErrors: Errors{
			Error{
				FieldNames:     []string{"title"},
				Classification: RequiredError,
				Message:        "Required",
			},
			Error{
				FieldNames:     []string{"title"},
				Classification: "LengthError",
				Message:        "Life is too short",
			},
		},
	},
	{
		description: "Nested struct field required",
		data: BlogPost{
			Id: 1,
			Post: Post{
				Title:   "Behold The Title!",
				Content: "And some content",
			},
		},
		expectedErrors: Errors{
			Error{
				FieldNames:     []string{"name"},
				Classification: RequiredError,
				Message:        "Required",
			},
		},
	},
	{
		description: "Required field missing in nested struct pointer",
		data: BlogPost{
			Id: 1,
			Post: Post{
				Title:   "Behold The Title!",
				Content: "And some content",
			},
			Author: Person{
				Name: "Matt Holt",
			},
			Coauthor: &Person{},
		},
		expectedErrors: Errors{
			Error{
				FieldNames:     []string{"name"},
				Classification: RequiredError,
				Message:        "Required",
			},
		},
	},
	{
		description: "All required fields specified in nested struct pointer",
		data: BlogPost{
			Id: 1,
			Post: Post{
				Title:   "Behold The Title!",
				Content: "And some content",
			},
			Author: Person{
				Name: "Matt Holt",
			},
			Coauthor: &Person{
				Name: "Jeremy Saenz",
			},
		},
		expectedErrors: Errors{},
	},
	{
		description: "Custom validation should put an error",
		data: BlogPost{
			Id: 1,
			Post: Post{
				Title:   "Too short",
				Content: "And some content",
			},
			Author: Person{
				Name: "Matt Holt",
			},
		},
		expectedErrors: Errors{
			Error{
				FieldNames:     []string{"title"},
				Classification: "LengthError",
				Message:        "Life is too short",
			},
		},
	},
	{
		description: "List Validation",
		data: []BlogPost{
			BlogPost{
				Id: 1,
				Post: Post{
					Title:   "First Post",
					Content: "And some content",
				},
				Author: Person{
					Name: "Leeor Aharon",
				},
			},
			BlogPost{
				Id: 2,
				Post: Post{
					Title:   "Second Post",
					Content: "And some content",
				},
				Author: Person{
					Name: "Leeor Aharon",
				},
			},
		},
		expectedErrors: Errors{},
	},
	{
		description: "List Validation w/ Errors",
		data: []BlogPost{
			BlogPost{
				Id: 1,
				Post: Post{
					Title:   "First Post",
					Content: "And some content",
				},
				Author: Person{
					Name: "Leeor Aharon",
				},
			},
			BlogPost{
				Id: 2,
				Post: Post{
					Title:   "Too Short",
					Content: "And some content",
				},
				Author: Person{
					Name: "Leeor Aharon",
				},
			},
		},
		expectedErrors: Errors{
			Error{
				FieldNames:     []string{"title"},
				Classification: "LengthError",
				Message:        "Life is too short",
			},
		},
	},
}

func TestValidation(t *testing.T) {
	for _, testCase := range validationTestCases {
		performValidationTest(t, testCase)
	}
}

func performValidationTest(t *testing.T, testCase validationTestCase) {
	httpRecorder := httptest.NewRecorder()
	m := martini.Classic()

	m.Post(testRoute, Validate(testCase.data), func(actual Errors) {
		expString := fmt.Sprintf("%+v", testCase.expectedErrors)
		actString := fmt.Sprintf("%+v", actual)
		if actString != expString {
			t.Errorf("For '%s': expected errors to be\n'%s'\nbut got\n'%s'",
				testCase.description, expString, actString)
		}
	})

	req, err := http.NewRequest("POST", testRoute, nil)
	if err != nil {
		panic(err)
	}

	m.ServeHTTP(httpRecorder, req)

	switch httpRecorder.Code {
	case http.StatusNotFound:
		panic("Routing is messed up in test fixture (got 404): check methods and paths")
	case http.StatusInternalServerError:
		panic("Something bad happened on '" + testCase.description + "'")
	}
}

type (
	validationTestCase struct {
		description    string
		data           interface{}
		expectedErrors Errors
	}
)
