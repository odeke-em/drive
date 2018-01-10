package main

import (
	"fmt"
	"os"

	"github.com/odeke-em/go-utils/fread"
)

func main() {
	linesChan := fread.Fread(os.Stdin)

	for line := range linesChan {
		fmt.Println(line)
	}
}
