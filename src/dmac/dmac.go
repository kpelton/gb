package dmac

import (
	"fmt"
	"component"
)

const (
	MMIO_SRC_HIGH = 0xff51
	MMIO_SRC_LO = 0xff52
	MMIO_DST_HIGH = 0xff53
	MMIO_DST_LO = 0xff54
	MMIO_START = 0xff55
	MMIO_DMA = 0xff46
	OAM_START = 0xfe00
	OAM_SIZE = 160
)

type DMAC struct {
	HDMA_hi_src uint8
	HDMA_lo_src uint8
	HDMA_hi_dst uint8
	HDMA_lo_dst uint8
	HDMA_start uint8
	HDMA_start_shadow uint8
	HDMA_done_shadow uint8
	HDMA_hi_src_shadow uint8
	HDMA_lo_src_shadow uint8
	HDMA_hi_dst_shadow uint8
	HDMA_lo_dst_shadow uint8

	hblank_bytes_done uint16
	hdma_complete_cycles uint16
	hblank_dma_active bool
	oam_dma_active bool
	oam_dma_addr uint8
	mmu component.MemComponent
	reg_list component.RegList
}

func NewDMAC(mmu component.MemComponent) *DMAC {
	m := new(DMAC)
	m.reg_list = component.RegList{
		{Name:"DMA",Addr:MMIO_DMA},
		{Name:"HDMA_SRC_HIGH",Addr:MMIO_SRC_HIGH},
		{Name:"HDMA_SRC_LO",Addr:MMIO_SRC_LO},
		{Name:"HDMA_DST_HIGH",Addr:MMIO_DST_HIGH},
		{Name:"HDMA_DST_LO",Addr:MMIO_DST_LO},
		{Name:"HDMA_start",Addr:MMIO_START},
	}
	m.mmu = mmu

	m.HDMA_start = 0xff
	m.hblank_dma_active =false
	m.oam_dma_active = false
	return m

}

func (m *DMAC) exec_dma(addr uint8) {
	var real_addr uint16
	var i uint16
	real_addr = uint16(addr) * 0x100
	for i = 0; i < OAM_SIZE; i++ {
		m.mmu.Write(OAM_START+i,m.mmu.Read(real_addr + i))
		
	}
	m.oam_dma_active=false

}
func (m* DMAC) Get_reg_list() component.RegList{
	return m.reg_list
}
func (m* DMAC) Reset() {
	m.HDMA_hi_src = 0
	m.HDMA_lo_src = 0
	m.HDMA_hi_dst = 0
	m.HDMA_lo_dst = 0
	m.HDMA_start = 0
}

func (m* DMAC) Update() uint16 {
	var cycles uint16
	if m.oam_dma_active {
		m.exec_dma(m.oam_dma_addr)
		cycles=4
	}
	if m.hdma_complete_cycles > 0 {
		cycles+=m.hdma_complete_cycles
		m.hdma_complete_cycles=0
	}
	return cycles


}



func  (m *DMAC) Hblank_DMA() {
	if m.hblank_dma_active == false {
		return
	}

	fmt.Println("HDMA FUNC")
	
	val:= m.HDMA_start_shadow
	src := uint16(m.HDMA_hi_src_shadow) <<8 | uint16(m.HDMA_lo_src_shadow& 0xf0)
	//Add bytes_done to src
	src += uint16(m.hblank_bytes_done)
        dst := uint16(m.HDMA_hi_dst_shadow&0x1f) <<8 | uint16(m.HDMA_lo_dst_shadow&0xf0)
	dst|=0x8000
	dst += uint16(m.hblank_bytes_done)
	
	length := (uint16( (val & 0x7f)) +1) *0x10
		fmt.Printf("->HDMA Hblank transfer:%04X %04x %x %x\n", src,dst,length,m.hblank_bytes_done)


		var i uint16
		for i=0; i < uint16(0x10); i++ {
			m.mmu.Write(dst +i,m.mmu.Read(src+i))
		}
		m.HDMA_start=uint8((length-(m.hblank_bytes_done))/0x10)-1 | 0x80
        src+=0x10
        dst+=0x10
	    m.HDMA_lo_src  = uint8((src) &0xf0)
		m.HDMA_hi_src = uint8((src&0xff00 >>8))
		m.HDMA_lo_dst  = uint8((dst) &0xf0)
		m.HDMA_hi_dst = uint8((dst&0xff00 >>8))
		m.hblank_bytes_done +=0x10
		if length == m.hblank_bytes_done {
			m.hblank_dma_active = false
			m.HDMA_start=0xff
			fmt.Printf("->DONE HDMA Hblank transfer:%04X %04x %x\n", src,dst,0x10)
			return
		}

		//fmt.Printf("Blocks remaining:%x\n",length/m.hblank_bytes_done)
        fmt.Printf("0x%x 0x%x 0x%x 0x%x\n",m.HDMA_lo_src,m.HDMA_hi_src,m.HDMA_lo_dst,m.HDMA_hi_dst)
	}
