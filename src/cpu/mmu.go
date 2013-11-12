package cpu

import (
	"carts"
	"fmt"
    "component"
)
type mmio_connection struct {
	comp component.MMIOComponent
	addr uint16
	name string

}
const (
	MAX_MMIO = 256
)
type MMU struct {
	cart   carts.Cart
	cpu    *CPU
	inbios bool
	HDMA_hi_src uint8
	HDMA_lo_src uint8
	HDMA_hi_dst uint8
	HDMA_lo_dst uint8
    HDMA_start uint8
	mmio_connections [MAX_MMIO]*mmio_connection
}


func NewMMU(cpu *CPU) *MMU {
	m := new(MMU)
	m.inbios = false
	m.cpu = cpu
	return m
}

func (m *MMU) Create_new_cart(filename string) {
	m.cart = carts.Load_cart(filename)
}

func (m *MMU) Connect_mmio(addr uint16,name string,comp component.MMIOComponent) {
	if addr < 0xff00 || m.mmio_connections[addr&0xff] != nil {
		panic("Unable to handle this mmio connection")
	}
	
	con := new(mmio_connection)
	con.addr = addr
	con.name = name
	con.comp = comp
	
	m.mmio_connections[addr &0xff] = con
	
}

func (m *MMU) write_mmio(addr uint16, val uint8) {
	con := m.mmio_connections[addr &0xff]
	if con != nil {
		con.comp.Write_mmio(addr,val)
		return
	}
	
	fmt.Printf("unhandled write:%04x:%04x\n", addr, val)
}

func (m *MMU) read_mmio(addr uint16) uint8 {
	var val uint8 = 0
	con := m.mmio_connections[addr &0xff]
	if con != nil {
		return con.comp.Read_mmio(addr)
	}
	fmt.Printf("unhandled read:%04x\n", addr)
	return val
}

func (m *MMU) write_b(addr uint16, val uint8) {

	if addr < 0x8000 {
		m.cart.Write_b(addr, val)
	} else if addr < 0xA000 {	
		m.cpu.gpu.Vram.Write_b(addr,val)
	} else if addr < 0xC000 {
		m.cart.Write_b(addr, val)
	} else if addr < 0xfe00 {
		m.cpu.dram.Write_b(addr,val)
	} else if addr >= 0xff30 && addr < 0xff40 {
		//fmt.Println(m.cpu.sound.Wram,(addr&0x00ff) - 0x30)
		m.cpu.sound.Wram[(addr&0x00ff)-0x30] = val
	}else if (addr >= 0xff10 && addr < 0xff27) {
		m.cpu.sound.Write_mmio(addr,val)
	}else if (addr >= 0xff40 && addr < 0xff46)  || addr >= 0xff47 && addr < 0xff4C || addr == 0xff4f || addr >= 0xff68 && addr < 0xff6C{
		m.cpu.gpu.Write_mmio(addr,val)
	} else if addr <= 0xfe9f {
		m.cpu.gpu.Oam[addr&0x00ff] = val
		
	} else if addr >= 0xff00 && addr <= 0xff70 || addr == 0xffff {
		m.write_mmio(addr, val)
	} else if addr >= 0xff80 {
		 m.cpu.dram.Write_b(addr,val)

	} else {
		fmt.Printf("MMU unhandled write:%04x:%04x\n", addr, val)

	}

}
func (m *MMU) read_b(addr uint16) uint8 {

	//   fmt.Printf("write:%04x:%04x\n",addr,val)
	var val uint8
	if addr < 0x8000 {
		val = m.cart.Read_b(addr)
	} else if addr < 0xA000 {
		val = m.cpu.gpu.Vram.Read_b(addr)
	} else if addr < 0xC000 {
		val = m.cart.Read_b(addr)
	} else if addr < 0xfe00 {
		val = m.cpu.dram.Read_b(addr)
	} else if addr >= 0xfe00 && addr <= 0xfe9f {
		val = m.cpu.gpu.Oam[addr&0x00ff]
	}else if (addr >= 0xff40 && addr < 0xff46)  {
	    val = m.cpu.gpu.Read_mmio(addr)
	}else if (addr >= 0xff10 && addr < 0xff27)  {
	    val = m.cpu.sound.Read_mmio(addr)
	} else if addr >= 0xff30&& addr < 0xff40 {
		val = m.cpu.sound.Wram[(addr&0x00ff)-0x30]
	} else if addr >= 0xff40 && addr < 0xff46  || addr >= 0xff47 && addr < 0xff4C  || addr == 0xff4f || addr >= 0xff68 && addr < 0xff6 {
		val = m.cpu.gpu.Read_mmio(addr)
	} else if addr >= 0xff00 && addr <= 0xff70 || addr == 0xffff {
		val = m.read_mmio(addr)
	} else if addr >= 0xff80 {
		val = m.cpu.dram.Read_b(addr,)
	} else {
		fmt.Printf("unhandled read:%04x:%04x\n", addr, val)
        //panic("Fail")
	}
	return val
}


func (m *MMU) read_w(addr uint16) uint16 {
	return uint16(m.read_b(addr)) | uint16((m.read_b(addr+1)))<<8
}
func (m *MMU) Read (addr uint16) uint8 {
	return m.read_b(addr)
}
func (m *MMU) Write (addr uint16,val uint8) {
	m.write_b(addr,val)
}
func (m *MMU) write_w(addr uint16, val uint16) {

	m.write_b(addr, uint8(val&0x00ff))
	m.write_b(addr+1, uint8((val&0xff00)>>8))

}
