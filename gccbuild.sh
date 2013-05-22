#!/bin/bash
export GOPATH=$(pwd) 
rm pkg/gccgo/libcpu.a
go build -compiler=gccgo -gccgoflags="-O3 -L/usr/lib/x86_64-linux-gnu/libSDL-1.2.so.0" main.go

go install -compiler gccgo -gccgoflags="-O3  -L/usr/lib/x86_64-linux-gnu/libSDL-1.2.so.0"  cpu ;

