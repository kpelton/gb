package cpu
import "fmt"

type MMU struct {

    mem [0x10000]uint8
    cart [0x8000]uint8
    vm  [0x2000] uint8
	oam [0xA0] uint8
	cpu *CPU
    inbios bool
}
func NewMMU(cpu *CPU)(*MMU) {
    m :=new(MMU)
    m.inbios = false
	m.cpu = cpu
	return m
}

func (m *MMU) Dump_mem() {
    j:=0;
    fmt.Printf("\n0x0000:")
    for i:=0x8000; i<0xafff; i++ {
        fmt.Printf("%02X ",m.vm[i])
        j++
        if j==16 {
            fmt.Printf("\n0x%04X:",i+1+0x0000)
            j=0
        } 
    }

        
    }
func (m *MMU) Dump_vm() {
    j:=0;
    fmt.Printf("\n0x8000:")
    for i:=0x0000; i<0x20000; i++ {
        fmt.Printf("%02X ",m.vm[i])
        j++
        if j==16 {
            fmt.Printf("\n0x%04X:",i+1+0x8000)
            j=0
        } 
    }    
}


func (m* MMU) exec_dma(addr uint8) () {
	var real_addr uint16 
	var i uint16
	real_addr = uint16(addr) *0x100
	
	for i = 0; i < 160; i++ {
		m.oam[i] = m.read_b(real_addr+i)
	}
}

func (m* MMU) write_mmio(addr uint16,val uint8) () {
    switch (addr) {
		case 0xff00:
            m.cpu.gp.P1 = val
		//		fmt.Printf("->P1:%04X\n",val)

        case 0xff40:
            m.cpu.gpu.LCDC = val
		//fmt.Printf("VAL:%04X\n",val)
			fmt.Printf("->LCDC:%04X\n",val)

        case 0xff41:
            m.cpu.gpu.STAT = val
		//	fmt.Printf("->STAT:%04X\n",val)

        case 0xff42:
            m.cpu.gpu.SCY = val
        case 0xff43:
            m.cpu.gpu.SCX = val
        case 0xff44:
            m.cpu.gpu.LY = 0
		    //fmt.Printf("->LY:%04X\n",val)

        case 0xff45:
            m.cpu.gpu.LYC = val
			//fmt.Printf("->LYC:%04X\n",val)
		
        case 0xff46:
	   m.exec_dma(val)
        case 0xff47:
           m.cpu.gpu.BGP=val 
          //fmt.Printf("->BGP:%04X\n",val)
	case 0xff4A:
            m.cpu.gpu.WY = val
	case 0xff4B:
            m.cpu.gpu.WX = val
		
    }

}
func (m* MMU) read_mmio(addr uint16) (uint8) {
    var val uint8 = 0
    switch (addr) {
		case 0xff00:
           val=m.cpu.gp.P1 		
		
		//fmt.Printf("<-P1:%04X\n",val)

		case 0xff04:
	 	    val = m.cpu.DIV

        case 0xff40:
            val= m.cpu.gpu.LCDC
		//		fmt.Printf("<-LCDC:%04X\n",val)

        case 0xff41:
            val=m.cpu.gpu.STAT
	//			fmt.Printf("<-STAT:%04X\n",val)

        case 0xff42:
            val=m.cpu.gpu.SCY
        case 0xff43:
            val=m.cpu.gpu.SCX
        case 0xff44:
            val=m.cpu.gpu.LY
        case 0xff45:
            val=m.cpu.gpu.LYC
	//fmt.Printf("->LYC:%04X\n",val)
	case 0xff46:
             panic("DMA register is not readable!")
	case 0xff47:
	   val= m.cpu.gpu.BGP
        case 0xff4A:
            val = m.cpu.gpu.WY 
		//	fmt.Printf("->WY:%04X\n",val)
		case 0xff4B:
		//	fmt.Printf("->WX:%04X\n",val)
            val = m.cpu.gpu.WX

    }

    return val
}
func (m *MMU)read_b(addr uint16) (uint8) {
    
    if  addr >= 0x8000 && addr < 0xa000  {
        return m.vm[addr & 0x1fff]  
    } else if addr >= 0x100 && addr < 0x8000  {

        return m.cart[addr]  
    }else if addr <= 0x100 && !m.inbios {
        return m.cart[addr]  

	} else if (addr >= 0xfe00 && addr <= 0xfe9f){
		fmt.Printf("%x\n",addr)
        return m.oam[addr & 0x00ff]  
		
    } else if addr == 0xff00 || addr == 0xff04 || (addr >= 0xff40 && addr <= 0xff4B){
        return m.read_mmio(addr)      

	}else if addr >= 0xe000 && addr < 0xfe00{
		return m.mem[addr-0x1000]    
	}
    return m.mem[addr]
    

}

func (m *MMU)read_w(addr uint16) (uint16) {
    return uint16(m.read_b(addr)) + uint16((m.read_b(addr+1))) << 8
}

func (m *MMU)load_cart(addr uint16,val uint8) () {

	m.cart[addr] = val

}

func (m *MMU)write_b(addr uint16,val uint8) () {

	 //   fmt.Printf("write:%04x:%04x\n",addr,val)

    if addr >= 0x8000 && addr < 0xA000{
        m.vm[addr & 0x1fff] = val
       // fmt.Printf("Video:0x%04X->0x%02X\n",addr,val) 
        
            m.vm[addr & 0x1fff] = val
        return
    }else if addr >=0x100 && addr < 0x8000 {
        //m.cart[addr] =val
	    fmt.Printf("INVALID write:%04x:%04x\n",addr,val)

        return 
    }else if addr <= 0x100 && !m.inbios{      
       m.cart[addr] = val
        return 
    } else if addr == 0xff00 || (addr >= 0xff40 && addr <=0xff4B){
        m.write_mmio(addr,val)
        return
	} else if (addr >= 0xfe00 && addr <= 0xfe9f){
		m.oam[addr & 0x00ff] = val
		return
    }  else if addr >= 0xe000 && addr < 0xfe00{
		m.mem[addr-0x1000]=val
		fmt.Println("shadow")
		return
	
    }

    m.mem[addr] = val
    
}

func (m *MMU)write_w(addr uint16,val uint16) () {
        
    m.write_b(addr,uint8(val & 0x00ff))
    m.write_b(addr+1,uint8((val & 0xff00)>>8))
    

}
