package log

import (
	"fmt"
	"io"
	"os"
)

type Loggerf func(string, ...interface{}) (int, error)
type Loggerln func(...interface{}) (int, error)

type logy struct {
	print   Loggerln
	printf  Loggerf
	println Loggerln
}

type logyIn struct {
	scan   Loggerln
	scanf  Loggerf
	scanln Loggerln
}

type Logger struct {
	Logf     Loggerf
	Logln    Loggerln
	Log      Loggerln
	LogErr   Loggerln
	LogErrf  Loggerf
	LogErrln Loggerln
	Scanf    Loggerf
	Scan     Loggerln
	Scanln   Loggerln
}

func noopFmter(format string, args ...interface{}) (int, error) {
	return 0, nil
}

func nooper(args ...interface{}) (int, error) {
	return 0, nil
}

func newLoggerIn(fIn io.Reader) *logyIn {
	if fIn == nil {
		fIn = os.Stdin
	}

	finf := func(format string, args ...interface{}) (int, error) {
		return fmt.Fscanf(fIn, format, args...)
	}
	finln := func(args ...interface{}) (int, error) {
		return fmt.Fscanln(fIn, args...)
	}
	fin := func(args ...interface{}) (int, error) {
		return fmt.Fscan(fIn, args...)
	}

	return &logyIn{
		scan:   fin,
		scanf:  finf,
		scanln: finln,
	}
}

func newLoggerOut(f io.Writer) *logy {
	ff := nooper
	fl := noopFmter
	fln := nooper

	if f != nil {
		fl = func(format string, args ...interface{}) (int, error) {
			return fmt.Fprintf(f, format, args...)
		}

		fln = func(args ...interface{}) (int, error) {
			return fmt.Fprintln(f, args...)
		}

		ff = func(args ...interface{}) (int, error) {
			return fmt.Fprint(f, args...)
		}
	}

	return &logy{
		print:   ff,
		printf:  fl,
		println: fln,
	}
}

func New(stdin io.Reader, writers ...io.Writer) *Logger {
	var stdout, stderr io.Writer

	wLen := len(writers)
	if wLen >= 1 {
		stdout = writers[0]
	}
	if wLen >= 2 {
		stderr = writers[1]
	}

	stdouter := newLoggerOut(stdout)
	stderrer := newLoggerOut(stderr)
	stdiner := newLoggerIn(stdin)

	return &Logger{
		Logf:     stdouter.printf,
		Log:      stdouter.print,
		Logln:    stdouter.println,
		LogErr:   stderrer.print,
		LogErrf:  stderrer.printf,
		LogErrln: stderrer.println,
		Scanf:    stdiner.scanf,
		Scanln:   stdiner.scanln,
	}
}
