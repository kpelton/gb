package cpu


import (
    "testing" //import go package for testing related functionality

    )
func TestLD_r1_r2(T *testing.T){
   
    c:= NewCpu()

    c.reg8["B"] = 0xff
    c.ops[0x78](c) //LD A,B
    if (c.reg8["A"] !=0xff) {
        T.Error("Fail for LD r1,r2")
    }

    c.reg8["H"] = 0xff
    c.reg8["L"] = 0x11
    c.ops[0xf9](c) //LD A,B
    if (c.reg16["SP"] !=0xff11) {
        T.Error("Fail for LD r16,r16",c.reg16)
    }



}

func TestLD_a_n(T *testing.T){

    c:= NewCpu()
    c.mmu.write_w(c.reg16["PC"]+1,0xAA)

    c.reg8["A"] = 0x11

    c.ops[0xE0](c) //LD A,B


    if c.mmu.read_b(0xffaa) !=0x11 {
        T.Error("Fail for LD (n),A",c.reg8,c.mmu.read_b(0xffaa))
    }

    c.mmu.write_w(c.reg16["PC"]+1,0xAA)

    c.mmu.write_w(0xffAA,0xff)


  

    c.ops[0xF0](c) //LD A,B


    if c.reg8["A"] !=0xff {
        T.Error("Fail for LD A,(n)",c.reg8,c.mmu.read_b(0xffaa))
    }




}

func TestLD_A_C(T *testing.T){
   
    c:= NewCpu()

    c.reg8["C"] = 0xaf
    
    c.mmu.write_b(0xffaf,0xff)
    c.ops[0xF2](c) //LD A,(C)

    if (c.reg8["A"] !=0xff) {
        T.Error("Fail for LD A,(C)",c.mmu.mem[0xffaf])
    }

   c.reg8["C"] = 0x11
    
   c.ops[0xE2](c) //LD (C),A
    if (c.mmu.read_w(0xff11)!=0xff) {
        T.Error("Fail for LD(C),A",c.reg8,c.mmu.mem[0xff11])
    }


}

func TestLD_A_HLI(T *testing.T){
   
    c:= NewCpu()

 
    c.reg8["H"] = 0xff
    c.reg8["L"] = 0xEE 

    c.mmu.write_b(0xffee,0xff)
    c.ops[0x2a](c) //LD A,(HLD)
    
   if (c.reg8["A"] !=0xff && c.reg8["H"] != 0x0 || c.reg8["L"] != 0xEF) {
        T.Error("Fail for LD A,(HLI)",c.reg8)
    }
    
    c.reg8["H"] = 0x0
    c.reg8["L"] = 0xEE 

    c.mmu.write_b(0x00ee,0xff)
    c.ops[0x2a](c) //LD A,(HLD)
    
   if (c.reg8["A"] !=0xff || c.reg8["H"] != 0 || c.reg8["L"] != 0xEF) {
        T.Error("Fail for LD A,(HLI)",c.reg8)   
   }
   c.reg8["H"] = 0x0
   c.reg8["L"] = 0x0
   c.reg8["A"] = 3
   c.ops[0x22](c) //LD (HLD),A

   if (c.mmu.read_b(0) != 3 && c.reg8["L"] == 1) {
        T.Error("Fail for LD A,(HLI)",c.reg8)
   }
   c.reg8["H"] = 0xff
   c.reg8["L"] = 0x0
   c.reg8["A"] = 4
   c.ops[0x22](c) //LD (HLD),A

   if (c.mmu.read_b(0xff00) != 4 || c.reg8["H"] == 0x0) {
        T.Error("Fail for LD A,(HLI)",c.reg8)
   }
    

}
func TestLD_A_HLD(T *testing.T){
   
    c:= NewCpu()

 
    c.reg8["H"] = 0xff
    c.reg8["L"] = 0xEE 

    c.mmu.write_b(0xffee,0xff)
    c.ops[0x3a](c) //LD A,(HLD)
    
   if (c.reg8["A"] !=0xff || c.reg8["L"]!=0xED) {
        T.Error("Fail for LD A,(HLD)",c.reg8)
    }
    
    c.reg8["H"] = 0x0
    c.reg8["L"] = 0xEE 

    c.mmu.write_b(0x00ee,0xff)
    c.ops[0x3a](c) //LD A,(HLD)
    
   if (c.reg8["A"] !=0xff || c.reg8["H"] != 0 || c.reg8["L"] != 0xED) {
        T.Error("Fail for LD A,(HLI)",c.reg8)
   }
   c.reg8["H"] = 0x0
   c.reg8["L"] = 0x0
   c.reg8["A"] = 3
   c.ops[0x32](c) //LD (HLD),A

   if (c.mmu.read_b(0) != 3 || c.reg8["L"] != 0xff) {
        T.Error("Fail for LD A,(HLI)",c.reg8)
   }
   c.reg8["H"] = 0xff
   c.reg8["L"] = 0x0
   c.reg8["A"] = 4
   c.ops[0x32](c) //LD (HLD),A

   if (c.mmu.read_b(0xff00) != 4 || c.reg8["H"] != 0xfE) {
        T.Error("Fail for LD A,(HLI)",c.reg8)
   }
    

}



