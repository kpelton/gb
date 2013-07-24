#!/bin/bash
export GOPATH=$(pwd) 
rm pkg/gccgo/libcpu.a
go install -compiler gccgo -gccgoflags="-O3  -lm -lSDL_ttf -lSDL -I/usr/include/SDL -I/usr/local/include/SDL"  github.com/banthar/Go-SDL/sdl;
go install -compiler gccgo -gccgoflags="-O3  -lm -lSDL_ttf -lSDL -I/usr/include/SDL -I/usr/local/include/SDL"  cpu ;
go build -compiler=gccgo -gccgoflags="-O3  -lm -lSDL_ttf -lSDL -I/usr/include/SDL -I/usr/local/include/SDL" main.go



