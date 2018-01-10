package main

import (
	"fmt"

	expb "github.com/odeke-em/exponential-backoff"
)

func consume(result interface{}, err error) {
	fmt.Printf("result: %v err: %v\n", result, err)
}

func main() {
	backer := expb.NewUrlGetter("https://golang.org/pkg/net/httpx", 5)
	expb.ExponentialBackOff(backer, consume)
	fmt.Println("expb")
}
