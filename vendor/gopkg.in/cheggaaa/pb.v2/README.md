# Terminal progress bar for Go  

[![Coverage Status](https://coveralls.io/repos/github/cheggaaa/pb/badge.svg?branch=v2)](https://coveralls.io/github/cheggaaa/pb?branch=v2)

### It's beta, some features may be changed

This is proposal for the second version of progress bar   
- based on text/template   
- can take custom elements   
- using colors is easy   

## Installation

```
go get gopkg.in/cheggaaa/pb.v2
```   

## Usage   

```Go
package main

import (
	"gopkg.in/cheggaaa/pb.v2"
	"time"
)

func main() {
	simple()
	fromPreset()
	customTemplate(`Custom template: {{counters . }}`)
	customTemplate(`{{ red "With colors:" }} {{bar . | green}} {{speed . | blue }}`)
	customTemplate(`{{ red "With funcs:" }} {{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{speed . | rndcolor }}`)
	customTemplate(`{{ bar . "[<" "·····•·····" (rnd "ᗧ" "◔" "◕" "◷" ) "•" ">]"}}`)
}

func simple() {
	count := 1000
	bar := pb.StartNew(count)
	for i := 0; i < count; i++ {
		bar.Increment()
		time.Sleep(time.Millisecond * 2)
	}
	bar.Finish()
}

func fromPreset() {
	count := 1000
	//bar := pb.Default.Start(total)
	//bar := pb.Simple.Start(total)
	bar := pb.Full.Start(count)
	defer bar.Finish()
	bar.Set("prefix", "fromPreset(): ")
	for i := 0; i < count/2; i++ {
		bar.Add(2)
		time.Sleep(time.Millisecond * 4)
	}
}

func customTemplate(tmpl string) {
	count := 1000
	bar := pb.ProgressBarTemplate(tmpl).Start(count)
	defer bar.Finish()
	for i := 0; i < count/2; i++ {
		bar.Add(2)
		time.Sleep(time.Millisecond * 4)
	}
}

```