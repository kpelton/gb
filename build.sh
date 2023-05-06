#!/bin/bash
export GO="/home/kyle/Downloads/go/bin/go"
export GOPATH=$(pwd) 
export PATH=$PATH:$GOOROOT/bin
export GO111MODULE="off"
$GO install cpu ;$GO build main.go
#go install cpu; go build main.go
