package main

import (
	"fmt"
	"os"

	"github.com/odeke-em/xon/pkger/src"
)

func main() {
	pkgInfo, err := pkger.Recon(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Println(pkgInfo)
	}
}
