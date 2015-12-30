package mmu

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
	inbios bool
	mmio_connections [MAX_MMIO]*mmio_connection
	range_connections [MAX_MMIO]*range_connection
	range_count uint8
}


func NewMMU() *MMU {
	m := new(MMU)
	m.inbios = false
	return m
}
func (m *MMU) Reset () {

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

func (m *MMU) Write_b(addr uint16, val uint8) {
	
	if addr >= 0xff00 && addr <= 0xff70 || addr == 0xffff {
		m.write_mmio(addr, val)
		return
	}
    //m.Print_map()
	var i uint8
	for i=0; i<m.range_count; i++ {
		if addr >= m.range_connections[i].addr_lo &&
			addr < m.range_connections[i].addr_hi {
			con := m.range_connections[i]
			con.comp.Write(addr,val)
			//fmt.Printf("%v:Writing %s %x:%x \n",1,con.name,addr,val)
			return 

		} 
	}
}

func (m *MMU) Print_map() {
    fmt.Printf("===Address Map===\n")
    var i uint8
    for i=0; i<m.range_count; i++ {
        fmt.Printf("%v:%x-%x\n",m.range_connections[i].name, m.range_connections[i].addr_lo,m.range_connections[i].addr_hi)
    }



}
func (m *MMU) Read_b(addr uint16) uint8 {

	if addr >= 0xff00 && addr <= 0xff70 || addr == 0xffff {
		return m.read_mmio(addr)
	}
    


	var i uint8
	for i=0; i<m.range_count; i++ {
		if addr >= m.range_connections[i].addr_lo &&
			addr < m.range_connections[i].addr_hi {
			con := m.range_connections[i]
			val :=con.comp.Read(addr)
	//		fmt.Printf("%v:Reading %s %x:%x \n",m.cpu.clock.Cycles,con.name,addr,val)
			return val

		} 
	}
	return 0
}


func (m *MMU) Read_w(addr uint16) uint16 {
	return uint16(m.Read_b(addr)) | uint16((m.Read_b(addr+1)))<<8
}
func (m *MMU) Read (addr uint16) uint8 {
	return m.Read_b(addr)
}
func (m *MMU) Write (addr uint16,val uint8) {
	m.Write_b(addr,val)
}
func (m *MMU) Write_w(addr uint16, val uint16) {

	m.Write_b(addr, uint8(val&0x00ff))
	m.Write_b(addr+1, uint8((val&0xff00)>>8))


}
