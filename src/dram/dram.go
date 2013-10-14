package dram

import (
	"fmt"

)

type DRAM struct {
	ram         [0x1000]uint8
    exp_ram     [0x7000]uint8
	z_ram       [0x7f]uint8
	SVBK  uint8
}
const (
	BANK_0_LO = 0xc000
	BANK_0_HI = 0xd000

	EBANK_LO = 0xd000
	EBANK_HI = 0xe000

	ECHO_HI = 0xfe00
	ECHO_LO  = 0xe000

	Z_RAM_LO   = 0xff80
	Z_RAM_HI  = 0xffff
	
	SVBK_MMIO = 0xff70
)
	

func NewDRAM() *DRAM {
	m := new(DRAM)
	m.SVBK = 1
	
	return m

}

func (m *DRAM) Write_mmio(addr uint16,val uint8)  {
	if addr == SVBK_MMIO {
		m.SVBK = val & 0x7
		if m.SVBK == 0 {
			m.SVBK = 1
		}
	} else { 
		panic("unhandled DRAM mmio write")
	}

}

func (m *DRAM) Read_mmio(addr uint16) uint8  {
	if addr == SVBK_MMIO {
		return m.SVBK
	} else { 
		panic("unhandled DRAM mmio write")
	}

}


func (m *DRAM) Read_b(addr uint16) uint8 {

	//   fmt.Printf("write:%04x:%04x\n",addr,val)
	var val uint8

	if addr >= BANK_0_LO && addr < BANK_0_HI {
		val = m.ram[addr&0xfff]
    }else if addr  >= EBANK_LO && addr < EBANK_HI {
		offset :=(addr&0xfff) +(0x1000 *uint16(m.SVBK-1) )
        val = m.exp_ram[offset]
	}else if addr >= ECHO_LO && addr < ECHO_HI {
		val = m.ram[(addr-0x2000)&0x1fff]
	} else if addr >= Z_RAM_LO && addr < Z_RAM_HI {
		val = m.z_ram[(addr&0x00ff)-0x80]
	} else {
		fmt.Printf("unhandled read:%04x:%04x\n", addr, val)
	}
	return val
}

func (m *DRAM) Write_b(addr uint16,val uint8)  {

	 //  fmt.Printf("write:%04x:%04x\n",addr,val)

	if addr >= BANK_0_LO && addr < BANK_0_HI {
		m.ram[addr&0xfff] = val
    }else if addr  >= EBANK_LO && addr < EBANK_HI {
		offset :=(addr&0xfff) +(0x1000 *uint16(m.SVBK-1) )
		m.exp_ram[offset] = val
	}else if addr >= ECHO_LO && addr < ECHO_HI {
		m.ram[(addr-0x2000)&0x1fff] = val
	} else if addr >= Z_RAM_LO && addr < Z_RAM_HI {
		m.z_ram[(addr&0x00ff)-0x80] = val
	} else {
		fmt.Printf("unhandled write:%04x:%04x\n", addr, val)
	}
}


