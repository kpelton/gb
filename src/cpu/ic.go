package cpu

//import (
//   "fmt"
//    "time"
//)

type IC struct {
    IE uint8  //Interrupt Enable (R/W)
    IF uint8 // Interrupt Flag (R/W)
    
}

const (
   V_BLANK = 1
   LCDC    = 0x2
   TIMER = 0x4 
   SERIAL  = 0x8
)

func NewIC() *IC {
    return new(IC)
}

func (i *IC) Assert(signal uint8) {
    //check to see if it is masked off
    //fmt.Println("ASSERT",signal,i.IE,i.IF)

    i.IF |= signal
}

func (i *IC) Disassert(signal uint8) {

     i.IF &=  ^signal
     //fmt.Println("Disassert",signal,i.IF,i.IE)
}


func (i *IC)  Handle() uint16 {
    var value uint16 = 0
    //fmt.Println(i.IE,i.IF)

     switch {
  
        case (i.IF & V_BLANK == V_BLANK) && (i.IE & V_BLANK == V_BLANK) :
            i.Disassert(V_BLANK)
            value = 0x40 

        case (i.IF & TIMER == TIMER) && (i.IE & TIMER == TIMER) :
            i.Disassert(TIMER)
            value = 0x50
        
    }
    return(value)
}

