package serial

import (
	"gb/component"
	"gb/constants"
	"gb/ic"
)

type FakeSerial struct {
	SB          uint8
	SC          uint8
	ic          *ic.IC
	cycle_count uint16
	started     bool
	reg_list    component.RegList
}

func (g *FakeSerial) Get_reg_list() component.RegList {
	return g.reg_list
}
func NewFakeSerial(ic *ic.IC) *FakeSerial {
	serial := new(FakeSerial)
	serial.reg_list = component.RegList{
		{Name: "SC", Addr: 0xff01},
		{Name: "SB", Addr: 0xff02},
	}

	serial.ic = ic
	return serial
}

func (s *FakeSerial) Update(cycles uint16) uint8 {

	if s.started {
		if s.cycle_count <= 0 {
			s.SB = 0xff
			s.SC &= (^uint8(0x80))
			s.ic.Assert(constants.SERIAL)
			s.started = false
			s.cycle_count = 0
		}
		s.cycle_count -= cycles
	}
	return 0

}

func (s *FakeSerial) Read_mmio(addr uint16) uint8 {
	switch addr {
	case SB_ADDR:
		return s.SB
	case SC_ADDR:
		return s.SC
	default:
		panic("mis-routed serial write!")
	}
}

func (s *FakeSerial) Reset() {
	s.SB = 0
	s.SC = 0
}

func (s *FakeSerial) Write_mmio(addr uint16, val uint8) {
	switch addr {

	case SB_ADDR:
		//fmt.Printf("->SERIALB:%04X\n", val)
		s.SB = val
	case SC_ADDR:
		s.SC = val
		//fmt.Printf("->SERIALC:%04X\n", val)
		if !s.started && val&0x81 == 0x81 {
			s.started = true
			s.cycle_count = HZ_8192_t

		}

	default:
		panic("mis-routed serial write!")
	}
}
