#!/bin/bash
export GOPATH=$(pwd) 
rm pkg/gccgo/libcpu.a
go build -compiler=gccgo -gccgoflags="-O3 -I /usr/lib/go/pkg/gccgo" main.go
go install -compiler gccgo -gccgoflags="-O3 -I /usr/lib/go/pkg/gccgo"  cpu ;

