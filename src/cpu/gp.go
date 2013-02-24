package cpu

import (
//	"fmt"
    "github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"

)

type GP struct {
	P1 uint8
	cpu *CPU
	K_LEFT uint8
	K_RIGHT uint8
	K_UP uint8
	K_DOWN uint8
	pad uint8
	other uint8
}

func NewGP(cpu *CPU) *GP {
    g := new(GP)
	g.P1 = 0x0f
	g.other=0x0f
	g.pad =0x0f
	g.cpu = cpu
	sdl.EnableKeyRepeat(1,1)

    return g 
}

func (g *GP) handleKeyDown(e *sdl.KeyboardEvent ) {

	if g.P1 &0x10 == 0x10  {
		switch (e.Keysym.Sym) {
			case sdl.K_RETURN:
			    g.other  &= ^uint8(0x08)
			case sdl.K_SPACE:
			   g.other &=  ^uint8(0x04)
			case sdl.K_x:
			    g.other &=  ^uint8(0x02)
			case sdl.K_z:
          	    g.other &=  ^uint8(0x01)

		}	

	}	
	if g.P1 &0x20 == 0x20  {
		switch (e.Keysym.Sym) {
		    case sdl.K_DOWN:
			    g.pad  &= ^uint8(0x08)
			case sdl.K_UP:
			   g.pad &=  ^uint8(0x04)
			case sdl.K_LEFT:
			    g.pad &=  ^uint8(0x02)
			case sdl.K_RIGHT:
          	    g.pad  &=  ^uint8(0x01)
		}
	}
}

func (g *GP) handleKeyUp(e *sdl.KeyboardEvent ) {

//	if g.P1 &0x10 == 0x10  {
		switch (e.Keysym.Sym) {
			case sdl.K_RETURN:
			    g.other |=0x08
			case sdl.K_SPACE:
			    g.other |=0x04
			case sdl.K_x:
			    g.other |=0x02
			case sdl.K_z:
			    g.other |=0x01

	//	}	

	}	
//	if g.P1 &0x20 == 0x20  {
		switch (e.Keysym.Sym) {
			case sdl.K_DOWN:
			    g.pad |=0x08
			case sdl.K_UP:
			    g.pad |=0x04
			case sdl.K_LEFT:
			    g.pad |=0x02
			case sdl.K_RIGHT:
			    g.pad |=0x01

	//	}
	}

	//fmt.Printf("P1:0x%02x\n",g.P1)
}


func (g *GP) Update(){
	
	for {
		select {

		case event := <-sdl.Events:
			switch e := event.(type) {	
			case sdl.KeyboardEvent:
				if e.Type == 2  { //KeyDown
					g.handleKeyDown(&e)
				
				}else{
					g.handleKeyUp(&e)
//					g.P1 = 0x0f
					if 	g.cpu.mmu.read_b(0xffff) &0x10 == 0x10 {
						//INT
						g.cpu.mmu.write_b(0xff0f,(g.cpu.mmu.read_b(0xff0f) |0x08))
					}
				}


				
			}
		default:

			break
	

		}
		break
	}

	if g.P1 &0x20 == 0x20 {
		g.P1 |= g.pad
	}
	if g.P1 &0x10 == 0x10 {
		g.P1 |= g.other
	}
	//g.P1|=g.old			
//	fmt.Printf("P1:0x%x,PAD:0x%0x,OTHER:0x%0x\n",g.P1,g.pad,g.other)

	
}


