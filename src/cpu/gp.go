package cpu

import (
	//"fmt"
    "github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"

)

type GP struct {
	P1 uint8
	cpu *CPU
	K_LEFT uint8
	K_RIGHT uint8
	K_UP uint8
	K_DOWN uint8
	old uint8
}

func NewGP(cpu *CPU) *GP {
    g := new(GP)
	g.P1 = 0x2f
	g.old=0x2f
	g.cpu = cpu
	sdl.EnableKeyRepeat(1,1)

    return g 
}

func (g *GP) handleKeyDown(e *sdl.KeyboardEvent ) {

	if g.P1 &0x10 == 0x10  {
		switch (e.Keysym.Sym) {
			case sdl.K_RETURN:
			    g.old  &= ^uint8(0x08)
			case sdl.K_SPACE:
			   g.old &=  ^uint8(0x04)
			case sdl.K_x:
			    g.old &=  ^uint8(0x02)
			case sdl.K_z:
          	    g.old &=  ^uint8(0x01)

		}	

	}	
	if g.P1 &0x20 == 0x20  {
		switch (e.Keysym.Sym) {
		    case sdl.K_DOWN:
			    g.old  &= ^uint8(0x08)
			case sdl.K_UP:
			   g.old &=  ^uint8(0x04)
			case sdl.K_LEFT:
			    g.old &=  ^uint8(0x02)
			case sdl.K_RIGHT:
          	    g.old &=  ^uint8(0x01)
		}
	}
}

func (g *GP) handleKeyUp(e *sdl.KeyboardEvent ) {

	if g.P1 &0x10 == 0x10  {
		switch (e.Keysym.Sym) {
			case sdl.K_RETURN:
			    g.old |=0x08
			case sdl.K_SPACE:
			    g.old |=0x04
			case sdl.K_x:
			    g.old |=0x02
			case sdl.K_z:
			    g.old |=0x01

		}	

	}	
	if g.P1 &0x20 == 0x20  {
		switch (e.Keysym.Sym) {
			case sdl.K_DOWN:
			    g.old |=0x08
			case sdl.K_UP:
			    g.old |=0x04
			case sdl.K_LEFT:
			    g.old |=0x02
			case sdl.K_RIGHT:
			    g.old |=0x01

		}
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

			//fmt.Printf("0x%x\n",g)

				
			}
		default:

			break
	

		}
		break
	}

	g.P1=g.old
}


