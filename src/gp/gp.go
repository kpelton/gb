package gp

import (
		"fmt"
	//"github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"
	//"github.com/banthar/Go-SDL/sdl"
	"github.com/veandco/go-sdl2/sdl"
	"component"
)
const ( 
	GP_MMIO = 0xff00
)
type GP struct {
	P1      uint8
	K_LEFT  uint8
	K_RIGHT uint8
	K_UP    uint8
	K_DOWN  uint8
	pad     uint8
	other   uint8
   
	reg_list component.RegList
	
}

func NewGP() *GP {
	g := new(GP)
	g.Reset()
	//:sdl.EnableKeyRepeat(1, 1)
	g.reg_list = component.RegList{
		{Name:"GP" , Addr:GP_MMIO},
	}
	return g
}
func (g* GP) Get_reg_list() component.RegList{
	return g.reg_list
}
func (g *GP) Reset() {
	g.P1 = 0x0f
	g.other = 0x0f
	g.pad = 0x0f
}
func (g *GP) Read_mmio(addr uint16) uint8 {
	var val uint8
	switch addr {
	case GP_MMIO:
		g.Update()
		val = g.P1
	default:
		panic("GP: unhandled gp read")
	}
	return val
}
func (g *GP) Write_mmio(addr uint16,val uint8) {
	switch addr {
	case GP_MMIO:
		g.P1 = val
	default:
		panic("GP: unhandled gp write")
	}
}


func (g *GP) handleKeyDown(e *sdl.KeyDownEvent) {

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

func (g *GP) handleKeyUp(e *sdl.KeyUpEvent) bool {


	switch e.Keysym.Sym {
	case sdl.K_RETURN:
		g.other |= 0x08
	case sdl.K_SPACE:
		g.other |= 0x04
	case sdl.K_x:
		g.other |= 0x02
	case sdl.K_z:
		g.other |= 0x01
    case sdl.K_F1:

       //return true to indicate global event
       return true

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
    return false
	//fmt.Printf("P1:0x%02x\n",g.P1)
}
func (g *GP) LoopUpdate() (uint8) {
    for {
    g.Update()
    }
}
func (g *GP) Update() (uint8) {
    var int_raised uint8 = 0
		ev := sdl.PollEvent()

		switch e := ev.(type) {

		case *sdl.KeyUpEvent:
                if  g.handleKeyUp(e) == true {
                    int_raised = 0xff
                    fmt.Println("gp global")
                }else {
                    int_raised = 0x10
                }

		case *sdl.KeyDownEvent:
				g.handleKeyDown(e)
                int_raised = 0x10	    

		default:
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
