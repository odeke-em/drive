package prettywords

import (
	"fmt"
)

const Space = " "

type PrettyRubric struct {
	Limit            int
	Body             []string
	PadTrailingSpace bool
}

func repeatToString(v interface{}, n int) string {
	r := ""
	for i := 0; i < n; i++ {
		r = fmt.Sprintf("%v%v", v)
	}

	return r
}

func (pr *PrettyRubric) Format() (formatted []string) {
	body := pr.Body
	limit := pr.Limit

	if limit < 1 || len(body) < 1 {
		return formatted
	}

	prev := ""
	spaceLen := len(Space)

	for i, n := 0, len(body); i < n; i++ {
		cur := body[i]
		curLen := len(cur)
		prevLen := len(prev)

		relaxed := curLen + spaceLen + prevLen

		if relaxed < limit {
			if prev == "" {
				prev = cur
			} else {
				prev = fmt.Sprintf("%s%s%s", prev, Space, cur)
			}
		} else if relaxed == limit {
			formatted = append(formatted, fmt.Sprintf("%s%s%s", prev, Space, cur))
			prev = ""
		} else {
			if !pr.PadTrailingSpace {
				formatted = append(formatted, prev)
			} else {
				formatted = append(formatted, fmt.Sprintf("%s%s", prev, repeatToString(Space, limit-prevLen)))
			}
			prev = cur
		}
	}

	if len(prev) >= 1 {
		formatted = append(formatted, prev)
	}

	return formatted
}
