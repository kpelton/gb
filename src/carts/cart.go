package carts

import (
	"fmt"
	"os"
	"strings"
)

type Cart interface {
	Read_b(uint16) uint8
	Write_b(uint16, uint8)
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

	case C_ROM_RAM:
		fmt.Printf("ROM_RAM\n")
	default:

		fmt.Printf("Unknown!\n")
		//  panic("Unsupported cart!!!!")
		cart = NewROM_MBC1(cart_name, data, size, true)

	}
	return cart
}

/////MBC0///////
type MBC0 struct {
	cart [0x8000]uint8
}

func NewMBC0(cart_data []uint8) *MBC0 {
	m := new(MBC0)
	copy(m.cart[:], cart_data)

	return m
}

func (m *MBC0) Read_b(addr uint16) uint8 {
	return m.cart[addr]
}

func (m *MBC0) Write_b(addr uint16, val uint8) {
	fmt.Printf("WRITE TO ROM FAILED!!\n")
}

/////MBC1///////
const (
	SIXTEEN_MB = 0
	FOUR_MB    = 1
)

type ROM_MBC1 struct {
	cart        [0x200000]uint8
	bank        uint16
	ram_enabled bool
	ram_bank    uint8
	ram         [0x20000]uint8
	memory_mode uint8
	file        os.File
	name        string
	has_battery bool
}

func NewROM_MBC1(name string, cart_data []uint8, size int, has_battery bool) *ROM_MBC1 {
	m := new(ROM_MBC1)
	fmt.Println(size)
	copy(m.cart[:], cart_data)
	m.memory_mode = SIXTEEN_MB
	m.name = name
	m.has_battery = has_battery
	if has_battery == true {
		m.Load_ram()
	}
	return m
}
func (m *ROM_MBC1) Load_ram() {
	save_name := m.name + ".data"

	file, err := os.OpenFile(save_name, os.O_RDWR, 666) // For read access.
	if err != nil {
		//fmt.Println(err)
		file, err = os.Create(save_name) // For read access.
	} else {
		fmt.Println("Read save data")
		file.Read(m.ram[0:])
	}

	m.file = *file

}

func (m *ROM_MBC1) Save_ram() {

	m.file.WriteAt(m.ram[0:], 0)
}

func (m *ROM_MBC1) Read_b(addr uint16) uint8 {
	var retval uint8

	if addr < 0x4000 {
		retval = m.cart[addr]
	} else if addr < 0x8000 {
		retval = m.cart[uint32(addr)+(uint32(m.bank)*0x4000)]
	} else {
		if (m.memory_mode == FOUR_MB && m.ram_enabled == true) || (m.memory_mode == SIXTEEN_MB) {
			bank_offset := uint16(uint32(m.ram_bank) * 0x2000)
			fixed_addr := uint16(addr-0xa000) + bank_offset
			fmt.Printf("RAM  BANK READ:%v  %04X->%04X:%x\n", m.ram_bank, addr, fixed_addr, retval)

			retval = m.ram[fixed_addr]
		} else {
			panic("Tried to read from ram that wasn't enabled!")
		}
	}

	return retval
}

func (m *ROM_MBC1) Write_b(addr uint16, val uint8) {
	if addr >= 0x2000 && addr < 0x4000 {
		if val > 1 {
			//fmt.Println("ROM Bank from",m.bank,val-1)
			m.bank = uint16((val) - 1)
		} else {
			m.bank = uint16(0)
		}
	} else if addr < 0x2000 {
		if m.memory_mode == FOUR_MB {
			if val == 0x0A {
				m.ram_enabled = true
				fmt.Println("RAM enabled", val)

			} else {
				fmt.Println("RAM Disabled", val)
				m.ram_enabled = false

			}
		}

	} else if addr >= 0x4000 && addr < 0x6000 {

		//fmt.Println("RAM bank", "from", m.ram_bank, "to", val&0xf)
		m.ram_bank = (val & 0xf)
	} else if addr >= 0x6000 && addr < 0x7000 {
		if val > 0 {
			m.memory_mode = SIXTEEN_MB
			fmt.Println("RAM Memory mode", val, "selected")
		} else {
			m.memory_mode = FOUR_MB
		}

	} else if addr >= 0xA000 && addr < 0xc000 {

		if (m.memory_mode == FOUR_MB && m.ram_enabled == true) || (m.memory_mode == SIXTEEN_MB) {

			bank_offset := uint16(uint32(m.ram_bank) * 0x2000)
			fixed_addr := uint16(addr-0xa000) + bank_offset
			fmt.Printf("RAM  BANK WRITE:%v  %04X->%04X:%x\n", m.ram_bank, addr, fixed_addr, val)

			m.ram[fixed_addr] = val
			if m.has_battery {
				m.Save_ram()
			}
		} else {
			panic("Tried to read from ram that wasn't enabled!")
		}
	}

}

/////MBC2///////
type ROM_MBC2 struct {
	cart [0x200000]uint8
	bank uint16
}

func NewROM_MBC2(cart_data []uint8, size int) *ROM_MBC2 {
	m := new(ROM_MBC2)
	fmt.Println(size)
	copy(m.cart[:], cart_data)
	return m
}

func (m *ROM_MBC2) Read_b(addr uint16) uint8 {
	var retval uint8

	if addr < 0x4000 {
		//always ROM bank #0
		retval = m.cart[addr]
	} else if addr < 0xc000 {
		retval = m.cart[uint32(addr)+(uint32(m.bank)*0x4000)]
	}
	return retval
}

func (m *ROM_MBC2) Write_b(addr uint16, val uint8) {
	if addr > 0x2000 && addr < 0x4000 {
		if val > 1 {

			//fmt.Println("Bank from",m.bank,val-1)
			m.bank = uint16(val - 1)
		} else {
			m.bank = uint16(0)
		}
	}

}
