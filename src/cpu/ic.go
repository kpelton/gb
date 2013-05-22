package cpu

import (
   "fmt"
//    "time"
)

type Ic struct {
    IE uint8  //Interrupt Enable (R/W)
    IF uint8 // Interrupt Flag (R/W)
    
}

const (
   V_BLANK = 0
   LCDC 
   TIMER 
   SERIAL 
)

func New() *Ic {
    return new(Ic)
}

func (i *Ic) Assert(signal uint8) {
    //check to see if it is masked off
    if (i.IE & signal == signal) {
        i.IF |= signal
    } else {
        fmt.Println("Interrupt",signal,"Masked off!!!")
    }
}

func (i *Ic) Disassert(signal uint8) {
    //check to see if it is masked off
     i.IF &= ^signal
}


func (i *Ic)  Handle() uint8 {
      var value uint8 = 0
      switch {
  
        case i.IF & V_BLANK == V_BLANK:
            i.Disassert(V_BLANK)
            value = 0x40            
        case i.IF & TIMER == TIMER:
            i.Disassert(V_BLANK)
            value = 0x50
            
    }
    return(value)
}

