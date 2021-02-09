// Copyright 2015 Emmanuel Odeke. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package semalim

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

type uintOrStringCreator struct {
	id uint64
}

func (vc uintOrStringCreator) Id() interface{} {
	return vc.id
}

var sentinelError = fmt.Errorf("i am a sample error")

func (vc uintOrStringCreator) Do() (interface{}, error) {
	var value interface{} = vc.id
	var err = sentinelError

	if vc.id&1 == 0 {
		err = nil
		value = fmt.Sprintf("myValue is: %v", vc.id)
	}

	// Sleep here for a few, simulate some work
	sleepNumber := time.Duration(1e4 * rand.Float64())
	<-time.After(sleepNumber)

	return value, err
}

func TestRun(t *testing.T) {
	jobBench := make(chan Job)
	go func() {
		defer close(jobBench)

		n := uint64(5000)

		for i := uint64(0); i < n; i++ {
			jobBench <- uintOrStringCreator{id: i}
		}
	}()

	results := Run(jobBench, 10)

	for result := range results {
		if result == nil {
			t.Fatalf("a nil result was returned")
		}

		idV, value, err := result.Id(), result.Value(), result.Err()

		id, conforms := idV.(uint64)
		if !conforms {
			t.Errorf("expected a uint64 Id, got %v", id)
		}

		if id&1 == 0 {
			if _, ok := value.(string); !ok {
				t.Errorf("since id %v is even, expected a string, got %v", id, value)
			}
			if err != nil {
				t.Errorf("for even values expected nil err")
			}
		} else {
			if _, ok := value.(uint64); !ok {
				t.Errorf("since id %v is even, expected a string, got %v", id, value)
			}
			if err == nil {
				t.Errorf("for even values expected non-nil err")
			}
		}
	}
}
