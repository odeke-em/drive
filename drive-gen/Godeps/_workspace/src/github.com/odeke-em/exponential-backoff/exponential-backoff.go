package expb

import (
	"fmt"
	"math/rand"
	"time"
)

type StatusChecker func(q interface{}) (ok, retryable bool)

type ExponentialBacker struct {
	Debug       bool
	Do          Producer
	RetryCount  uint32
	StatusCheck StatusChecker
}

type pair struct {
	first interface{}
	last  interface{}
}

type Callback func(result interface{}, err error)
type Producer func() (result interface{}, err error)

func ExponentialBackOffSync(bk *ExponentialBacker) (interface{}, error) {

	done := make(chan *pair)

	go func() {
		defer close(done)
		retries := uint32(0)
		for {
			res, err := bk.Do()
			ok, retryable := bk.StatusCheck(res)

			if ok || !retryable || retries >= bk.RetryCount {
				done <- &pair{first: res, last: err}
				break
			}

			ms := time.Duration(1e9*rand.Float64()) + ((1 << retries) * time.Second)
			if bk.Debug {
				fmt.Printf("trying again in %v\n", ms)
			}

			duration := time.Duration(ms)
			time.Sleep(duration)

			retries += 1
		}
	}()

	v := <-done

	res := v.first
	err, _ := v.last.(error)

	return res, err
}

func ExponentialBackOff(bk *ExponentialBacker, cb Callback) {
	res, err := ExponentialBackOffSync(bk)
	cb(res, err)
}
