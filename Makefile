# Makefile for cross-compilation
#
OS := $(shell uname)
BINDIR := ./bin
MD5_TEXTFILE := $(BINDIR)/md5Sums.txt

MAIN_FILE_DIR := ./cmd/drive

ifeq ($(OS), Darwin)
  MD5_UTIL = md5
else
  MD5_UTIL = md5sum
endif

all: compileThemAll md5SumThemAll

compileThemAll: armv5 armv6 armv7 armv8 darwin linux

md5SumThemAll:
	rm -f $(MD5_TEXTFILE)
	find $(BINDIR) -type f -name "drive_*" -exec $(MD5_UTIL) {} >> $(MD5_TEXTFILE) \;

armv5:
	CGO_ENABLED=0 GOOS=linux GOARM=5 GOARCH=arm go build -o $(BINDIR)/drive_armv5 $(MAIN_FILE_DIR)
armv6:
	CGO_ENABLED=0 GOOS=linux GOARM=6 GOARCH=arm go build -o $(BINDIR)/drive_armv6 $(MAIN_FILE_DIR)
armv7:
	CGO_ENABLED=0 GOOS=linux GOARM=7 GOARCH=arm go build -o $(BINDIR)/drive_armv7 $(MAIN_FILE_DIR)
armv8:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o $(BINDIR)/drive_armv8 $(MAIN_FILE_DIR)
darwin:
	CGO_ENABLED=0 GOOS=darwin go build -o $(BINDIR)/drive_darwin $(MAIN_FILE_DIR)
linux:
	CGO_ENABLED=0 GOOS=linux go build -o $(BINDIR)/drive_linux $(MAIN_FILE_DIR)
