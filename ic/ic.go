package ic

import (
	"fmt"
	"gb/component"
	"gb/constants"
)

type IC struct {
	IE       uint8 //Interrupt Enable (R/W)
	IF       uint8 // Interrupt Flag (R/W)
	reg_list component.RegList
}

const (
	MMIO_IE = 0xffff
	MMIO_IF = 0xff0f
)

func (ic *IC) Get_reg_list() component.RegList {
	return ic.reg_list
}
func NewIC() *IC {
	ic := new(IC)
	ic.reg_list = component.RegList{
		{Name: "IE", Addr: MMIO_IE},
		{Name: "IF", Addr: MMIO_IF},
	}
	ic.Reset()
	return ic
}
func (i *IC) Reset() {
	i.IF = 0
	i.IE = 0
}
func (i *IC) Write_mmio(addr uint16, val uint8) {
	switch addr {
	case 0xff0f:
		i.IF = val
	case 0xffff:
		i.IE = val
	default:
		panic("IC:unhandled mmio write")
	}
}

func (i *IC) Read_mmio(addr uint16) uint8 {
	var val uint8
	switch addr {
	case 0xff0f:
		val = i.IF | 0xe0
	case 0xffff:
		val = i.IE
	default:
		panic("IC:unhandled mmio write")

	}
	return val
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

	case (i.IF&constants.V_BLANK == constants.V_BLANK) && (i.IE&constants.V_BLANK == constants.V_BLANK):
		i.Disassert(constants.V_BLANK)
		//          fmt.Println("X")
		//fmt.Println("VBLANK")

		return (0x40)
	case (i.IF&constants.LCDC == constants.LCDC) && (i.IE&constants.LCDC == constants.LCDC):
		i.Disassert(constants.LCDC)
		fmt.Println("LCDC")

		// fmt.Println("INT","constants.LCDC")
		return (0x48)

	case (i.IF&constants.TIMER == constants.TIMER) && (i.IE&constants.TIMER == constants.TIMER):
		i.Disassert(constants.TIMER)
		fmt.Println("TIMER")
		return (0x50)

	case (i.IF&constants.GAME == constants.GAME) && (i.IE&constants.GAME == constants.GAME):
		i.Disassert(constants.GAME)
		return (0x60)
	case (i.IF&constants.SERIAL == constants.SERIAL) && (i.IE&constants.SERIAL == constants.SERIAL):
		i.Disassert(constants.SERIAL)
		fmt.Println("X")
		return (0x58)

	}
	return (value)
}
