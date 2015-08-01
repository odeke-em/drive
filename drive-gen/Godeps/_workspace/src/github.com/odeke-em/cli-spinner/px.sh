#!/usr/bin/env bash
cat << ! > ~/.bashrc
export GOPATH="\$HOME/gopath"
export PATH="\$GOPATH:\$GOPATH/bin:\$PATH"
!
