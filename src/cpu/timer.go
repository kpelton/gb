package cpu

import (
      "fmt"
)

type Timer struct {
    TAC uint8 // Timer Control (R/W)
    TMA uint8 // Timer Modulo (R/W)
	TIMA uint8 // Timer counter (R/W)
    last_update uint64 // in clock cycles
}

const (
   HZ_4096 =  0
   HZ_262_144 = 1
   HZ_65_536 = 2
   HZ_16_384 = 3
   START_TIMER = 0x4
   HZ_4096_t = 1024
   HZ_262_144_t = 4
   HZ_65_536_t = 64
   HZ_16_384_t = 256

)

func NewTimer() *Timer {
    timer :=new(Timer)
    timer.last_update = 0
    return timer
}
func (timer *Timer) update_regs(ic *IC) {
    if (timer.TIMA == 0xff) {
        timer.TIMA = timer.TMA
        ic.Assert(TIMER)
    }else{
       timer.TIMA +=1
    }
    timer.last_update = 0

}
func (timer *Timer) Update(ic *IC, cycles uint64) {
            

    if timer.TAC & START_TIMER == START_TIMER {
        timer.last_update +=cycles

        switch (timer.TAC & 0x3) {
            case HZ_4096:
               // fmt.Println("WAIT")
                if  timer.last_update >= HZ_4096_t  {
                                    fmt.Println("4096",timer.last_update,cycles,timer.TMA,timer.TIMA)

                                        timer.update_regs(ic);
                }

             case HZ_16_384:
                if   timer.last_update  >= HZ_16_384_t  {
                    timer.update_regs(ic);
                    fmt.Println("16384",timer.last_update,timer.TMA,timer.TIMA)
                }
            
            case HZ_65_536:
                if  timer.last_update  >= HZ_65_536_t  {
                    timer.update_regs(ic);
                    fmt.Println("65536",timer.last_update,timer.TMA,timer.TIMA)
                }
            case HZ_262_144:
                if   timer.last_update  >= HZ_262_144_t {
                                         fmt.Println("262144",timer.last_update,cycles,timer.TMA,timer.TIMA)

                    timer.update_regs(ic);

                }

            default:
                fmt.Printf("Unsupported timer frequency!\n");
        }

    }



   
}