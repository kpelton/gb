package cpu

import "fmt"

type MMU struct {
	mem [0x10000]uint8
	//    cart [0x10000]uint8
	cart   Cart
	vm     [0x2000]uint8
	oam    [0xA0]uint8
	cpu    *CPU
	block  uint16
	inbios bool
}

func NewMMU(cpu *CPU) *MMU {
	m := new(MMU)
	m.inbios = false
	m.cpu = cpu
	m.block = 0
	return m
}

const (
	REG_CART_TYPE = 0x147
	REG_CART_SIZE = 0x148
	REG_RAM_SIZE  = 0x148

	C_ROM_ONLY          = 0
	C_ROM_MBC1          = 1
	C_ROM_MBC1_RAM      = 2
	C_ROM_MBC1_RAM_BATT = 3
	C_ROM_MBC2          = 5
	C_ROM_MBC2_BATT     = 6
	C_ROM_MBC3_RAM_BATT = 13
	C_ROM_RAM           = 8
)

func (m *MMU) create_new_cart(data []uint8, size int) {
	fmt.Printf("Cart Type:0%02x\n:", data[REG_CART_TYPE])
	fmt.Printf("Cart Size:0%02x:\n", data[REG_CART_SIZE])
	fmt.Printf("Ram Size:0%02x:\n", data[REG_RAM_SIZE])

	cart_type := data[REG_CART_TYPE]
	fmt.Printf("Cart Type:")
	switch cart_type {
	case C_ROM_ONLY:
		fmt.Printf("ROM_ONLY\n")
		m.cart = NewMBC0(data[:0x8000])
	case C_ROM_MBC1:
		fmt.Printf("ROM_MBC1\n")
		m.cart = NewROM_MBC1(data, size)

	case C_ROM_MBC1_RAM:
		fmt.Printf("ROM_MBC1_RAM\n")
		m.cart = NewROM_MBC1(data, size)

	case C_ROM_MBC1_RAM_BATT:
		m.cart = NewROM_MBC1(data, size)

		fmt.Printf("ROM_MBC1_RAM_BATT\n")
	case C_ROM_MBC2:
		fmt.Printf("ROM_MBC2\n")
	case C_ROM_MBC2_BATT:
		m.cart = NewROM_MBC2(data, size)
	case C_ROM_MBC3_RAM_BATT:
		m.cart = NewROM_MBC1(data, size)

	case C_ROM_RAM:
		fmt.Printf("ROM_RAM\n")
	default:

		fmt.Printf("Unknown!\n")
		//  panic("Unsupported cart!!!!")
		m.cart = NewROM_MBC1(data, size)

	}

}

func (m *MMU) Dump_mem() {
	j := 0
	fmt.Printf("\n0x0000:")
	for i := 0x8000; i < 0xafff; i++ {
		fmt.Printf("%02X ", m.vm[i])
		j++
		if j == 16 {
			fmt.Printf("\n0x%04X:", i+1+0x0000)
			j = 0
		}
	}

}
func (m *MMU) Dump_vm() {
	j := 0
	for i := 0x0; i < 0x2000; i++ {
		fmt.Printf("%02X ", m.vm[i])
		j++
		if j == 16 {
			fmt.Printf("\n0x%04X:", i+1+0x8000)
			j = 0
		}
	}
}

func (m *MMU) exec_dma(addr uint8) {
	var real_addr uint16
	var i uint16
	real_addr = uint16(addr) * 0x100
	for i = 0; i < 160; i++ {
		m.oam[i] = m.read_b(real_addr + i)

	}

}

