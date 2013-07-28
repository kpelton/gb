package cpu

import (
	"fmt"

//    "time"
)

type IC struct {
	IE uint8 //Interrupt Enable (R/W)
	IF uint8 // Interrupt Flag (R/W)

}

const (
	V_BLANK = 1
	LCDC    = 0x2
	TIMER   = 0x4
	SERIAL  = 0x8
	GAME    = 0x10
    NONE    =0x0
)

func NewIC() *IC {
	return new(IC)
}

func (i *IC) Assert(signal uint8) {
	//check to see if it is masked off
	//fmt.Println("ASSERT",signal,i.IE,i.IF)

	i.IF |= signal
}

func (i *IC) Disassert(signal uint8) {

	i.IF &= ^signal
	//fmt.Println("Disassert",signal,i.IF,i.IE)
}

func (i *IC) Handle() uint16 {
	var value uint16 = 0

	//fmt.Println("IF",i.IF,i.IE)
	switch {

	case (i.IF&V_BLANK == V_BLANK) && (i.IE&V_BLANK == V_BLANK):
		i.Disassert(V_BLANK)
		//          fmt.Println("X")

		return (0x40)
	case (i.IF&LCDC == LCDC) && (i.IE&LCDC == LCDC):
		i.Disassert(LCDC)
		// fmt.Println("INT","LCDC")
		return (0x48)

	case (i.IF&TIMER == TIMER) && (i.IE&TIMER == TIMER):
		i.Disassert(TIMER)

		return (0x50)

	case (i.IF&GAME == GAME) && (i.IE&GAME == GAME):
		i.Disassert(GAME)
		return (0x60)
	case (i.IF&SERIAL == SERIAL) && (i.IE&SERIAL == SERIAL):
		i.Disassert(SERIAL)
		fmt.Println("X")
		return (0x58)

	}
	return (value)
}
