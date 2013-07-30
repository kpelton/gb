package gp

import (
	//	"fmt"
	//"github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"
	"github.com/banthar/Go-SDL/sdl"
)

type GP struct {
	P1      uint8
	K_LEFT  uint8
	K_RIGHT uint8
	K_UP    uint8
	K_DOWN  uint8
	pad     uint8
	other   uint8
}

func NewGP() *GP {
	g := new(GP)
	g.P1 = 0x0f
	g.other = 0x0f
	g.pad = 0x0f
	sdl.EnableKeyRepeat(1, 1)

	return g
}

func (g *GP) handleKeyDown(e *sdl.KeyboardEvent) {

	switch e.Keysym.Sym {
	case sdl.K_RETURN:
		g.other &= ^uint8(0x08)
	case sdl.K_SPACE:
		g.other &= ^uint8(0x04)
	case sdl.K_x:
		g.other &= ^uint8(0x02)
	case sdl.K_z:
		g.other &= ^uint8(0x01)

	}

	switch e.Keysym.Sym {
	case sdl.K_DOWN:
		g.pad &= ^uint8(0x08)
	case sdl.K_UP:
		g.pad &= ^uint8(0x04)
	case sdl.K_LEFT:
		g.pad &= ^uint8(0x02)
	case sdl.K_RIGHT:
		g.pad &= ^uint8(0x01)
	}
}

func (g *GP) handleKeyUp(e *sdl.KeyboardEvent) {


	switch e.Keysym.Sym {
	case sdl.K_RETURN:
		g.other |= 0x08
	case sdl.K_SPACE:
		g.other |= 0x04
	case sdl.K_x:
		g.other |= 0x02
	case sdl.K_z:
		g.other |= 0x01



	}

	switch e.Keysym.Sym {
	case sdl.K_DOWN:
		g.pad |= 0x08
	case sdl.K_UP:
		g.pad |= 0x04
	case sdl.K_LEFT:
		g.pad |= 0x02
	case sdl.K_RIGHT:
		g.pad |= 0x01


	}

	//fmt.Printf("P1:0x%02x\n",g.P1)
}

func (g *GP) Update() (uint8) {
    var int_raised uint8 = 0
	for {
		ev := sdl.PollEvent()

		//fmt.Println(ev)
		switch e := ev.(type) {

		case *sdl.KeyboardEvent:
    		if e.Type == sdl.KEYDOWN {
				g.handleKeyDown(e)
                int_raised = 0x10	    
			} else {
				g.handleKeyUp(e)
                int_raised = 0x10	    

			}

		default:
			break
		}
		break
	}

	if g.P1&0x20 == 0x20 {
		g.P1 |= g.pad
	}
	if g.P1&0x10 == 0x10 {
		g.P1 |= g.other
	}
    return int_raised
}
