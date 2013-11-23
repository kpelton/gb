package dram

import (
	"fmt"
	"component"
)

type DRAM struct {
	ram         [0x1000]uint8
    exp_ram     [0x7000]uint8
	z_ram       [0x7f]uint8
	SVBK  uint8
	reg_list component.RegList
	range_list component.RangeList

}
const (
	BANK_0_LO = 0xc000
	BANK_0_HI = 0xd000

	EBANK_LO = 0xd000
	EBANK_HI = 0xe000

	ECHO_B0_HI = 0xf000
	ECHO_B0_LO  = 0xe000

	ECHO_B1_HI = 0xfe00
	ECHO_B1_LO  = 0xf000

	Z_RAM_LO   = 0xff80
	Z_RAM_HI  = 0xffff
	
	SVBK_MMIO = 0xff70
)
	

func NewDRAM() *DRAM {
	m := new(DRAM)
	m.SVBK = 1
	m.reg_list = component.RegList{
		{Name:"SVBK",Addr:SVBK_MMIO},
	}
	m.range_list = component.RangeList{
		{Name:"DRAM",Addr_lo:0xc000,Addr_hi:0xfe00},
		{Name:"Z_RAM",Addr_lo:Z_RAM_LO,Addr_hi:Z_RAM_HI},
	}
	return m

}
func (m* DRAM) Get_reg_list() component.RegList{
	return m.reg_list
}

func (m* DRAM) Get_range_list() component.RangeList{
	return m.range_list
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


func (m *DRAM) Read(addr uint16) uint8 {

	//   fmt.Printf("write:%04x:%04x\n",addr,val)
	var val uint8

	if addr >= BANK_0_LO && addr < BANK_0_HI {
		val = m.ram[addr&0xfff]
    }else if addr  >= EBANK_LO && addr < EBANK_HI {
		offset :=(addr&0xfff) +(0x1000 *uint16(m.SVBK-1) )
        val = m.exp_ram[offset]
	}else if addr >= ECHO_B0_LO && addr < ECHO_B0_HI {
		val = m.ram[(addr-0x2000)&0xfff]
	}else if addr >= ECHO_B1_LO && addr < ECHO_B1_HI {
		new_addr := addr -0x3000 
		offset :=(new_addr&0xfff) +(0x1000 *uint16(m.SVBK-1) )
        val = m.exp_ram[offset]
	} else if addr >= Z_RAM_LO && addr < Z_RAM_HI {
		val = m.z_ram[(addr&0x00ff)-0x80]
	} else {
		fmt.Printf("unhandled read:%04x:%04x\n", addr, val)
	}
	return val
}

func (m *DRAM) Write(addr uint16,val uint8)  {

	//fmt.Printf("DRAM write:%04x:%04x\n",addr,val)

	if addr >= BANK_0_LO && addr < BANK_0_HI {
		//fmt.Printf("BANK0 write:%04x:%04x\n",addr&0xfff,val)
		m.ram[addr&0xfff] = val
    }else if addr  >= EBANK_LO && addr < EBANK_HI {
		offset :=(addr&0xfff) +(0x1000 *uint16(m.SVBK-1) )
		m.exp_ram[offset] = val
	}else if addr >= ECHO_B0_LO && addr < ECHO_B0_HI {
		m.ram[(addr-0x2000)&0xfff] = val
	}else if addr >= ECHO_B1_LO && addr < ECHO_B1_HI {
		new_addr := addr-0x3000 
		offset :=(new_addr&0xfff) +(0x1000 *uint16(m.SVBK-1) )
		m.exp_ram[offset] = val
	} else if addr >= Z_RAM_LO && addr < Z_RAM_HI {
		offset:=(addr&0x00ff)-0x80
	//	fmt.Printf("%x\n",offset)
		m.z_ram[offset] = val
	} else {
		fmt.Printf("unhandled write:%04x:%04x\n", addr, val)
	}
}


