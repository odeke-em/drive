package binding

import "testing"

func TestBind(t *testing.T) {
	for _, testCase := range formTestCases {
		performFormTest(t, Bind, testCase)
	}
	for _, testCase := range jsonTestCases {
		performJsonTest(t, Bind, testCase)
	}
	for _, testCase := range multipartFormTestCases {
		performMultipartFormTest(t, Bind, testCase)
	}
	for _, testCase := range fileTestCases {
		performFileTest(t, Bind, testCase)
	}
}
