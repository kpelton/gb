package cpu
//import "fmt"

type MMU struct {

    mem [0x10000]uint8

}

func (m *MMU)read_b(addr uint16) (uint8) {

    return m.mem[addr]  

}

func (m *MMU)read_w(addr uint16) (uint16) {
    return uint16(m.mem[addr+1]) <<8 | uint16(m.mem[addr])  
   
}

func (m *MMU)write_b(addr uint16,val uint8) () {

    m.mem[addr] = val

}

func (m *MMU)write_w(addr uint16,val uint16) () {
        
    m.mem[addr] = uint8(val & 0x00ff)
    m.mem[addr+1] = uint8((val & 0xff00)>>8)
    


}
