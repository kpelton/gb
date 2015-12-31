package carts

import (
	"fmt"
)

/////MBC2///////
type ROM_MBC2 struct {
	cart [0x200000]uint8
	bank uint16
	GenCart
}

func NewROM_MBC2(cart_data []uint8, size int) *ROM_MBC2 {
	m := new(ROM_MBC2)
	fmt.Println(size)
	copy(m.cart[:], cart_data)
	return m
}

func (m *ROM_MBC2) Read(addr uint16) uint8 {
	var retval uint8

	if addr < 0x4000 {
		//always ROM bank #0
		retval = m.cart[addr]
	} else if addr < 0xc000 {
		retval = m.cart[uint32(addr)+(uint32(m.bank)*0x4000)]
	}
	return retval
}

func (m *ROM_MBC2) Write(addr uint16, val uint8) {
	if addr > 0x2000 && addr < 0x4000 {
		if val > 1 {

			//fmt.Println("Bank from",m.bank,val-1)
			m.bank = uint16(val - 1)
		} else {
			m.bank = uint16(0)
		}
	}

}