func (m *DMAC) gen_dma() {
	val:= m.HDMA_start_shadow
	src := uint16(m.HDMA_hi_src ) <<8 | uint16(m.HDMA_lo_src&0xf0)
	dst := uint16(m.HDMA_hi_dst &0x1f) <<8 | uint16(m.HDMA_lo_dst&0xf0)
	dst |=0x8000
	length := (uint16( (val & 0x7f)) + 1) *0x10
	fmt.Printf("->DONE HDMA transfer:%04X %04x %x\n", src,dst,length)
	var i uint16
	for i = 0; i < uint16(length); i++ {
		m.mmu.Write(dst,m.mmu.Read(src))
        src+=1
        dst+=1
	}
	m.HDMA_lo_src  = uint8((src) &0xf0)
	m.HDMA_hi_src = uint8(((src)&0xff00)>>8)
	m.HDMA_lo_dst  = uint8((dst) &0xf0)
	m.HDMA_hi_dst = uint8(((dst)&0xff00)>>8)
	m.HDMA_start=0xff
	m.hdma_complete_cycles =(length/0x10)*16*4
	fmt.Println("HDMA cycles completed:",m.hdma_complete_cycles)
    fmt.Printf("0x%x 0x%x 0x%x 0x%x\n",m.HDMA_lo_src,m.HDMA_hi_src,m.HDMA_lo_dst,m.HDMA_hi_dst)
}

func (m *DMAC) Write_mmio(addr uint16,val uint8)  {
	switch addr {
	case 0xff46:
		// m.Dump_vm()
//		m.exec_dma(val)
		m.oam_dma_active = true
		m.oam_dma_addr =  val
	case 0xff51:
		m.HDMA_hi_src = val
	case 0xff52:
		m.HDMA_lo_src = val
	case 0xff53:
		m.HDMA_hi_dst = val
	case 0xff54:
		m.HDMA_lo_dst = val
	case 0xff55:

		if  val & 0x80 == 0x80 &&m.hblank_dma_active ==false  {
			m.HDMA_start_shadow = val &0x7f

			m.HDMA_hi_src_shadow =m.HDMA_hi_src
			m.HDMA_lo_src_shadow =m.HDMA_lo_src
			m.HDMA_hi_dst_shadow =m.HDMA_hi_dst
			m.HDMA_lo_dst_shadow =m.HDMA_lo_dst
			m.HDMA_start = val 
			m.hblank_bytes_done = 0
			m.hblank_dma_active = true
			fmt.Println("HDMA STARTED")
		}else if val&0x80 != 0x80&&  m.hblank_dma_active == true {
			m.hblank_bytes_done = 0
			m.hblank_dma_active = false

		}else if  m.hblank_dma_active == false{
		m.HDMA_start_shadow = val &0x7f
		m.gen_dma()
		}

	default:
		panic("unhandled dmac mmio write")
	}

}

func (m *DMAC) Read_mmio(addr uint16) uint8  {
	var val uint8 = 0
	switch addr {
	case MMIO_DMA:
		val = 0xff
	case MMIO_START:
		val = m.HDMA_start
	default:
		fmt.Printf("unhandled dmac mmio read:%04x\n", addr)
	}
	return val
}



