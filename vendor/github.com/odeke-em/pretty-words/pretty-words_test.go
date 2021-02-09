package prettywords

import (
	"strings"
	"testing"
)

func TestBlankPassage(t *testing.T) {
	pr := PrettyRubric{
		Body:  []string{},
		Limit: 0,
	}

	formatted := pr.Format()
	if len(formatted) != 0 {
		t.Errorf("expected zero-length formatted text for Limit=0, got %v", formatted)
	}

	pr.Limit = 80
	formatted = pr.Format()
	if len(formatted) != 0 {
		t.Errorf("expected zero-length formatted text for Limit=0, got %v", formatted)
	}

	pr.Body = []string{"This is a test right here"}
	pr.Limit = -1
	if len(formatted) != 0 {
		t.Errorf("expected zero-length formatted text for Limit=0, got %v", formatted)
	}
}

func TestFormatting(t *testing.T) {
	pr := PrettyRubric{
		Body:  strings.Split("This right here is a real test right here right now right here", " "),
		Limit: 10,
	}

	formatted := pr.Format()
	joined := strings.Join(formatted, "\n")
	expected := "This right\nhere is a\nreal test\nright here\nright now\nright here"

	if expected != joined {
		t.Errorf("expected \n\t>\033[92m%q\033[00m,\ngot \n\t<\033[93m%q\033[00m", expected, joined)
	}
}
