package cpu
import "fmt"

type MMU struct {

    mem [0x10000]uint8
    vm  [0x2000] uint8
    gpu *GPU 
}
func NewMMU(gpu *GPU )(*MMU) {
    m :=new(MMU)
    m.gpu = gpu
    return m
}
func (m *MMU) Dump_vm() {
    j:=0;
    fmt.Printf("\n0x8000:")
    for i:=0; i<0x2000; i++ {
        fmt.Printf("%02X ",m.vm[i])
        j++
        if j==16 {
            fmt.Printf("\n0x%04X:",i+1+0x8000)
            j=0
        } 
    }

        
    }

func (m* MMU) write_mmio(addr uint16,val uint8) () {
    switch (addr) {
        case 0xff40:
            m.gpu.LCDC = val
        case 0xff41:
            m.gpu.STAT = val
        case 0xff42:
            m.gpu.SCY = val
        case 0xff43:
            m.gpu.SCX = val
        case 0xff44:
            m.gpu.LY = val
        case 0xff45:
            m.gpu.LYC = val
    }

}
func (m* MMU) read_mmio(addr uint16) (uint8) {
    var val uint8 = 0
    switch (addr) {
        case 0xff40:
            val= m.gpu.LCDC
        case 0xff41:
            val=m.gpu.STAT
        case 0xff42:
            val=m.gpu.SCY
        case 0xff43:
            val=m.gpu.SCX
        case 0xff44:
            val=m.gpu.LY
        case 0xff45:
            val=m.gpu.LYC
    }

    return val
}
func (m *MMU)read_b(addr uint16) (uint8) {
    
    if  addr >= 0x8000 && addr < 0xa000  {
        return m.vm[addr & 0x1fff]  
    } else if addr >= 0xff40 && addr < 0xff46{
        return m.read_mmio(addr)      
    }
    return m.mem[addr]
    

}

func (m *MMU)read_w(addr uint16) (uint16) {
    return uint16(m.read_b(addr)) + uint16((m.read_b(addr+1))) << 8
}

func (m *MMU)write_b(addr uint16,val uint8) () {

    if addr >= 0x8000 && addr < 0xA000{
        m.vm[addr & 0x1fff] = val
        //fmt.Printf("Video:0x%04X->0x%02X\n",addr,val) 

            m.vm[addr & 0x1fff] = val
        return
    } else if addr >= 0xff40 && addr < 0xff46{
        m.write_mmio(addr,val)
        return      
    }   


    
    m.mem[addr] = val
    
}

func (m *MMU)write_w(addr uint16,val uint16) () {
        
    m.write_b(addr,uint8(val & 0x00ff))
    m.write_b(addr+1,uint8((val & 0xff00)>>8))
    

}