func TestLD_hl_r1(T *testing.T){
   
    c:= NewCpu()
    
    c.reg8["H"] = 0xff
    c.reg8["L"] = 0xEE
    c.reg8["A"]= 0xef

    c.ops[0x77](c) //LD A,(HL)


    if (c.mmu.mem[0xffee] !=0xef) {
        T.Error("Fail on A,(HL)",c.reg8,c.mmu.mem[0xffee])
    }
   }


func TestLD_r1_hl(T *testing.T){
   
    c:= NewCpu()
    
    c.reg8["H"] = 0xff
    c.reg8["L"] = 0x0
    
    c.mmu.write_w(0xff00,0xbeef)
    c.ops[0x7e](c) //LD A,(HL)


    if (c.reg8["A"] !=0xef) {
        T.Error("Fail on A,(HL)",c.reg8)
    }
    //flip bits and try the other way
    c.reg8["H"] = 0x00
    c.reg8["L"] = 0xff
    c.mmu.write_w(0x00ff,0xbeef)
    c.ops[0x7E](c) //LD A,(HL)
    
    if c.reg8["A"] !=0xef {
        T.Error("Fail on A,(HL)",c.reg8)
    }
 

}

func TestLD_r1_rr(T *testing.T){
   
    c:= NewCpu()
    
    c.reg8["D"] = 0xff
    c.reg8["E"] = 0x0
    
    c.mmu.write_w(0xff00,0xbeef)
   
    c.ops[0x1A](c) //LD A,B

    
    if (c.reg8["A"] !=0xef) {
        T.Error("Fail on A,(DE)",c.reg8)
    }
    //flip bits and try the other way
    c.reg8["D"] = 0x00
    c.reg8["E"] = 0xff
    c.mmu.write_w(0x00ff,0xbeef)
    c.ops[0x1A](c) //LD A,B

    if c.reg8["A"] !=0xef {
        T.Error("Fail on A,(DE)",c.reg8)
    }

}

func TestLD_RR_nn(T *testing.T){

    c:= NewCpu()
    c.mmu.write_w(c.reg16["PC"]+1,0xBBCC)
    c.ops[0x01](c) //LD B,n
    if (c.reg8["B"] !=0xBB  ) {
        T.Error("Fail for LD BC,nn",c.reg8)
    }
    if (c.reg8["C"] !=0xCC  ) {
        T.Error("Fail for BC,nn",c.reg8)
    }


    c.mmu.write_w(c.reg16["PC"]+1,0xCCBB)
    c.ops[0x01](c) //LD B,n
    if (c.reg8["B"] !=0xCC  ) {
        T.Error("Fail for LD BC,nn",c.reg8)
    }
    if (c.reg8["C"] !=0xBB  ) {
        T.Error("Fail for BC,nn",c.reg8)
    }

    c.mmu.write_w(c.reg16["PC"]+1,0xCCBB)
    c.ops[0x21](c) //LD B,n
    if (c.reg8["H"] !=0xCC  ) {
        T.Error("Fail for LD BC,nn",c.reg8)
    }
    if (c.reg8["L"] !=0xBB  ) {
        T.Error("Fail for BC,nn",c.reg8)
    }






}


func TestLD_nn_n(T *testing.T){
   
    c:= NewCpu()
    c.mmu.write_b(c.reg16["PC"]+1,0xAA)
    c.ops[0x06](c) //LD B,n
    if (c.reg8["B"] !=0xAA) {
        T.Error("Fail for LD nn,n")
    }
    
   
}

func TestLD_r_n(T *testing.T){
    c:= NewCpu()
    for i := 0; i < 0x3; i++ {
        c.mmu.write_b(c.reg16["PC"]+1,uint8(i & 0xff))        
        c.ops[0x3e](c) //LD A,n
        if (c.reg8["A"] != uint8(i &0xff)) {
            T.Error("Fail for LD r,n",c.reg8)
        }
    }
}

