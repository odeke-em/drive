package statos

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"testing"
)

func producer(ws *WriteCloserStatos) chan bool {
	done := make(chan bool)

	go func() {
		ticker := time.Tick(1e9)
		for i := 0; i < 20; i += 1 {
			sReader := strings.NewReader("blooming here")
			content := make([]byte, sReader.Len())
			n, err := sReader.Read(content)
			n, err = ws.Write(content)
			// Throttle
			<-ticker
			if n < 1 || err != nil {
				fmt.Printf("while writing bytes: encountered n: %v err: %v\n", n, err)
			}
		}

		ws.Close()
		done <- true
	}()

	return done
}

func wProgresser(ws *WriteCloserStatos, end chan bool) chan bool {
	done := make(chan bool)

	go func() {
		commChan := ws.ProgressChan()
		for n := range commChan {
			fmt.Printf("%v\r", n)
		}
		done <- true
	}()

	return done
}

func TestWriteCloser(t *testing.T) {
	destName := strings.Join([]string{
		".",
		fmt.Sprintf("destv%v.dest", rand.Int()),
	}, "x")

	destAbsPath, fullPErr := filepath.Abs(filepath.Join(".", destName))
	if fullPErr != nil {
		t.Errorf("%v", fullPErr)
		return
	}

	destFile, err := os.Create(destAbsPath)
	if err != nil {
		t.Errorf("%s: %v\n", destName, err)
		return
	}

	defer func() {
		if rmErr := os.RemoveAll(destAbsPath); rmErr != nil {
			fmt.Fprintf(os.Stderr, "Unlink: %s \033[91m: %v\033[00m\n", destName, rmErr)
		}
	}()

	ws := NewWriteCloser(destFile)

	producerChan := producer(ws)
	done := wProgresser(ws, producerChan)

	<-done
}
