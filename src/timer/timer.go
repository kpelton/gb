package timer

import (
    "fmt"
)

type Timer struct {
	TAC         uint8 // Timer Control (R/W)
	TMA         uint8 // Timer Modulo (R/W)
	TIMA        uint8 // Timer counter (R/W)
	last_update int   // in clock cycles
}

const (
	HZ_4096      = 0
	HZ_262_144   = 1
	HZ_65_536    = 2
	HZ_16_384    = 3
	START_TIMER  = 0x4
	HZ_4096_t    = 1024
	HZ_262_144_t = 16
	HZ_65_536_t  = 64
	HZ_16_384_t  = 256
)

func NewTimer() *Timer {
	timer := new(Timer)
	timer.Reset()

	return timer
}
func (timer *Timer) Reset() {
	timer.last_update = 0
	timer.TAC = 0
	timer.TIMA = 0
	timer.TMA = 0
}


func (timer *Timer) Write_mmio(addr uint16, val uint8) {
	fmt.Printf("TIMER:%x %x\n",addr,val)
	switch addr {
	case 0xff05:
		timer.TIMA = val
	case 0xff06:
		timer.TMA = val
	case 0xff07:
		timer.TAC = val
	default:
		panic("TIMER:unhandled timer mmio write")
	}
}

func (timer *Timer) Read_mmio(addr uint16) uint8 {
	var val uint8
				fmt.Printf("TIMER READ:%x \n",addr)

	switch addr {

	case 0xff05:
		val =timer.TIMA
	case 0xff06:
		val = timer.TMA 
	case 0xff07:
		val = timer.TAC 
	default:
		panic("TIMER:unhandled timer mmio read")
	}
	return val
}

func (timer *Timer) update_regs() (bool){
	timer.TIMA += 1
	if timer.TIMA == 0 {
		timer.TIMA = timer.TMA
	    return true
    }
    return false
}
func (timer *Timer) check_cycles(t_type int,cycles uint64) (uint8) {
    var raised_int uint8 = 0x0
    timer.last_update -= int(cycles)
	for int(timer.last_update) < 1 {
        if timer.update_regs() {
            raised_int = 0x4
        }
		timer.last_update += t_type
        
	}
    return raised_int
}

func (timer *Timer) Update( cycles uint64) (uint8) {

	//   fmt.Printf("TIMA:%x,%v\n",timer.TIMA,int(timer.last_update))

	if timer.TAC&START_TIMER == START_TIMER {
		// fmt.Printf("CYLCES PASSED :%x\n",cycles)
		// fmt.Println("CYCLES:",timer.last_update)
		t_type := 0
		switch timer.TAC & 0x3 {
		case HZ_4096:
			t_type = HZ_4096_t
		case HZ_16_384:
			t_type = HZ_16_384_t
		case HZ_65_536:
			t_type = HZ_65_536_t
		case HZ_262_144:
			t_type = HZ_262_144_t
		default:
			panic("Unsupported timer frequency!\n")
		}
        return timer.check_cycles(t_type,cycles)
	} else {
		timer.last_update = 0
	}
    return 0
}