func TestMMU(T *testing.T){

    c:= NewCpu()
    c.mmu.write_w(0xff00,0x1122)
    c.mmu.write_b(0x0000,0x11)

    if c.mmu.mem[0xff00] != 0x22 {
        T.Error("Fail MMU read_w(lower) expected 0x22 got:", c.mmu.mem[0xff00] )
    }
    if c.mmu.mem[0xff01] != 0x11 {
        T.Error("Fail MMU read_w(upper) exepcted 0x11 got:",c.mmu.mem[0xff00]  )
    }
    if c.mmu.read_b(0x0000) != 0x11 {
        T.Error("Fail MMU read_b: expected 0x11 got",c.mmu.mem[0x0000]  )
    }
    
    if c.mmu.read_w(0xff00) != 0x1122 {
        T.Error("Fail MMU read_w:", c.mmu.read_w(0xff00) )
    }

    if c.mmu.read_w(0x000) != 0x0011 {
        T.Error("Fail MMU read_w:", c.mmu.read_w(0x0000) )
    }

    if c.mmu.read_b(0xff00) != 0x22 {
        T.Error("Fail MMU read_b expected 0x22")
    }
    if c.mmu.read_b(0xff01) != 0x11 {
        T.Error("Fail MMU read_b expected 0x11")
    }
    if c.mmu.read_w(0xff00) != 0x1122 {
        T.Error("Fail MMU read_b expected 0x11")
    }



}
func TestLD_r_nn(T *testing.T){
   
    c:= NewCpu()

    c.mmu.write_b(c.reg16["PC"]+1,0xff)
    c.mmu.write_b(c.reg16["PC"]+2,0x00)    
    c.mmu.write_w(0x00ff,0x1122)
    

    c.ops[0xFA](c) //LD A,(nn)

    if (c.reg8["A"] !=0x22) {
        T.Error("Fail for LD r,nn",c.reg8,c.mmu.mem[0xff00])
    }


    
    c.mmu.write_b(c.reg16["PC"]+1,0x00)
    c.mmu.write_b(c.reg16["PC"]+2,0xff)
    c.mmu.write_w(0xff00,0x1122)
   
    c.ops[0xFA](c) //LD A,(nn)

    if (c.reg8["A"] !=0x22) {
        T.Error("Fail for LD r,nn",c.reg16,c.mem[0:6])
    }
    

   
    
    c.mmu.write_b(c.reg16["PC"]+1,0xff)
    c.mmu.write_b(c.reg16["PC"]+2,0xf0)
    c.mmu.write_w(0xf0ff,0x2211)

    c.ops[0xFA](c) //LD A,(nn)

    if (c.reg8["A"] !=0x11) {
        T.Error("Fail for LD r,nn",c.reg16,c.mem[0:6])
    }

   
}
func TestLD_SP_nn(T *testing.T){
   
    c:= NewCpu()

 

   
    
    c.mmu.write_b(c.reg16["PC"]+1,0xff)
    c.mmu.write_b(c.reg16["PC"]+2,0xf0)
    c.mmu.write_w(0xf0ff,0x2211)
    c.reg16["SP"] =0x2211

    c.ops[0x08](c) //LD A,(nn)

    if (c.mmu.read_w(0xf0ff)!=0x2211) {
        T.Error("Fail for LD SP,nn",c.reg16)
    }
    c= NewCpu()
    
    c.mmu.write_b(c.reg16["PC"]+1,0xff)
    c.mmu.write_b(c.reg16["PC"]+2,0x00)
    c.mmu.write_w(0x00ff,0x2211)
    c.reg8["A"] =0x11

    c.ops[0xEA](c) //LD A,(nn)

    if (c.mmu.read_b(0x00ff)!=0x11) {
        T.Error("Fail for LD SP,nn",c.reg8,c.mmu.read_b(0x00ff))
    }
   
}


func TestLD_SP_n(T *testing.T){
   
    c:= NewCpu()

    
    c.mmu.write_b(c.reg16["PC"]+1,124)
    c.mmu.write_w(0xff,0x2211)
    c.reg16["SP"] =1

    c.ops[0xf8](c) //LD SP_nn

    if (c.reg16["SP"]!=124) {
        T.Error("Fail for LD SP,nn",c.reg16,c.reg8)
    }
    c.mmu.write_b(c.reg16["PC"]+1,0xfb)
    c.mmu.write_w(0xff,0x2211)
    c.reg16["SP"] =0

    c.ops[0xf8](c) //LD A,(nn)


    if c.reg8["H"]!=0xff ||c.reg8["L"]!=0xfb   {
        T.Error("Fail for LD SP,nn",c.reg16,c.reg8)
    }




    c= NewCpu()

}
    
