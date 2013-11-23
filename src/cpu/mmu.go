package cpu

import (
	"carts"
	//"fmt"
    "component"
)
type mmio_connection struct {
	comp component.MMIOComponent
	addr uint16
	name string

}

type range_connection struct {
	comp component.MemComponent
	name string
	addr_hi uint16
	addr_lo uint16
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
	range_connections [MAX_MMIO]*range_connection
	range_count uint8
}


func NewMMU(cpu *CPU) *MMU {
	m := new(MMU)
	m.inbios = false
	m.cpu = cpu
	return m
}

func (m *MMU) Create_new_cart(filename string) {
	m.cart = carts.Load_cart(filename)
	range_list := m.cart.Get_range_list()
	for i := range range_list {
		m.Connect_range(range_list[i],m.cart) 
	}
}

func (m *MMU) Get_range_list() component.RangeList {
	return component.RangeList{}
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

func (m *MMU) Connect_range(r component.Range,comp component.MemComponent) {
	if m.range_count == 0xff  {
		panic("Unable to handle this mem range connection")
	}
	con := new(range_connection)
	con.name = r.Name
	con.addr_lo = r.Addr_lo
	con.addr_hi = r.Addr_hi
	con.comp = comp
	m.range_connections[m.range_count] = con
	m.range_count+=1
}


func (m *MMU) write_mmio(addr uint16, val uint8) {
	con := m.mmio_connections[addr &0xff]
	if con != nil {
		con.comp.Write_mmio(addr,val)
	//	fmt.Printf("%v:Writing %s %x\n",m.cpu.clock.Cycles,con.name,val) 
	

	
		return
	}
	
	//fmt.Printf("unhandled write:%04x:%04x\n", addr, val)
} 

func (m *MMU) read_mmio(addr uint16) uint8 {
	var val uint8 = 0
	con := m.mmio_connections[addr &0xff]
	if con != nil {
		
		val :=con.comp.Read_mmio(addr)
	//	fmt.Printf("%v:Reading %s %x \n",m.cpu.clock.Cycles,con.name,val) 
	//	m.cpu.Dump()
		return val
	}
	//fmt.Printf("unhandled read:%04x\n", addr)
	return val
}

func (m *MMU) write_b(addr uint16, val uint8) {


	
	if addr >= 0xff00 && addr <= 0xff70 || addr == 0xffff {
		m.write_mmio(addr, val)
		return
	}
	var i uint8
	for i=0; i<m.range_count; i++ {
		if addr >= m.range_connections[i].addr_lo &&
			addr < m.range_connections[i].addr_hi {
			con := m.range_connections[i]
			con.comp.Write(addr,val)
			//fmt.Printf("%v:Writing %s %x:%x \n",m.cpu.clock.Cycles,con.name,addr,val)
			return 

		} 
	}

	//if addr >= 0xff30 && addr < 0xff40 {
		//fmt.Println(m.cpu.sound.Wram,(addr&0x00ff) - 0x30)
		//m.cpu.sound.Wram[(addr&0x00ff)-0x30] = val
//	if (addr >= 0xff10 && addr < 0xff27) {
//		m.cpu.sound.Write_mmio(addr,val)
//	} else 
//	} else {
	//	fmt.Printf("MMU unhandled write:%04x:%04x\n", addr, val)

//	}

}


func (m *MMU) read_b(addr uint16) uint8 {

	//   fmt.Printf("write:%04x:%04x\n",addr,val)
	if addr >= 0xff00 && addr <= 0xff70 || addr == 0xffff {
		return m.read_mmio(addr)
	}
	var i uint8
	for i=0; i<m.range_count; i++ {
		if addr >= m.range_connections[i].addr_lo &&
			addr < m.range_connections[i].addr_hi {
			con := m.range_connections[i]
			val :=con.comp.Read(addr)
			//fmt.Printf("%v:Reading %s %x:%x \n",m.cpu.clock.Cycles,con.name,addr,val)
			return val

		} 
	}

//	var val uint8
//	if (addr >= 0xff10 && addr < 0xff27)  {
//	    val = m.cpu.sound.Read_mmio(addr)
//	} else {
//	fmt.Printf("unhandled read:%04x\n", addr)
        //panic("Fail")
//	}
	return 0
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
