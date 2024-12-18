package serial

import (
	"fmt"
	"gb/component"
	"gb/constants"
	"gb/ic"
	"io"

	serial "github.com/tarm/goserial"
)

type RealSerial struct {
	SB          uint8
	SC          uint8
	ic          *ic.IC
	cycle_count uint16
	started     bool
	port        io.ReadWriteCloser
	reg_list    component.RegList
}

func (g *RealSerial) Get_reg_list() component.RegList {
	return g.reg_list
}

func NewRealSerial(ic *ic.IC, port string) *RealSerial {
	nserial := new(RealSerial)
	nserial.reg_list = component.RegList{
		{Name: "SC", Addr: 0xff01},
		{Name: "SB", Addr: 0xff02},
	}
	nserial.ic = ic
	c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 115200}
	s, err := serial.OpenPort(c)
	if err != nil {
		panic(err)
	}
	nserial.port = s
	var buf [4096]uint8

	s.Read(buf[0:])

	return nserial

}
func (g *RealSerial) Reset() {
	g.SC = 0
	g.SB = 0
}

func (s *RealSerial) Update(cycles uint16) uint8 {
	var buf [1]uint8

	for i := 0; i < 5; i++ {
		if s.started {
			n, _ := s.port.Read(buf[0:])
			if n > 0 {
				s.SB = buf[n-1]
				fmt.Println("PORT", buf)

				s.started = false
				s.ic.Assert(constants.SERIAL)
				s.SC &= (^uint8(0x80))

			}
			s.cycle_count = 0
		} else {
			s.cycle_count -= cycles
		}

	}
	return 0

}

func (s *RealSerial) Read_mmio(addr uint16) uint8 {
	switch addr {
	case SB_ADDR:
		return s.SB
	case SC_ADDR:
		return s.SC
	default:
		panic("mis-routed serial write!")
	}
}

func (s *RealSerial) Write_mmio(addr uint16, val uint8) {
	switch addr {

	case SB_ADDR:
		fmt.Printf("->SERIALB:%04X\n", val)
		s.SB = val
		var x [1]uint8
		x[0] = s.SB

		fmt.Println("WRITE", x[0])
		s.port.Write(x[0:])
	case SC_ADDR:
		s.SC = val
		fmt.Printf("->SERIALC:%04X\n", val)
		if !s.started && val&0x80 == 0x80 {
			s.started = true

		}

	default:
		panic("mis-routed serial write!")
	}
}
