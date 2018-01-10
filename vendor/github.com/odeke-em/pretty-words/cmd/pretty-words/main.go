package main

import (
	"fmt"
	"os"
	"strings"

	prettywords "github.com/odeke-em/pretty-words"
)

func main() {
	pr := prettywords.PrettyRubric{
		Limit: 80,
	}

	rest := os.Args[1:]
	pr.Body = rest
	formatted := pr.Format()
	fmt.Println(strings.Join(formatted, "\n"))
}
