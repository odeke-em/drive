all: y.go
	go install

y.go: cc.y
	go tool yacc cc.y

try: y.go
	go install
	go run main.go /Users/rsc/g/go/src/libmach/8db.c

html: y.go
	go install
	go run tohtml.go  /Users/rsc/g/go/src/libmach/8db.c >x.html
