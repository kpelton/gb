package serial
import (
    "fmt"   
    "ic"
      "constants"
)


type FakeSerial struct {
    SB uint8
    SC uint8
    ic *ic.IC
}

func NewFakeSerial(ic *ic.IC) *FakeSerial {
	serial := new(FakeSerial)
    serial.ic = ic
	return serial
}

func (s *FakeSerial) Update() uint8 {
    return constants.SERIAL
}

func (s *FakeSerial) Read(addr uint16) uint8 {
	switch addr {
	case SB_ADDR:
    	return s.SB
	case SC_ADDR:
        return s.SC 
    default:
        panic("mis-routed serial write!")
    }
}

func (s *FakeSerial) Write(addr uint16,val uint8) {
	switch addr {

	case SB_ADDR:
        fmt.Printf("->SERIALB:%04X\n", val)
		s.SB = val
	case SC_ADDR:
        s.SC = val
        if val & 0x81 == 0x81{
            s.SB = 0xff
            s.SC &=  val &(^uint8(0x80))
            s.ic.Assert(constants.SERIAL)

        }
        
    default:
        panic("mis-routed serial write!")
    }
}
