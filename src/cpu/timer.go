package cpu

import (
    "fmt"
    "time"
)

type Timer struct {
    TAC uint8 // Timer Control (R/W)
    TMA uint8 // Timer Modulo (R/W)
	TIMA uint8 // Timer counter (R/W)
    last_update time.Time
}

const (
   KHZ_4096 =  0
   KHZ_262_144 = 1
   KHZ_65_536 = 2
   KGZ_16_384 = 3
   START_TIMER = 0x4

)

func NewTimer() *Timer {
    timer :=new(Timer)
    timer.last_update = time.Now()
    return timer
}

func (timer *Timer) Update() bool {
    if timer.TAC & START_TIMER == START_TIMER {
        elapsed := time.Since(timer.last_update)
        switch (timer.TAC & 0x3) {
            case KHZ_4096:
                if   elapsed >= 10000*time.Microsecond  {
                    timer.last_update = time.Now()
                    if (timer.TIMA == 0xff) {
                        timer.TIMA = timer.TMA
                        //fmt.Printf("TIME_INT#\n");

                        return true  //MSG int

                    }else{
                        timer.TIMA +=1
                    }

                    //fmt.Println(elapsed,timer.TMA,timer.TIMA)
                }
            default:
                fmt.Printf("Unsupported timer frequency!");
        }

    }
    return false // no int 
}


   