package fread

import (
	"bufio"
	"io"
	"strings"
)

func strNoop(s string) bool {
	return false
}

func fReadFile_(f io.Reader, ignorer func(string) bool) (linesChan chan string) {
	linesChan = make(chan string)
	scanner := bufio.NewScanner(f)

	go func() {
		defer close(linesChan)

		for scanner.Scan() {
			line := scanner.Text()
			line = strings.Trim(line, " ")
			line = strings.Trim(line, "\n")
			if ignorer != nil && ignorer(line) {
				continue
			}

			linesChan <- line
		}
	}()

	return linesChan
}

func Fread(f io.Reader) (linesChan chan string) {
	return fReadFile_(f, strNoop)
}

func FreadWithIgnorer(f io.Reader, ignorer func(string) bool) chan string {
	return fReadFile_(f, ignorer)
}
