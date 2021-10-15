# Golang Gameboy
Gameboy emulator written in golang with sound,network and gameboy color support.
## Games
Mario DX

![mario](images/mario.png)

Zelda

![zelda](images/zelda.png)

## Install
Tested on ubuntu 20.04 golang 1.13

```
export GOPATH=. #build directory
go get -v github.com/tarm/goserial
go get -v github.com/veandco/go-sdl2/sdl
./build.sh
```

## Running
```
./main cpu_instrs.gb # or any other .gb/.gbc rom
```
