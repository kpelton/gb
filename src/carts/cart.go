package carts

import (
	"fmt"
	"os"
	"strings"
	"component"
)
type GenCart struct {

}
type Cart interface {
	Read(uint16) uint8
	Write(uint16, uint8)
	Get_range_list() component.RangeList

}
const (
	SIXTEEN_MB = 0
	FOUR_MB    = 1
)
func (c* GenCart ) Get_range_list() component.RangeList{
	return component.RangeList{
		{Name:"CART_ROM",Addr_lo:0x000,Addr_hi:0x8000},
		{Name:"CART_RAM",Addr_lo:0xa000,Addr_hi:0xc000},

	}
}

const (
	REG_CART_TYPE       = 0x147
	REG_CART_SIZE       = 0x148
	REG_RAM_SIZE        = 0x148
	REG_CART_NAME       = 0x134
	REG_CART_NAME_SIZE  = 16
	C_ROM_ONLY          = 0
	C_ROM_MBC1          = 1
	C_ROM_MBC1_RAM      = 2
	C_ROM_MBC1_RAM_BATT = 3
	C_ROM_MBC2          = 5
	C_ROM_MBC2_BATT     = 6
	C_ROM_MBC3_RAM_BATT = 13
	C_ROM_MBC5 = 0x19
	C_ROM_MBC5_RAM_BATT  = 0x1b

	C_ROM_MBC5_RUM  = 0x1C
	C_ROM_MBC5_RUM_RAM_BATT  = 0x1e
	C_ROM_RAM           = 8
)

func Load_cart(filename string) Cart {
	fi, err := os.Open(filename)
	buf := make([]uint8, 0x400000)

	n, err := fi.Read(buf)

	if err != nil || n == 0 {
		panic(err)
	}
	return create_new_cart(buf, n)
}

func create_new_cart(data []uint8, size int) Cart {

	fmt.Printf("Cart Type:0%02x\n:", data[REG_CART_TYPE])
	fmt.Printf("Cart Size:0%02x:\n", data[REG_CART_SIZE])
	fmt.Printf("Ram Size:0%02x:\n", data[REG_RAM_SIZE])
	var cart Cart

	cart_type := data[REG_CART_TYPE]
	var name [16]uint8
	length := 0
	for offset := REG_CART_NAME; offset-REG_CART_NAME < REG_CART_NAME_SIZE; offset++ {
		if data[offset] == 0 {
			break
		}
		name[offset-REG_CART_NAME] = data[offset]
		length++
	}
	cart_name := strings.ToLower(fmt.Sprintf("%s", name[0:length]))
	fmt.Printf("Cart Name:%s\n", cart_name)

	fmt.Printf("Cart Type:")
	switch cart_type {
	case C_ROM_ONLY:
		fmt.Printf("ROM_ONLY\n")
		cart = NewMBC0(data[:0x8000])
	case C_ROM_MBC1:
		fmt.Printf("ROM_MBC1\n")
		cart = NewROM_MBC1(cart_name, data, size, false)

	case C_ROM_MBC1_RAM:
		fmt.Printf("ROM_MBC1_RAM\n")
		cart = NewROM_MBC1(cart_name, data, size, false)

	case C_ROM_MBC1_RAM_BATT:
		cart = NewROM_MBC1(cart_name, data, size, true)

		fmt.Printf("ROM_MBC1_RAM_BATT\n")
	case C_ROM_MBC2:
		fmt.Printf("ROM_MBC2\n")
	case C_ROM_MBC2_BATT:
		cart = NewROM_MBC2(data, size)
	case C_ROM_MBC3_RAM_BATT:
		cart = NewROM_MBC1(cart_name, data, size, true)
	case C_ROM_MBC5_RAM_BATT:
		cart = NewROM_MBC5(cart_name, data, size, true)
	case C_ROM_MBC5:
		cart = NewROM_MBC5(cart_name, data, size, false)
	case C_ROM_MBC5_RUM:
		cart = NewROM_MBC5(cart_name, data, size, false)
	case C_ROM_MBC5_RUM_RAM_BATT:
		cart = NewROM_MBC5(cart_name, data, size, true)


	case C_ROM_RAM:
		fmt.Printf("ROM_RAM\n")
	default:

//		fmt.Printf("Unknown!\n")
//		  panic("Unsupported cart!!!!")
		cart = NewROM_MBC1(cart_name, data, size, true)

	}
	return cart
}

