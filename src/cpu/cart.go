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
const (

    SIXTEEN_MB = 0
    FOUR_MB =1
    


)
type ROM_MBC1 struct {
    cart [0x200000]uint8
    bank uint16
    ram_enabled  bool 
    ram_bank uint8
    ram [0x8000]uint8
    memory_mode uint8
   
}
func NewROM_MBC1(cart_data []uint8,size int)(*ROM_MBC1) {
    m :=new(ROM_MBC1)
    fmt.Println(size)
    copy(m.cart[:],cart_data)
    m.memory_mode = SIXTEEN_MB
	return m
}

func (m * ROM_MBC1)Read_b(addr uint16) (uint8) {    
    var retval uint8
    
    if addr < 0x4000  {
       retval = m.cart[addr]
    }else if addr >= 0x4000 && addr < 0x8000{
       retval = m.cart[uint32(addr) +(uint32(m.bank) * 0x4000) ]
    }else if addr >= 0xA000 && addr <= 0xc000{
       if (m.memory_mode == FOUR_MB && m.ram_enabled == true) || (m.memory_mode == SIXTEEN_MB) {      
            bank_offset := uint16(uint32(m.ram_bank) * 0x2000)
            fixed_addr := uint16(addr -0xa000) + bank_offset
            retval =       m.ram[fixed_addr]            
            fmt.Printf("RAM  BANK READ:%v  %04X->%04X:%x\n", m.ram_bank,addr,fixed_addr,retval)
        } else {
            panic("Tried to read from ram that wasn't enabled!")
        }
    }

    return retval
}

func (m * ROM_MBC1 )Write_b(addr uint16,val uint8)() {
    if addr >= 0x2000 && addr < 0x4000  {
        if (val >1){
               //fmt.Println("ROM Bank from",m.bank,val-1)
                m.bank = uint16((val) -1) 
            }else{
                m.bank = uint16(0)
            }
    } else if addr < 0x2000 {
            if m.memory_mode == FOUR_MB{
                if val == 0x0A {
                    m.ram_enabled = true
                    fmt.Println("RAM enabled",val ) 

            }else{            
                fmt.Println("RAM Disabled",val ) 
                m.ram_enabled = false

            }
        }

    }else if addr >= 0x4000 && addr < 0x6000{

        fmt.Println("RAM bank", "from",m.ram_bank,"to",val)
        m.ram_bank = (val) 
    }else if addr >= 0x6000 && addr < 0x7000{
        if val >0  {
            m.memory_mode = SIXTEEN_MB
            fmt.Println("RAM Memory mode",val,"selected")
        }else{
            m.memory_mode = FOUR_MB  
        }

    }else if addr >= 0xA000 && addr < 0xc000{

    if (m.memory_mode == FOUR_MB && m.ram_enabled == true) || (m.memory_mode == SIXTEEN_MB) {      

      bank_offset := uint16(uint32(m.ram_bank) * 0x2000)
      fixed_addr := uint16(addr -0xa000) + bank_offset
      m.ram[fixed_addr] = val
      fmt.Printf("RAM  BANK WRITE:%v  %04X->%04X:%x\n", m.ram_bank,addr,fixed_addr,val)
        } else {
            panic("Tried to read from ram that wasn't enabled!")
        }
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

