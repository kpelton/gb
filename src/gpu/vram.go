package gpu

import (
	"fmt"

)

type VRAM struct {
	Vm        [0x4000]uint8
	Oam          [0xA0] uint8
	VBANK      uint8
}
const (
	VRAM_LO = 0x8000
	VRAM_HI = 0xC000
	VBANK_MMIO = 0xff4f

)
	

func newVRAM() *VRAM {
	m := new(VRAM)
	m.Reset()
	return m
}

func (m *VRAM)  Reset() {
	m.VBANK = 0
}




func (m *VRAM) Write_mmio(addr uint16,val uint8)  {
	if addr == VBANK_MMIO {
		m.VBANK = val & 0x1
		//fmt.Printf("VBANK:%x\n",m.VBANK)
	} else { 
		panic("VRAM:unhandled VRAM mmio write")
	}

}

func (m *VRAM) Read_mmio(addr uint16) uint8  {
	if addr == VBANK_MMIO {
		return m.VBANK 

	} else { 
		panic("VRAM:unhandled VRAM mmio read")
	}


}


func (m *VRAM) Read_b(addr uint16) uint8 {

	var val uint8

	if addr >= VRAM_LO && addr < VRAM_HI {
		offset := (uint16(m.VBANK) *0x2000) +addr &0x1fff
		val = m.Vm[offset]
    }else if addr >= 0xfe00 && addr <= 0xfe9f {
		val = m.Oam[addr&0x00ff]
	//	fmt.Printf("oam read:%04x:%04x\n",addr,val)
	} else {
		fmt.Printf("VRAM:unhandled read:%04x:%04x\n", addr, val)
	}
	
	return val
}

func (m *VRAM) Write_b(addr uint16,val uint8)  {

	if addr >= VRAM_LO && addr < VRAM_HI {
		offset := (uint16(m.VBANK) *0x2000) +addr &0x1fff
		m.Vm[offset] = val
		//fmt.Printf("%v:VRAM::%04x:%04x\n",m.VBANK, addr, val)
	} else if addr >= 0xfe00 && addr <= 0xfe9f {
		m.Oam[addr&0x00ff] = val

	//	fmt.Printf("oam write:%04x:%04x\n",addr,val)
	} else {

		fmt.Printf("VRAM:unhandled write:%04x:%04x\n", addr, val)
	}

}


