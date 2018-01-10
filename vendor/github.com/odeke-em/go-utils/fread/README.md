## fread

Package to read from an io.Reader similar to C's fread(\*FILE,...)

and it returns a channel of strings `chan string`


## Usage in programs

```go
import "github.com/odeke-em/go-utils/fread"

...

f := os.Stdin  // Could be something else
linesChan := fread.Fread(f)
linesChanWithIgnorer := fread.Fread(f, func(s string) bool { return len(s) >= 1 && s[0] == "#"} )
```

## Sample program

```shell
$ go get github.com/odeke-em/go-utils/cmd/stdin-io
```

That will just read content from stdin and dump it to stdout.