func Test_ADD(T *testing.T){

    c:= NewCpu()


    
    var i uint8
    for i=1; i<0xff; i++ {
        c.reg8["B"] = i
        c.reg8["A"] = 0
            
        c.ops[0x80](c)

        if c.reg8["A"]!=i || c.reg8["FL"] != 0  {
            T.Error("Fail for ADD A,B",c.reg8)
        }
    }
    c.reg8["B"] = 0xff
    c.reg8["A"] = 1

    c.ops[0x80](c)

    if c.reg8["A"]!=0 || c.reg8["FL"] != 0x90  {
        T.Error("Fail for ADD A,B",c.reg8)
    }

    
    c.reg8["B"] = 0xff
    c.reg8["A"] = 2

    c.ops[0x80](c)

    if c.reg8["A"]==0 || c.reg8["FL"] != 0x10  {
        T.Error("Fail for ADD A,B",c.reg8)
    }

    c.reg8["B"] = 0x0
    c.reg8["A"] = 0

    c.ops[0x80](c)

    if c.reg8["A"]!=0 || c.reg8["FL"] != 0x80  {
        T.Error("Fail for ADD A,B",c.reg8)
    }
    
    c.reg8["H"] = 0xde
    c.reg8["L"] = 0xad
    c.reg8["A"] = 0
    c.mmu.write_w(0xdead,0xff)
    c.ops[0x86](c)

    if c.reg8["A"]!=0xff || c.reg8["FL"] != 0x0  {
        T.Error("Fail for ADD A,B",c.reg8)
    }

    


}   
func Test_SUB(T *testing.T){

    c:= NewCpu()


    
    var i uint8
    for i=1; i<0xff; i++ {
        c.reg8["B"] = i
        c.reg8["A"] = 0xff
            
        c.ops[0x90](c)

        if c.reg8["A"]!=0xff-i || c.reg8["FL"] != 0x40  {
            T.Error("Fail for SUB A,B",c.reg8)
        }
    }
  
     c.reg8["B"] = 1
    c.reg8["A"] = 0

    c.ops[0x90](c)

    if c.reg8["A"]!=0xff || c.reg8["FL"] != 0x50  {
        T.Error("Fail for SUB A,B",c.reg8)
    }

  
    c.reg8["B"] = 0x0
    c.reg8["A"] = 0

    c.ops[0x90](c)

    if c.reg8["A"]!=0 || c.reg8["FL"] != 0xc0  {
        T.Error("Fail for SUB A,B",c.reg8)
    }

    
}    




func Test_PUSH(T *testing.T){
   
    c:= NewCpu()

    c.reg8["H"] = 0xde
    c.reg8["L"] = 0xad
    c.reg16["SP"]=10

    
    for c.reg16["SP"]>0 {
        c.ops[0xE5](c) //LD A,(nn)  
   if  (c.mmu.read_w(c.reg16["SP"]) != 0xdead)  {
        c.Print_dump()
        T.Error("Fail for PUSH HL",c.reg16,c.mmu.read_w(c.reg16["SP"]+2))
    } 
}
    for c.reg16["SP"]!=10 {
        c.ops[0xC1](c) //LD A,(nn) 
        c.reg8["B"] = 0
        c.reg8["C"] = 0
        c.Print_dump()
   if  (c.reg8["B"] ==  0xde && c.reg8["B"] ==  0xad)  {
        c.Print_dump()
        T.Error("Fail for POP BC",c.reg16,c.mmu.read_w(c.reg16["SP"]-2))
    } 

}



}


func Test_INC_DEC(T *testing.T){
    c:= NewCpu()
    c.reg8["B"] = 0x1
    c.ops[0x04](c)
    if  c.reg8["B"] != 0x2  {
            T.Error("Fail for INC B",c.reg8)
        }

    c.ops[0x05](c)
    
    
    if  c.reg8["B"] != 0x1  {
            T.Error("Fail for DEC B",c.reg8)
        }

    
    c.reg8["D"] = 0x33
    c.reg8["E"] = 0x01
    c.ops[0x13](c)    
    if  c.reg8["E"] != 0x2  {
               c.Print_dump()
            T.Error("Fail for INC DE")
        }




}

func Test_CP(T *testing.T){

    c:= NewCpu()


    
    var i uint8
    for i=1; i<0xff; i++ {
        c.reg8["B"] = i
        c.reg8["A"] = 0xff
            
        c.ops[0xb8](c)

        if  c.reg8["FL"] != 0x40  {
            T.Error("Fail for SUB A,B",c.reg8)
        }
    }
     c.reg16["PC"] = 0  
     c.reg8["B"] = 1
    c.reg8["A"] = 0

    c.ops[0xB8](c)

    if c.reg8["A"]!=0x0 || c.reg8["FL"] != 0x50  {
        T.Error("Fail for SUB A,B",c.reg8)
    }

    c.reg16["PC"] = 0  
    c.reg8["B"] = 0x0
    c.reg8["A"] = 0

    c.ops[0xB8](c)

    if c.reg8["A"]!=0 || c.reg8["FL"] != 0xc0  {
        T.Error("Fail for SUB A,B",c.reg8)
    }
    
}    
