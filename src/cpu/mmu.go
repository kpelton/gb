package cpu

import (
	"carts"
	"fmt"

)

type MMU struct {
	cart   carts.Cart
	cpu    *CPU
	inbios bool
    KEY1 uint8
	HDMA_hi_src uint8
	HDMA_lo_src uint8
	HDMA_hi_dst uint8
	HDMA_lo_dst uint8
    HDMA_start uint8
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


func (m *MMU) exec_dma(addr uint8) {
	var real_addr uint16
	var i uint16
	real_addr = uint16(addr) * 0x100
	for i = 0; i < 160; i++ {
		m.cpu.gpu.Oam[i] = m.read_b(real_addr + i)

	}

}

func (m *MMU) write_mmio(addr uint16, val uint8) {
	switch addr {
	case 0xff00:

		m.cpu.gp.P1 = val
		//		fmt.Printf("->P1:%04X\n",val)
	case 0xff01:
       // fmt.Printf("->SERIALB:%04X\n", val)

        m.cpu.serial.Write(addr,val)        
	case 0xff02:
		//fmt.Printf("->SERIALC:%04X\n", val)
         m.cpu.serial.Write(addr,val)        


	case 0xff0F:
		//`fmt.Printf("->IF:%04X\n", val)
		m.cpu.ic.IF = val


	case 0xff46:
		// m.Dump_vm()
		m.exec_dma(val)

	case 0xff4D:
		fmt.Printf("->KEY1:%04X\n", val &0x1)
        m.KEY1= val &0x1
        m.cpu.Ready_sswitch()

	case 0xff51:
		fmt.Printf("->SRC:HDMA_HIGH:%04X\n", val)
        m.HDMA_hi_src = val
	case 0xff52:
		fmt.Printf("->SRC:HDMA_LOW:%04X\n", val)
        m.HDMA_lo_src = val
	case 0xff53:
		fmt.Printf("->DST:HDMA_HIGH:%04X\n", val)
        m.HDMA_hi_dst = val
	case 0xff54:
		fmt.Printf("->DST:HDMA_LOW:%04X\n", val)
        m.HDMA_lo_dst = val
	case 0xff55:
		src := uint16(m.HDMA_hi_src) <<8 | uint16(m.HDMA_lo_src)
		dst := uint16(m.HDMA_hi_dst) <<8 | uint16(m.HDMA_lo_dst)
		if dst < 0x8000 {
			dst+=0x8000
		}
		fmt.Printf("0x%x\n",val)
		length := (uint16( (val & 0x7f)) + 1) *0x10

		fmt.Printf("->START transfer:%04X %04x %x\n", src,dst,length)

		var i uint16

		for i = 0; i < uint16(length); i++ {
			m.write_b(dst +i,m.read_b(src+i))
		}
    case 0xff70:
		m.cpu.dram.Write_mmio(addr,val)
	case 0xffff:
		m.cpu.ic.IE = val
		//fmt.Printf("->IE:%04X\n", val)
	default:
		fmt.Printf("unhandled write:%04x:%04x\n", addr, val)

	}

}

func (m *MMU) read_mmio(addr uint16) uint8 {
	var val uint8 = 0
	switch addr {

	case 0xff00:
		m.cpu.gp.Update()
		val = m.cpu.gp.P1
	//fmt.Printf("<-P1:%04X\n",val)
    case 0xff01:
        val= m.cpu.serial.Read(addr)     
         fmt.Printf("<-SERIALB:%04X\n", val)
	case 0xff02:
        val = m.cpu.serial.Read(addr)        
		fmt.Printf("<-SERIALC:%04X\n", val)
	case 0xff04:
		val = m.cpu.DIV
		//fmt.Printf("<-DIV:%04X\n",val)
	case 0xff46:
		val = 0xff
    case 0xff4D:
        fmt.Printf("<-KEY1:%04X\n", m.KEY1)
        val = m.KEY1
	case 0xff55:
		val = m.HDMA_start
		m.HDMA_start = 0
	case 0xff70:
		val =m.cpu.dram.Read_mmio(addr)
	case 0xffff:
		val = m.cpu.ic.IE
	case 0xff0F:
		val = m.cpu.ic.IF
    default:
		fmt.Printf("unhandled read:%04x\n", addr)


	}

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
	}else if (addr >= 0xff05 && addr < 0xff08) {
		m.cpu.timer.Write_mmio(addr,val)
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
	}else if (addr >= 0xff05 && addr < 0xff08) {
		val = m.cpu.timer.Read_mmio(addr)
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
func (m *MMU) write_w(addr uint16, val uint16) {

	m.write_b(addr, uint8(val&0x00ff))
	m.write_b(addr+1, uint8((val&0xff00)>>8))

}
