package cpu


import (
    "testing" //import go package for testing related functionality
    "fmt"
    )
func TestLD_r1_r2(T *testing.T){
   
    c:= NewCpu()

    c.reg8["B"] = 0xff
    c.ops[0x78](c) //LD A,B
    if (c.reg8["A"] !=0xff) {
        T.Error("Fail for LD r1,r2")
    }

}
func TestLD_hl_r1(T *testing.T){
   
    c:= NewCpu()
    
    c.reg8["H"] = 0xff
    c.reg8["L"] = 0xEE
    c.reg8["A"]= 0xef

    c.ops[0x77](c) //LD A,(HL)


    if (c.mem[0xffee] !=0xef) {
        T.Error("Fail on A,(HL)",c.reg8,c.mem[0xffee])
    }
   }

func TestLD_r1_hl(T *testing.T){
   
    c:= NewCpu()
    
    c.reg8["H"] = 0xff
    c.reg8["L"] = 0x0
    
    c.mem[0xff00] = 0xbeef
    c.ops[0x7e](c) //LD A,(HL)


    if (c.reg8["A"] !=0xef) {
        T.Error("Fail on A,(HL)",c.reg8)
    }
    //flip bits and try the other way
    c.reg8["H"] = 0x00
    c.reg8["L"] = 0xff
    c.mem[0x00ff] = 0xbeef
    c.ops[0x7E](c) //LD A,(HL)
    
    if c.reg8["A"] !=0xef {
        T.Error("Fail on A,(HL)",c.reg8,c.mem[0x00ff])
    }
 

}
func TestLD_r1_rr(T *testing.T){
   
    c:= NewCpu()
    
    c.reg8["D"] = 0xff
    c.reg8["E"] = 0x0
    
    c.mem[0xff00] = 0xbeef
    c.ops[0x1A](c) //LD A,B


    if (c.reg8["A"] !=0xef) {
        T.Error("Fail on A,(DE)",c.reg8)
    }
    //flip bits and try the other way
    c.reg8["D"] = 0x00
    c.reg8["E"] = 0xff
    c.mem[0x00ff] = 0xbeef
    c.ops[0x1A](c) //LD A,B

    if c.reg8["A"] !=0xef {
        T.Error("Fail on A,(DE)",c.reg8)
    }

}




func TestLD_nn_n(T *testing.T){
   
    c:= NewCpu()

    c.mem[c.reg8["PC"]+1] = 0xff
    c.ops[0x06](c) //LD B,n
    if (c.reg8["B"] !=0xff) {
        T.Error("Fail for LD nn,n")
    }
    
   
}
func TestLD_r_n(T *testing.T){
    c:= NewCpu()
    for i := 0; i < 0x4; i++ {
        c.mem[c.reg16["PC"]+1] = uint16(i & 0xff) //LSB        
        c.ops[0x3e](c) //LD A,n
        if (c.reg8["A"] != uint8(i &0xff)) {
            T.Error("Fail for LD r,n",c.reg8)
        }
}
}


func TestLD_r_nn(T *testing.T){
   
    c:= NewCpu()

    c.mem[c.reg16["PC"]+1] = 0xff //LSB
    c.mem[c.reg16["PC"]+2] = 0xff
    c.mem[0xffff] = 0x1122

    c.ops[0xFA](c) //LD A,(nn)

    if (c.reg8["A"] !=0x22) {
        T.Error("Fail for LD r,nn",c.reg8)
    }

    fmt.Println(c.reg16)
    c.mem[c.reg16["PC"]+1] = 0xff //LSB
    c.mem[c.reg16["PC"]+2] = 0x00
    c.mem[0x00ff] = 0x1122
       
    c.ops[0xFA](c) //LD A,(nn)

    if (c.reg8["A"] !=0x22) {
        T.Error("Fail for LD r,nn",c.reg16,c.mem[0:6])
    }
    

    fmt.Println(c.reg16)
    c.mem[c.reg16["PC"]+1] = 0xff //LSB
    c.mem[c.reg16["PC"]+2] = 0xff
    c.mem[0xffff] = 0x2211
       
    c.ops[0xFA](c) //LD A,(nn)

    if (c.reg8["A"] !=0x11) {
        T.Error("Fail for LD r,nn",c.reg16,c.mem[0:6])
    }

   
}