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

type ackType struct{}

var sentinel = ackType{}

type Job interface {
	Id() interface{}
	Do() (interface{}, error)
}

type Result interface {
	Err() error
	Id() interface{}
	Value() interface{}
}

type resultSt struct {
	err   error
	id    interface{}
	value interface{}
}

func (rs resultSt) Err() error         { return rs.err }
func (rs resultSt) Id() interface{}    { return rs.id }
func (rs resultSt) Value() interface{} { return rs.value }

func Run(jobs chan Job, workerCount uint64) chan Result {
	results := make(chan Result)

	doneChan := make(chan ackType)
	slots := make(chan ackType, workerCount)

	slotWait := func() { <-slots }
	slotReady := func() { slots <- sentinel }
	jobDone := func() { doneChan <- sentinel }

	if workerCount < 1 {
		workerCount = 8 // TODO: Define a proper default
	}

	// Make free room for the workerCount
	for i := uint64(0); i < workerCount; i++ {
		slotReady()
	}

	doneCounterChan := make(chan ackType)
	go func() {
		defer close(doneCounterChan)
		for job := range jobs {
			if job != nil {
				doneCounterChan <- sentinel

				go func(j Job) {
					slotWait()

					result, err := j.Do()
					results <- resultSt{id: j.Id(), err: err, value: result}

					slotReady()
					jobDone()
				}(job)
			}
		}
	}()

	go func() {
		defer close(results)
		doneCount := uint64(0)
		for _ = range doneCounterChan {
			doneCount += 1
		}

		// Then finally wait until:
		for i := uint64(0); i < doneCount; i++ {
			<-doneChan
		}
	}()

	return results
}
