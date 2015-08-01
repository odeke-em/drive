package statos

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"testing"
)

func currentFile() string {
	_, filename, _, _ := runtime.Caller(1)
	return filename
}

func consumer(rs *ReadCloserStatos) chan bool {
	done := make(chan bool)

	go func() {
		ticker := time.Tick(1e8)
		for {
			bk := make([]byte, 64)
			// Throttle
			n, err := rs.Read(bk)
			<-ticker
			if n < 1 || err != nil {
				break
			}
		}
		done <- true
	}()

	return done
}

func progresser(rs *ReadCloserStatos, end chan bool) chan bool {
	done := make(chan bool)

	go func() {
		commChan := rs.ProgressChan()
		for n := range commChan {
			fmt.Printf("%v\r", n)
		}
		done <- true
	}()

	return done
}

func TestReader(t *testing.T) {
	curFile := currentFile()
	r, err := os.Open(curFile)
	if err != nil {
		fmt.Printf("%s: %v\n", curFile, err)
		return
	}
	rs := NewReadCloser(r)

	consumerChan := consumer(rs)
	done := progresser(rs, consumerChan)

	<-done

	defer rs.Close()
}
