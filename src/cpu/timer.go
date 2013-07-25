package cpu

import (
//      "fmt"
)

type Timer struct {
    TAC uint8 // Timer Control (R/W)
    TMA uint8 // Timer Modulo (R/W)
	TIMA uint8 // Timer counter (R/W)
    last_update  int // in clock cycles
}

const (
   HZ_4096 =  0
   HZ_262_144 = 1
   HZ_65_536 = 2
   HZ_16_384 = 3
   START_TIMER = 0x4
   HZ_4096_t = 1024
   HZ_262_144_t = 16
   HZ_65_536_t = 64
   HZ_16_384_t = 256

)

func NewTimer() *Timer {
    timer :=new(Timer)
    timer.last_update = 0
    return timer
}
func (timer *Timer) update_regs(ic *IC) {
       timer.TIMA +=1
       //fmt.Println("TIMA",timer.TIMA,timer.last_update)
       if (timer.TIMA == 0) { 
            timer.TIMA = timer.TMA
            ic.Assert(TIMER)
            //fmt.Println("ASSERTED TIMER")
       }
     

}
func (timer *Timer) Update(ic *IC, cycles uint64) {
            
    //   fmt.Printf("TIMA:%x,%v\n",timer.TIMA,int(timer.last_update))

    if timer.TAC & START_TIMER == START_TIMER {
     // fmt.Printf("CYLCES PASSED :%x\n",cycles)
                    // fmt.Println("CYCLES:",timer.last_update)
        t_type := 0
        switch (timer.TAC & 0x3) {
            case HZ_4096:
               t_type = HZ_4096_t
             case HZ_16_384:
                t_type =  HZ_16_384_t
            case HZ_65_536:
                t_type = HZ_65_536_t
            case HZ_262_144:
                t_type = HZ_262_144_t
            default:
                panic("Unsupported timer frequency!\n")
        }
        timer.last_update -= int(cycles)
        for int(timer.last_update) < 1{

                    timer.update_regs(ic) 
                    //fmt.Println(timer.last_update)
                    timer.last_update += t_type
                }
                
    } else{
                        timer.last_update =0
}






   
}