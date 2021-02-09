package binding

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var errorTestCases = []errorTestCase{
	{
		description: "No errors",
		errors:      Errors{},
		expected: errorTestResult{
			statusCode: http.StatusOK,
		},
	},
	{
		description: "Deserialization error",
		errors: Errors{
			{
				Classification: DeserializationError,
				Message:        "Some parser error here",
			},
		},
		expected: errorTestResult{
			statusCode:  http.StatusBadRequest,
			contentType: jsonContentType,
			body:        `[{"classification":"DeserializationError","message":"Some parser error here"}]`,
		},
	},
	{
		description: "Content-Type error",
		errors: Errors{
			{
				Classification: ContentTypeError,
				Message:        "Empty Content-Type",
			},
		},
		expected: errorTestResult{
			statusCode:  http.StatusUnsupportedMediaType,
			contentType: jsonContentType,
			body:        `[{"classification":"ContentTypeError","message":"Empty Content-Type"}]`,
		},
	},
	{
		description: "Requirement error",
		errors: Errors{
			{
				FieldNames:     []string{"some_field"},
				Classification: RequiredError,
				Message:        "Required",
			},
		},
		expected: errorTestResult{
			statusCode:  StatusUnprocessableEntity,
			contentType: jsonContentType,
			body:        `[{"fieldNames":["some_field"],"classification":"RequiredError","message":"Required"}]`,
		},
	},
	{
		description: "Bad header error",
		errors: Errors{
			{
				Classification: "HeaderError",
				Message:        "The X-Something header must be specified",
			},
		},
		expected: errorTestResult{
			statusCode:  StatusUnprocessableEntity,
			contentType: jsonContentType,
			body:        `[{"classification":"HeaderError","message":"The X-Something header must be specified"}]`,
		},
	},
	{
		description: "Custom field error",
		errors: Errors{
			{
				FieldNames:     []string{"month", "year"},
				Classification: "DateError",
				Message:        "The month and year must be in the future",
			},
		},
		expected: errorTestResult{
			statusCode:  StatusUnprocessableEntity,
			contentType: jsonContentType,
			body:        `[{"fieldNames":["month","year"],"classification":"DateError","message":"The month and year must be in the future"}]`,
		},
	},
	{
		description: "Multiple errors",
		errors: Errors{
			{
				FieldNames:     []string{"foo"},
				Classification: RequiredError,
				Message:        "Required",
			},
			{
				FieldNames:     []string{"foo"},
				Classification: "LengthError",
				Message:        "The length of the 'foo' field is too short",
			},
		},
		expected: errorTestResult{
			statusCode:  StatusUnprocessableEntity,
			contentType: jsonContentType,
			body:        `[{"fieldNames":["foo"],"classification":"RequiredError","message":"Required"},{"fieldNames":["foo"],"classification":"LengthError","message":"The length of the 'foo' field is too short"}]`,
		},
	},
}

func TestErrorHandler(t *testing.T) {
	for _, testCase := range errorTestCases {
		performErrorTest(t, testCase)
	}
}

func performErrorTest(t *testing.T, testCase errorTestCase) {
	httpRecorder := httptest.NewRecorder()

	ErrorHandler(testCase.errors, httpRecorder)

	actualBody, _ := ioutil.ReadAll(httpRecorder.Body)
	actualContentType := httpRecorder.Header().Get("Content-Type")

	if httpRecorder.Code != testCase.expected.statusCode {
		t.Errorf("For '%s': expected status code %d but got %d instead",
			testCase.description, testCase.expected.statusCode, httpRecorder.Code)
	}
	if actualContentType != testCase.expected.contentType {
		t.Errorf("For '%s': expected content-type '%s' but got '%s' instead",
			testCase.description, testCase.expected.contentType, actualContentType)
	}
	if string(actualBody) != testCase.expected.body {
		t.Errorf("For '%s': expected body to be '%s' but got '%s' instead",
			testCase.description, testCase.expected.body, actualBody)
	}
}

type (
	errorTestCase struct {
		description string
		errors      Errors
		expected    errorTestResult
	}

	errorTestResult struct {
		statusCode  int
		contentType string
		body        string
	}
)
