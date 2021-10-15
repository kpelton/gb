#!/bin/bash
GO=`which go`
export GOPATH=$(pwd) 
export PATH=$PATH:$GOOROOT/bin
$GO install cpu ;$GO build main.go
#go install cpu; go build main.go
