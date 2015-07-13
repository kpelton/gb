#!/bin/bash
export GOPATH=$(pwd) 
export PATH=$PATH:$GOOROOT/bin
GO=/home/kyle/Downloads/go/bin/go
export GOROOT=/home/kyle/Downloads/go

rm pkg/gccgo/libcpu.a
$GO install -compiler gccgo -gccgoflags="-O3  -lm -lSDL_ttf -lSDL -I/usr/include/SDL -I/usr/local/include/SDL"  github.com/banthar/Go-SDL/sdl;
$GO install -compiler gccgo -gccgoflags="-O3  -lm -lSDL_ttf -lSDL -I/usr/include/SDL -I/usr/local/include/SDL"  cpu ;
$GO build -compiler=gccgo -gccgoflags="-O3  -lm -lSDL_ttf -lSDL -I/usr/include/SDL -I/usr/local/include/SDL" main.go



