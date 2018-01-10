package spinner

import (
	"fmt"
	"os"
	"time"
)

var symbolList = []string{
	" | ",
	" / ",
	" – ",
	" \\ ",	
}

var symbolMap = map[string]string{
	" ⏳ ": "\033[37m",
	" ⌛ ": "\033[38m",
}

type Spinner struct {
	duration time.Duration
	trigger  interface{}
	sentinel interface{}
	closed   bool
	// sigChan: is the pipe that receives the start and stop
	sigChan chan interface{}
	// waitChan waits till the shutdown has fully propagated
	waitChan chan interface{}
}

func New(freq int64) *Spinner {
	if freq < 1 {
		freq = 10
	}
	sp := Spinner{
		duration: time.Duration(1e9 / freq),
		sentinel: nil,
		// sigChan will be created on .Start()
		sigChan:  nil,
		waitChan: make(chan interface{}),
	}

	sp.trigger = &sp
	return &sp
}

func (s *Spinner) Start() error {
	err := s.spin()
	if err == nil {
		s.sigChan <- s.trigger
	}
	return err
}

func (s *Spinner) Stop() {
	if !s.closed && s.sigChan != nil {
		s.sigChan <- s.sentinel
		close(s.sigChan)
		<-s.waitChan
		s.closed = true
	}
}

func (s *Spinner) Reset() {
	s.Stop()
	s.sigChan = nil
	s.closed = true
}

func (s *Spinner) Duration() time.Duration {
	return s.duration
}

func (s *Spinner) spin() error {
	if s.sigChan != nil { // Already in use
		return fmt.Errorf("already in use")
	}
	s.sigChan = make(chan interface{})
	go func() {
		// Block till the first symbol comes through
		<-s.sigChan

		throttle := time.Tick(s.duration)
		running := true
		for running {
			for _, segment := range symbolList {
				select {
				case in := <-s.sigChan:
					if in == s.sentinel {
						os.Stderr.Sync()
						running = false
						s.waitChan <- s.sentinel
						break
					}
				default:
				}
				// Print it to stderr to avoid symbol getting into piped content
				fmt.Fprintf(os.Stderr, "%s\r", segment)
				<-throttle
			}
		}
	}()
	return nil
}
