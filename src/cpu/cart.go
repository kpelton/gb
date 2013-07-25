package cpu
import (
	"fmt"
)

type Cart interface {
    Read_b(uint16) uint8
    Write_b(uint16,uint8)

}
/////MBC0///////
type MBC0 struct {
    cart [0x8000]uint8 
}

func NewMBC0(cart_data []uint8 )(*MBC0) {
    m :=new(MBC0)
    copy(m.cart[:],cart_data)

	return m
}

func (m * MBC0 )Read_b(addr uint16) (uint8) {
    return m.cart[addr]
}

func (m * MBC0 )Write_b(addr uint16,val uint8)() {
    fmt.Printf("WRITE TO ROM FAILED!!\n")
}
/////MBC1///////
type ROM_MBC1 struct {
    cart [0x200000]uint8
    bank uint16
    ram_enabled  bool 
    ram_bank uint8
    ram [0x8000]uint8
   
}
func NewROM_MBC1(cart_data []uint8,size int)(*ROM_MBC1) {
    m :=new(ROM_MBC1)
    fmt.Println(size)
    copy(m.cart[:],cart_data)
	return m
}

func (m * ROM_MBC1)Read_b(addr uint16) (uint8) {    
    var retval uint8
    
    if addr < 0x4000  {
       retval = m.cart[addr]
    }else if addr >= 0x4000 && addr < 0x8000{
       retval = m.cart[uint32(addr) +(uint32(m.bank) * 0x4000) ]
    }else if addr >= 0xA000 && addr < 0xc000{
       retval = m.ram[uint32(addr& 0x1fff) +(uint32(m.ram_bank) * 0x2000) ]
    }

    return retval
}

func (m * ROM_MBC1 )Write_b(addr uint16,val uint8)() {
    if addr >= 0x2000 && addr < 0x4000  {
        if (val >1){
            //fmt.Println("Bank from",m.bank,val-1)
            m.bank = uint16(val-1)  
            }else{
                m.bank = uint16(0)
            }
    } else if addr < 0x2000 {

        m.ram_enabled = true
        //fmt.Println("RAM enabled",val )

    }else if addr >= 0x4000 && addr < 0x6000{
        m.ram_bank = val
        fmt.Println("RAM_BANK",val)
    }else if addr >= 0xA000 && addr < 0xc000{
      m.ram[uint32(addr& 0x1fff) +(uint32(m.ram_bank) * 0x2000)] = val
      fmt.Println("RAM  BANK", m.ram_bank,val)
    }

}

/////MBC2///////
type ROM_MBC2 struct {
    cart [0x200000]uint8
    bank uint16
}
func NewROM_MBC2(cart_data []uint8,size int)(*ROM_MBC2) {
    m :=new(ROM_MBC2)
    fmt.Println(size)
    copy(m.cart[:],cart_data)
	return m
}

func (m * ROM_MBC2)Read_b(addr uint16) (uint8) {    
    var retval uint8
    
    if addr < 0x4000  {
        //always ROM bank #0
       retval = m.cart[addr]
    }else if addr <0xc000{
       retval = m.cart[uint32(addr) +(uint32(m.bank) * 0x4000)]
    }
    return retval
}

func (m * ROM_MBC2 )Write_b(addr uint16,val uint8)() {
    if addr > 0x2000 && addr < 0x4000 {
        if (val >1){
            //fmt.Println("Bank from",m.bank,val-1)
            m.bank = uint16(val-1)  
            }else{
                m.bank = uint16(0)
            }
    }
}