func (m *MMU) write_mmio(addr uint16, val uint8) {
	switch addr {
	case 0xff00:
		m.cpu.gp.P1 = val
		//		fmt.Printf("->P1:%04X\n",val)
	case 0xff02:
		fmt.Printf("->SERIAL:%04X\n", val)
		if ((m.mem[0xff02] & 0x80) == 0x80) && ((m.mem[0xff02] & 0x1) == 0x1) {

			m.cpu.ic.Assert(SERIAL)
			m.mem[0xff01] = 0xff
			m.mem[0xff02] = val & (^uint8(0x80))

		} else {
			m.mem[0xff02] |= val

		}

	case 0xff05:
		m.cpu.timer.TIMA = val
	case 0xff06:
		m.cpu.timer.TMA = val
	case 0xff07:
		m.cpu.timer.TAC = val
	case 0xff40:
		m.cpu.gpu.LCDC = val
		//fmt.Printf("VAL:%04X\n",val)
		fmt.Printf("->LCDC:%04X\n", val)
		m.cpu.gpu.mem_written = true
	case 0xff41:
		m.cpu.gpu.STAT |= val & 0xf8
		//m.cpu.Dump()
		fmt.Printf("->STAT:%04X\n", m.cpu.gpu.STAT)

	case 0xff42:
		m.cpu.gpu.SCY = val
	case 0xff43:
		//fmt.Printf("->SCX:%04X\n",val)
		//m.cpu.Dump()
		m.cpu.gpu.SCX = val
	case 0xff44:
		m.cpu.gpu.LY = 0
		//fmt.Printf("->LY:%04X\n",val)
	case 0xff45:
		m.cpu.gpu.LYC = val
		//fmt.Printf("->LYC:%04X\n",val)
	case 0xff46:
		// m.Dump_vm()
		m.exec_dma(val)
	case 0xff47:
		if val != m.cpu.gpu.BGP {
			m.cpu.gpu.update_palette(&m.cpu.gpu.bg_palette, val)
			m.cpu.gpu.BGP = val
			fmt.Println("BGP", val, m.cpu.gpu.obp0_palette)

		}
	case 0xff48:
		if val != m.cpu.gpu.OBP0 {
			m.cpu.gpu.OBP0 = val
			m.cpu.gpu.update_palette(&m.cpu.gpu.obp0_palette, val)
			fmt.Println(m.cpu.gpu.obp0_palette)

		}
	case 0xff49:
		if val != m.cpu.gpu.OBP1 {
			m.cpu.gpu.OBP1 = val
			m.cpu.gpu.update_palette(&m.cpu.gpu.obp1_palette, val)
		}
	case 0xff4A:
		m.cpu.gpu.WY = val
	case 0xff4B:
        fmt.Printf("->WX:%04X\n", val)
		m.cpu.gpu.WX = val
	case 0xffff:
		m.cpu.ic.IE = val
		//fmt.Printf("->IE:%04X\n", val)
	case 0xff0F:
		//`fmt.Printf("->IF:%04X\n", val)
		m.cpu.ic.IF = val
	}

}
func (m *MMU) read_mmio(addr uint16) uint8 {
	var val uint8 = 0
	switch addr {
	case 0xff00:
		val = m.cpu.gp.P1
	//fmt.Printf("<-P1:%04X\n",val)
	case 0xff04:
		val = m.cpu.DIV
	case 0xff05:
		val = m.cpu.timer.TIMA
	case 0xff06:
		val = m.cpu.timer.TMA
	case 0xff07:
		val = m.cpu.timer.TAC
	case 0xff40:
		val = m.cpu.gpu.LCDC
	case 0xff41:
		val = m.cpu.gpu.STAT
		//fmt.Printf("<-STAT:%04X\n", val)
	case 0xff42:
		val = m.cpu.gpu.SCY
	case 0xff43:
		val = m.cpu.gpu.SCX
	case 0xff44:
		val = m.cpu.gpu.LY
	case 0xff45:
		val = m.cpu.gpu.LYC
	case 0xff46:
		val = 0xff
	case 0xff47:
		val = m.cpu.gpu.BGP
	case 0xff48:
		val = m.cpu.gpu.OBP0
	case 0xff49:
		val = m.cpu.gpu.OBP1
	case 0xff4A:
		val = m.cpu.gpu.WY
	case 0xff4B:
		val = m.cpu.gpu.WX
	case 0xffff:
		val = m.cpu.ic.IE
	case 0xff0F:
		val = m.cpu.ic.IF
	}

	return val
}
func (m *MMU) read_b(addr uint16) uint8 {

	if addr >= 0x8000 && addr < 0xa000 {
		return m.vm[addr&0x1fff]

	} else if addr >= 0x100 && addr < 0x4000 {
		//always ROM bank #0
		return m.cart.Read_b(addr)

	} else if addr >= 0x4000 && addr < 0x8000 || addr >= 0xA000 && addr < 0xC000 {
		//  fmt.Printf("Bank:0x%X,Addr:0x%x,Cart:0x%X\n",m.bank,addr,uint32(addr) +(uint32(m.bank) * 0x4000) )
		//return m.cart[uint32(addr) +(uint32(m.bank) * 0x4000) ]
		return m.cart.Read_b(addr)
		//return m.cart[addr]

	} else if addr <= 0x100 && !m.inbios {
		return m.cart.Read_b(addr)

	} else if addr >= 0xfe00 && addr <= 0xfe9f {
		fmt.Printf("%x\n", addr)
		return m.oam[addr&0x00ff]

	} else if addr == 0xff00 || (addr >= 0xff04 && addr <= 0xff07) || (addr >= 0xff40 && addr <= 0xff4B) || addr == 0xff0f || addr == 0xffff {
		return m.read_mmio(addr)

	} else if addr >= 0xe000 && addr < 0xfe00 {
		return m.mem[addr-0x1000]
	}
	return m.mem[addr]

}

func (m *MMU) read_w(addr uint16) uint16 {
	return uint16(m.read_b(addr)) | uint16((m.read_b(addr+1)))<<8
}

func (m *MMU) write_b(addr uint16, val uint8) {

	//   fmt.Printf("write:%04x:%04x\n",addr,val)

	if addr >= 0x8000 && addr < 0xA000 {
		m.vm[addr&0x1fff] = val
		//m.cpu.Dump()
		//fmt.Printf("Video:0x%04X->0x%02X\n",addr,val) 
		//m.cpu.gpu.mem_written = true
		m.vm[addr&0x1fff] = val
		return

	} else if addr < 0x8000 || addr >= 0xA000 && addr < 0xC000 {
		m.cart.Write_b(addr, val)
		return
	} else if addr == 0xff00 || addr == 0xff02 || (addr >= 0xff04 && addr <= 0xff07) || (addr >= 0xff40 && addr <= 0xff4B) || addr == 0xff0f || addr == 0xffff {
		m.write_mmio(addr, val)
		return
	} else if addr >= 0xfe00 && addr <= 0xfe9f {
		m.oam[addr&0x00ff] = val
		return
	} else if addr >= 0xe000 && addr < 0xfe00 {
		m.mem[addr-0x1000] = val
		fmt.Println("shadow")
		return

	}

	m.mem[addr] = val

}

func (m *MMU) write_w(addr uint16, val uint16) {

	m.write_b(addr, uint8(val&0x00ff))
	m.write_b(addr+1, uint8((val&0xff00)>>8))

}
