package binding

import (
	"fmt"
	"testing"
)

func TestErrorsAdd(t *testing.T) {
	var actual Errors
	expected := Errors{
		Error{
			FieldNames:     []string{"Field1", "Field2"},
			Classification: "ErrorClass",
			Message:        "Some message",
		},
	}

	actual.Add(expected[0].FieldNames, expected[0].Classification, expected[0].Message)

	if len(actual) != 1 {
		t.Errorf("Expected 1 error, but actually had %d", len(actual))
		return
	}

	expectedStr := fmt.Sprintf("%#v", expected)
	actualStr := fmt.Sprintf("%#v", actual)

	if actualStr != expectedStr {
		t.Errorf("Expected:\n%s\nbut got:\n%s", expectedStr, actualStr)
	}
}

func TestErrorsLen(t *testing.T) {
	actual := errorsTestSet.Len()
	expected := len(errorsTestSet)
	if actual != expected {
		t.Errorf("Expected %d, but got %d", expected, actual)
		return
	}
}

func TestErrorsHas(t *testing.T) {
	if errorsTestSet.Has("ClassA") != true {
		t.Errorf("Expected to have error of kind ClassA, but didn't")
	}
	if errorsTestSet.Has("ClassQ") != false {
		t.Errorf("Expected to NOT have error of kind ClassQ, but did")
	}
}

func TestErrorGetters(t *testing.T) {
	err := Error{
		FieldNames:     []string{"field1", "field2"},
		Classification: "ErrorClass",
		Message:        "The message",
	}

	fieldsActual := err.Fields()

	if len(fieldsActual) != 2 {
		t.Errorf("Expected Fields() to return 2 errors, but got %d", len(fieldsActual))
	} else {
		if fieldsActual[0] != "field1" || fieldsActual[1] != "field2" {
			t.Errorf("Expected Fields() to return the correct fields, but it didn't")
		}
	}

	if err.Kind() != "ErrorClass" {
		t.Errorf("Expected the classification to be 'ErrorClass', but got '%s'", err.Kind())
	}

	if err.Error() != "The message" {
		t.Errorf("Expected the message to be 'The message', but got '%s'", err.Error())
	}
}

/*
func TestErrorsWithClass(t *testing.T) {
	expected := Errors{
		errorsTestSet[0],
		errorsTestSet[3],
	}
	actualStr := fmt.Sprintf("%#v", errorsTestSet.WithClass("ClassA"))
	expectedStr := fmt.Sprintf("%#v", expected)
	if actualStr != expectedStr {
		t.Errorf("Expected:\n%s\nbut got:\n%s", expectedStr, actualStr)
	}
}
*/

var errorsTestSet = Errors{
	Error{
		FieldNames:     []string{},
		Classification: "ClassA",
		Message:        "Foobar",
	},
	Error{
		FieldNames:     []string{},
		Classification: "ClassB",
		Message:        "Foo",
	},
	Error{
		FieldNames:     []string{"field1", "field2"},
		Classification: "ClassB",
		Message:        "Foobar",
	},
	Error{
		FieldNames:     []string{"field2"},
		Classification: "ClassA",
		Message:        "Foobar",
	},
	Error{
		FieldNames:     []string{"field2"},
		Classification: "ClassB",
		Message:        "Foobar",
	},
}
