package cpu

import (
	"carts"
	"constants"
	"fmt"
)

type MMU struct {
	mem [0x10000]uint8
	//      cart [0x10000]uint8
	cart   carts.Cart
	vm     [0x2000]uint8
	ram    [0x2000]uint8
	z_ram  [0x80]uint8
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

func (m *MMU) Create_new_cart(filename string) {
	m.cart = carts.Load_cart(filename)
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
		//fmt.Printf("->SERIAL:%04X\n", val)
		if ((m.mem[0xff02] & 0x80) == 0x80) && ((m.mem[0xff02] & 0x1) == 0x1) {

			m.cpu.ic.Assert(constants.SERIAL)
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
	case 0xff0F:
		//`fmt.Printf("->IF:%04X\n", val)
		m.cpu.ic.IF = val
	case 0xff10:
		m.cpu.sound.SND_MODE_1_SWP = val
	case 0xff11:
		m.cpu.sound.SND_MODE_1_LEN = val
	case 0xff12:
		m.cpu.sound.SND_MODE_1_ENVP = val
	case 0xff13:
		m.cpu.sound.SND_MODE_1_FREQ_LOW = val
	case 0xff14:
		m.cpu.sound.SND_MODE_1_FREQ_HI = val

	case 0xff16:
		m.cpu.sound.SND_MODE_2_LEN = val
	case 0xff17:
		m.cpu.sound.SND_MODE_2_ENVP = val
	case 0xff18:
		m.cpu.sound.SND_MODE_2_FREQ_LOW = val
	case 0xff19:
		m.cpu.sound.SND_MODE_2_FREQ_HI = val

	case 0xff1a:
		m.cpu.sound.SND_MODE_3 = val
	case 0xff1b:
		m.cpu.sound.SND_MODE_3_LEN = val
		//fmt.Println(val)
	case 0xff1c:
		m.cpu.sound.SND_MODE_3_OUTPUT = val
	case 0xff1d:
		m.cpu.sound.SND_MODE_3_FREQ_HI = val
	case 0xff1e:
		m.cpu.sound.SND_MODE_3_FREQ_HI = val

	case 0xff20:
		m.cpu.sound.SND_MODE_4_LEN = val
	case 0xff21:
		m.cpu.sound.SND_MODE_4_ENVP = val
	case 0xff22:
		m.cpu.sound.SND_MODE_4_POLY = val
	case 0xff23:
		m.cpu.sound.SND_MODE_4_COUNTER = val

	case 0xff24:
		m.cpu.sound.SND_CHN_CTRL = val
	case 0xff25:
		m.cpu.sound.SND_TERM_OUTPUT = val
	case 0xff26:
		m.cpu.sound.SND_MASTER_CTRL = val

	case 0xff40:
		m.cpu.gpu.LCDC = val
		//fmt.Printf("VAL:%04X\n",val)
		//m.cpu.Dump()
		//fmt.Printf("->LCDC:%04X,LY:0x%04X\n", val,m.cpu.gpu.LY)
		m.cpu.gpu.mem_written = true
	case 0xff41:
		m.cpu.gpu.STAT |= val & 0xf8
		//m.cpu.Dump()
		//fmt.Printf("->STAT:%04X %04X\n", m.cpu.gpu.STAT, val)

	case 0xff42:
		m.cpu.gpu.SCY = val
	case 0xff43:
		//fmt.Printf("->SCX:%04X\n",val)x
		//m.cpu.Dump()
		m.cpu.gpu.SCX = val
	case 0xff44:
		m.cpu.gpu.LY = 0
		//fmt.Printf("->LY:%04X\n",val)
	case 0xff45:
		//		m.cpu.Dump()

		m.cpu.gpu.LYC = val
		//fmt.Printf("->LYC:%04X %04X \n",val,m.cpu.gpu.cycle_count)
	case 0xff46:
		// m.Dump_vm()
		m.exec_dma(val)
	case 0xff47:
		if val != m.cpu.gpu.BGP {
			m.cpu.gpu.update_palette(&m.cpu.gpu.bg_palette, val)
			m.cpu.gpu.BGP = val
			//fmt.Println("BGP", val, m.cpu.gpu.obp0_palette)

		}
	case 0xff48:
		if val != m.cpu.gpu.OBP0 {
			m.cpu.gpu.OBP0 = val
			m.cpu.gpu.update_palette(&m.cpu.gpu.obp0_palette, val)
			//fmt.Println(m.cpu.gpu.obp0_palette)

		}
	case 0xff49:
		if val != m.cpu.gpu.OBP1 {
			m.cpu.gpu.OBP1 = val
			m.cpu.gpu.update_palette(&m.cpu.gpu.obp1_palette, val)
		}
	case 0xff4A:
		m.cpu.gpu.WY = val
		fmt.Printf("->WY:%04X\n", val)

	case 0xff4B:
		fmt.Printf("->WX:%04X\n", val)
		m.cpu.gpu.WX = val
	case 0xffff:
		m.cpu.ic.IE = val
		//fmt.Printf("->IE:%04X\n", val)

	}

}

func (m *MMU) read_mmio(addr uint16) uint8 {
	var val uint8 = 0
	switch addr {

	case 0xff00:
		m.cpu.gp.Update()
		val = m.cpu.gp.P1
	//fmt.Printf("<-P1:%04X\n",val)
	case 0xff04:
		val = m.cpu.DIV
		//fmt.Printf("<-DIV:%04X\n",val)

	case 0xff05:
		val = m.cpu.timer.TIMA
	case 0xff06:
		val = m.cpu.timer.TMA
	case 0xff07:
		val = m.cpu.timer.TAC

	case 0xff10:
		val = m.cpu.sound.SND_MODE_1_SWP
	case 0xff11:
		val = m.cpu.sound.SND_MODE_1_LEN
	case 0xff12:
		val = m.cpu.sound.SND_MODE_1_ENVP
	case 0xff13:
		val = m.cpu.sound.SND_MODE_1_FREQ_LOW
	case 0xff14:
		val = m.cpu.sound.SND_MODE_1_FREQ_HI

	case 0xff16:
		val = m.cpu.sound.SND_MODE_2_LEN
	case 0xff17:
		val = m.cpu.sound.SND_MODE_2_ENVP
	case 0xff18:
		val = m.cpu.sound.SND_MODE_2_FREQ_LOW
	case 0xff19:
		val = m.cpu.sound.SND_MODE_2_FREQ_HI

	case 0xff1a:
		val = m.cpu.sound.SND_MODE_3
	case 0xff1b:
		val = m.cpu.sound.SND_MODE_3_LEN
	case 0xff1c:
		val = m.cpu.sound.SND_MODE_3_OUTPUT
	case 0xff1d:
		val = m.cpu.sound.SND_MODE_3_FREQ_HI
	case 0xff1e:
		val = m.cpu.sound.SND_MODE_3_FREQ_HI

	case 0xff20:
		val = m.cpu.sound.SND_MODE_4_LEN
	case 0xff21:
		val = m.cpu.sound.SND_MODE_4_ENVP
	case 0xff22:
		val = m.cpu.sound.SND_MODE_4_POLY
	case 0xff23:
		val = m.cpu.sound.SND_MODE_4_COUNTER

	case 0xff24:
		val = m.cpu.sound.SND_CHN_CTRL
	case 0xff25:
		val = m.cpu.sound.SND_TERM_OUTPUT
	case 0xff26:
		val = m.cpu.sound.SND_MASTER_CTRL

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

func (m *MMU) write_b(addr uint16, val uint8) {

	if addr < 0x8000 {
		m.cart.Write_b(addr, val)
	} else if addr < 0xA000 {
		m.vm[addr&0x1fff] = val
	} else if addr < 0xC000 {
		m.cart.Write_b(addr, val)
	} else if addr < 0xe000 {
		m.ram[addr&0x1fff] = val
	} else if addr < 0xfe00 {
		m.ram[(addr-0x2000)&0x1fff] = val
		fmt.Println("shadow")
	} else if addr >= 0xff30 && addr < 0xff40 {
		//fmt.Println(m.cpu.sound.Wram,(addr&0x00ff) - 0x30)
		m.cpu.sound.Wram[(addr&0x00ff)-0x30] = val
	} else if addr <= 0xfe9f {
		m.oam[addr&0x00ff] = val
	} else if addr >= 0xff00 && addr <= 0xff4b || addr == 0xffff {
		m.write_mmio(addr, val)
	} else if addr >= 0xff80 {
		m.z_ram[(addr&0xff)-0x80] = val
	} else {
		fmt.Printf("unhandled write:%04x:%04x\n", addr, val)

	}

}
func (m *MMU) read_b(addr uint16) uint8 {

	//   fmt.Printf("write:%04x:%04x\n",addr,val)
	var val uint8
	if addr < 0x8000 {
		val = m.cart.Read_b(addr)
	} else if addr < 0xA000 {
		val = m.vm[addr&0x1fff]
	} else if addr < 0xC000 {
		val = m.cart.Read_b(addr)
	} else if addr < 0xe000 {
		val = m.ram[addr&0x1fff]
	} else if addr < 0xfe00 {
		val = m.ram[(addr-0x2000)&0x1fff]
	} else if addr <= 0xfe9f {
		val = m.oam[addr&0x00ff]
	} else if addr >= 0xff30 && addr < 0xff40 {
		val = m.cpu.sound.Wram[(addr&0x00ff)-0x30]
	} else if addr >= 0xff00 && addr <= 0xff4b || addr == 0xffff {
		val = m.read_mmio(addr)
	} else if addr >= 0xff80 {
		val = m.z_ram[(addr&0x00ff)-0x80]
	} else {
		fmt.Printf("unhandled write:%04x:%04x\n", addr, val)
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
