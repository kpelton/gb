package carts

import (
	"fmt"
)

/////MBC0///////
type MBC0 struct {
	cart [0x8000]uint8
	GenCart
}

func NewMBC0(cart_data []uint8) *MBC0 {
	m := new(MBC0)
	copy(m.cart[:], cart_data)

	return m
}

func (m *MBC0) Read(addr uint16) uint8 {
	return m.cart[addr]
}

func (m *MBC0) Write(addr uint16, val uint8) {
	fmt.Printf("WRITE TO ROM FAILED!!\n")
}
