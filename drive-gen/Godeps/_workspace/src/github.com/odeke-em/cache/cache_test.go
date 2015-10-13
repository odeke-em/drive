package expirable

import (
	"fmt"
	"path/filepath"
	"strings"

	"testing"
	"time"
)

type expirableValue struct {
	value     interface{}
	entryTime time.Time
}

func (e *expirableValue) Value() interface{} {
	if e == nil {
		return nil
	}
	return e.value
}

func (e *expirableValue) Expired(q time.Time) bool {
	if e == nil {
		return true
	}

	return e.entryTime.Before(q)
}

var newExpirableValue = func(v interface{}) *expirableValue {
	return newExpirableValueWithOffset(v, 0)
}

func newExpirableValueWithOffset(v interface{}, expiry int) *expirableValue {
	return &expirableValue{
		value:     v,
		entryTime: time.Now().Add(time.Duration(expiry)),
	}
}

func parDir(p string) string {
	return filepath.Dir(p)
}

func TestInit(t *testing.T) {
	fmt.Print("testInit")
	store := New()

	manifestStr := `
        /openSrc/node-tika
        /openSrc/Radicale
        /openSrc/rodats
        /openSrc/ical.js
        /openSrc/aws-sdk-js
        /openSrc/dpxdt
        /openSrc/emmodeke-vim-clone
        /openSrc/ytrans.js
        /openSrc/oxy
        /openSrc/drive-js
        /openSrc/git
        /openSrc/ldappy/
        `

	manifest := strings.Split(manifestStr, "\n")

	expiryMs := 25

	tick := time.Tick(time.Duration(expiryMs))

	manifestCount := len(manifest)

	done := make(chan bool, manifestCount)

	for _, item := range manifest {
		go func(p string) {
			par := parDir(p)
			exp := newExpirableValueWithOffset(p, expiryMs)

			store.Put(par, exp)
			retr, ok := store.Get(par)

			if retr != exp {
				fmt.Printf("%s encountered a clashing concurrent access %v %v\n", p, retr, exp)
			}

			fmt.Printf("%s still Fresh? %v\n", par, ok)
			done <- true
		}(item)
	}

	for i := 0; i < manifestCount; i++ {
		<-done
	}

	// fmt.Println(store)
	<-tick

	// Now expecting stale changes

	for _, item := range manifest {
		go func(p string) {
			par := parDir(p)
			retr, ok := store.Get(par)

			if ok {
				t.Errorf("%s should have expired", par)
			}

			fmt.Printf("%s Retr: %v still Fresh? %v\n", par, retr)
		}(item)
	}
}
