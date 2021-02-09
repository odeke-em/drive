### strcmp-checker

* It reads lines from stdin, compares each line read with itself then if the line is not equal to itself. Just to stress test a weird exception that am investigating on Linux Mint 17 from [drive issue #406](https://github.com/odeke-em/drive/issues/406)

### Installation

```shell
$ go get github.com/odeke-em/go-utils/strcmp-check
```

### Usage

```shell
$ echo "hey" | strcmp-check
$ cat content | strcmp-check
```
