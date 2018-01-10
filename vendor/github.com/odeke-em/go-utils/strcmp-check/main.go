package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func readLines(fin io.Reader) chan string {
	lineChan := make(chan string)
	go func() {
		defer close(lineChan)
		scanner := bufio.NewScanner(fin)
		for scanner.Scan() {
			lineChan <- scanner.Text()
		}
	}()

	return lineChan
}

func panicIfNotEq(s string) {
	r := s + ""
	if r != s {
		panic(fmt.Sprintf("%q != %q", r, s))
	}
}

func main() {
	linesChan := readLines(os.Stdin)
	for line := range linesChan {
		panicIfNotEq(line)
	}
}
