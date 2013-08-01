#!/bin/bash
GO=/home/kyle/Downloads/go/bin/go
export GOPATH=$(pwd) 
export GOROOT=/home/kyle/Downloads/go
export PATH=$PATH:$GOOROOT/bin
$GO install cpu ;$GO build main.go
#go install cpu; go build main.go
