package cpu

import (
        "fmt"
        "os"
        )

 const (
        reg8  = iota
        reg16
        reg16_combo
        regc
        reghld
        reghli
        regldh
        memreg
        memnn
        memn
        nn
        n
        invalid
    )    


type Action func(*CPU)
type SetVal func(*CPU,uint16) 
type GetVal func(*CPU) uint16
type OpMap map[uint16] Action 
type RegMap8 map[string] uint8
type RegMap16 map[string] uint16
type Memory [0x10000]uint8


type CPU struct {
    ops OpMap
    reg8 RegMap8
    reg16 RegMap16
    mem Memory
    mmu MMU
}
//DO the thang

func (c* CPU)load_bios() {

    fi, err := os.Open("GB_BIOS.bin")
    if err != nil { panic(err) }    

    buf := make([]uint8, 256)

    fi.Read(buf)

    fmt.Println(buf)
    var i uint16
    for i=0; i<256; i++ {
        c.mmu.write_b(i,buf[i])
    }
    fmt.Println(c.mmu.mem[0:256])

}

func (c *CPU)Exec() () {

    c.load_bios()
    var op uint16
    for {

            op = uint16(c.mmu.read_w(c.reg16["PC"]))
            fmt.Printf("0x%X\n",op)
            if op&0x00ff != 0xcb{
                op&=0xff
                
            } else{
               op=0xcb00| ((op & 0xff00) >>8)
            }

            //fmt.Printf("0x%04X\n",op)
            c.ops[op](c)
        
            fmt.Println(c.reg8,c.reg16)
        }    

}

func (c *CPU)tick(val uint16) () {
    //fmt.Println("tick",val);
}



func get_ld_type(arg string) (int) {
    //possible types that we can accOBept
   
    var arg_type int
    
    switch  {
        case arg == "(HLD)":
            arg_type = reghld
        case arg == "(HLI)":
            arg_type = reghli
         case arg == "(LDH)":
            arg_type = regldh
        case arg == "(nn)":
            arg_type = memnn
        case arg == "(n)":
            arg_type = memn
        case len(arg) == 4:
            arg_type = memreg       
        case arg == "n":
            arg_type = n
         case (len(arg) == 1 && arg != "n"):
            arg_type = reg8
        case arg == "nn":
            arg_type = nn
        case (arg == "PC" || arg == "SP"):
            arg_type = reg16

        case len(arg) == 2:
            arg_type = reg16_combo
        case arg == "(C)":
            arg_type = regc
        default:
            arg_type = invalid
    }   
    
    return arg_type

}

func (c *CPU)do_instr (desc string ,ticks uint16,args uint16) {

    c.tick(ticks)
    c.reg16["PC"]+=args
    fmt.Println(desc,ticks,args,c.reg16["PC"])
}


func gen_set_val(a_type int,reg string) (SetVal) {
   lambda := func (c *CPU,val uint16) {}
    switch (a_type) {
        case reg8:
            lambda = func (c *CPU,val uint16) { fmt.Println(reg,val);c.reg8[reg] = uint8(val) }
        case regc:
            lambda = func (c *CPU,val uint16) {
                c.mmu.write_b(uint16(0xff00 |uint16(c.reg8["C"])),uint8(val))
            }
        case reg16_combo:
            lambda = func (c *CPU,val uint16) {
                c.reg8[string(reg[0])] = uint8(val & 0xff00>>8)
                c.reg8[string(reg[1])] = uint8(val &0x00ff)
            }
         case reg16:
            lambda = func (c *CPU,val uint16) {
                c.reg16[reg]=val
             }

        case memreg:
            lambda = func (c *CPU,val uint16) {
                reg_high := c.reg8[string(reg[1])]
                reg_low  := c.reg8[string(reg[2])]
                addr := (uint16(reg_high) <<8) | uint16(reg_low)
                c.mmu.write_w(addr,val)
             }


        case memn:
            lambda = func (c *CPU,val uint16) {
                addr :=0xff00|uint16(c.mmu.read_b(c.reg16["PC"]+1))
                c.mmu.write_b(addr,uint8(val))
            
            }
        case reghli:
            lambda = func (c *CPU,val uint16) {
                reg_high := c.reg8[string(reg[1])]
                reg_low  := c.reg8[string(reg[2])]
                addr := (uint16(reg_high)) <<8 | uint16(reg_low)
                addr++

                c.reg8[string(reg[1])] = uint8(addr &0xff00 >>8)
                c.reg8[string(reg[2])] = uint8(addr &0x00ff)
                fmt.Printf("0x%x,0x%x\n",addr,reg_low)                        
                fmt.Println(reg,c.reg8,addr,val)
                c.mmu.write_b(addr-1,uint8(val))
            }
        case reghld:
            lambda = func (c *CPU,val uint16) {
                reg_high := c.reg8[string(reg[1])]
                reg_low  := c.reg8[string(reg[2])]
                addr := (uint16(reg_high)) <<8 | uint16(reg_low)
                addr--

                c.reg8[string(reg[1])] = uint8(addr &0xff00 >>8)
                c.reg8[string(reg[2])] = uint8(addr &0x00ff)

                
                c.mmu.write_b(addr+1,uint8(val))
            }
        default:
            lambda = func (c *CPU,val uint16) {
            fmt.Println("UNHANDLED Set  ERROR",reg)} 

    }

    return lambda
}

func gen_get_val(a_type int,reg string) (GetVal) {
 
    lambda := func (c *CPU) (uint16){return 0}
  
    switch (a_type) {    
        case reg8:
            lambda = func (c *CPU) (uint16) { fmt.Println("Get reg8");return uint16(c.reg8[reg]) } 
        case memn:
           lambda = func (c *CPU) (uint16) {
                addr :=uint16(0xff00| uint16(c.mmu.read_b(c.reg16["PC"]+1)))
                return c.mmu.read_w(addr)
            }
           
        case regc:
           lambda = func (c *CPU) (uint16) { 
                addr :=uint16(0xff00 |uint16(c.reg8["C"]))
                return uint16(c.mmu.read_b(addr))
            }

        case memreg: 
           lambda = func (c *CPU) (uint16) { 
                reg_high := c.reg8[string(reg[1])]
                reg_low  := c.reg8[string(reg[2])]
                addr := (uint16(reg_high) <<8) | uint16(reg_low)

                return c.mmu.read_w(addr)
           }
       
        case reg16_combo:
           lambda = func (c *CPU) (uint16) {
                var reg_high uint8 =  c.reg8[string(reg[0])]
                var reg_low uint8 =   c.reg8[string(reg[1])]
                fmt.Println("Get reg16 combo")
                return uint16(reg_high)<<8 | uint16(reg_low)
               
            }

        case memnn:
           lambda = func (c *CPU) (uint16) {
                addr :=  uint16(c.mmu.read_b(c.reg16["PC"]+2)) <<8 | uint16(c.mmu.read_b(c.reg16["PC"]+1))
                return c.mmu.read_w(addr)
            }
        case n:
          lambda = func (c *CPU) (uint16) {
                return uint16(c.mmu.read_b(c.reg16["PC"]+1))
          }
        case nn:
           lambda = func (c *CPU) (uint16) {
                return uint16(c.mmu.read_b(c.reg16["PC"]+2)) <<8 | uint16(c.mmu.read_b(c.reg16["PC"]+1))
            }      
       
        case reghld :
            lambda = func (c *CPU) (uint16) {
                //parse reg_right
                reg_high := c.reg8[string(reg[1])]
                reg_low  := c.reg8[string(reg[2])]

                addr := (uint16(reg_high) <<8) | uint16(reg_low)

                
                addr--
                c.reg8[string(reg[1])] = uint8(addr &0xff00 >>8)
                c.reg8[string(reg[2])] = uint8(addr &0x00ff)
               
                
                return uint16(c.mmu.read_w(addr+1)) 
               
           }

       
       case reghli :
            //parse reg_right
             lambda = func (c *CPU) (uint16) {
                reg_high := c.reg8[string(reg[1])]
                reg_low  := c.reg8[string(reg[2])]
                addr := (uint16(reg_high) <<8) | uint16(reg_low)
                addr++
                c.reg8[string(reg[1])] = uint8(addr &0xff00 >>8)
                c.reg8[string(reg[2])] = uint8(addr &0x00ff)

                return uint16(c.mmu.read_b(addr-1))
            }

    }
  return lambda
}

func (c *CPU) addr_type(a_type int,reg_right string) (uint8,uint16) {

    var val_right8 uint8=0
    var val_right16 uint16=0

    switch (a_type) {
        case reg8: 
            val_right8 = c.reg8[reg_right]
            val_right16= uint16(val_right8)

        case reg16_combo:
             var reg_high uint8 =  c.reg8[string(reg_right[0])]
            var reg_low uint8 =   c.reg8[string(reg_right[1])]            
            val_right16 = uint16(reg_high)<<8 | uint16(reg_low)
      
        case reg16:
                        
            val_right16 = c.reg16[reg_right]
        case memreg: 
            reg_high := c.reg8[string(reg_right[1])]
            reg_low  := c.reg8[string(reg_right[2])]
            addr := (uint16(reg_high) <<8) | uint16(reg_low)

            val_right16=c.mmu.read_w(addr)
            val_right8=uint8(val_right16)
            
        case reghld :
            //parse reg_right
            reg_high := c.reg8[string(reg_right[1])]
            reg_low  := c.reg8[string(reg_right[2])]
            addr := (uint16(reg_high) <<8) | uint16(reg_low)
            
            val_right16=c.mmu.read_w(addr)
            val_right8=uint8(val_right16)

            if (uint16(c.reg8[string(reg_right[1])]) <<8 | uint16(reg_low)) > 255 {
                c.reg8[string(reg_right[1])]--
            } else {
                c.reg8[string(reg_right[2])]--
            } 


       
       case reghli :
            //parse reg_right
            reg_high := c.reg8[string(reg_right[1])]
            reg_low  := c.reg8[string(reg_right[2])]
            addr := (uint16(reg_high) <<8) | uint16(reg_low)
            
            val_right16=c.mmu.read_w(addr)
            val_right8=uint8(val_right16)

            if (uint16(c.reg8[string(reg_right[1])]) <<8 | uint16(reg_low)) > 255 {
                c.reg8[string(reg_right[1])]++
            } else {
                c.reg8[string(reg_right[2])]++
            } 
           

    
        case memnn:
            //addr := uint16(c.mem[c.reg16["PC"]+2] &0xff <<8) | uint16(c.mem[c.reg16["PC"]+1]&0xff)
            addr :=  uint16(c.mmu.read_b(c.reg16["PC"]+2)) <<8 | uint16(c.mmu.read_b(c.reg16["PC"]+1))
            val_right16 = c.mmu.read_w(addr)

            val_right8 = uint8(val_right16 &0xff)            
            
        case nn:
            //addr :=uint16(c.mem[c.reg16["PC"]+2] &0xff <<8)  | uint16(c.mem[c.reg16["PC"]+1]&0xff)
            val_right16 =  uint16(c.mmu.read_b(c.reg16["PC"]+2)) <<8 | uint16(c.mmu.read_b(c.reg16["PC"]+1))
            val_right8 = uint8(val_right16)
        case n:
            val_right16 = uint16(c.mmu.read_b(c.reg16["PC"]+1))
            val_right8 = uint8(val_right16)

       case memn:
            addr :=uint16(0xff00| uint16(c.mmu.read_b(c.reg16["PC"]+1)))

            val_right16 = c.mmu.read_w(addr)
            val_right8 = uint8(val_right16)
           
                         
        case regc:
            addr :=uint16(0xff00 |uint16(c.reg8["C"]))
            val_right16 = c.mmu.read_w(addr)
            
            val_right8 = uint8(val_right16)
            
        
        }
        return val_right8,val_right16

}

func (c *CPU) fz( i uint8, as uint8)() {
    c.reg8["FL"]=0
   
    if i == 0 {
       
        c.reg8["FL"]|=0x80
    }
 
    if as == 1 {
        c.reg8["FL"]|=0x40
    }
  
}
func gen_alu(op_type string,reg_left string, reg_right string,ticks uint16, args uint16 ) (Action) {
    type_right:=get_ld_type(reg_right)
    type_left:=get_ld_type(reg_left)

    desc := op_type+" "+reg_left +","+reg_right

    var lambda Action = nil
    undefined :=  func(c *CPU) {fmt.Println("Undefined ALU op" +reg_left+ ","+reg_right)}   
    var val_right8 uint8
    var val_right16 uint16

    switch (op_type) {
        case "ADD":
            lambda = func(c *CPU)  {
                                val_right8,val_right16 = c.addr_type(type_right,reg_right)
                                prev:= c.reg8["A"]
                                c.reg8["A"] +=val_right8
                                c.fz(c.reg8["A"],0)

                                if prev > c.reg8["A"] {
                                    c.reg8["FL"] |= 0x10
                                }
                            
                                c.do_instr(desc,ticks,args)
                            }
    
        case "SUB":
            lambda = func(c *CPU)  {
                                val_right8,val_right16 = c.addr_type(type_right,reg_right)
                                prev:= c.reg8["A"]
                                c.reg8["A"] -=val_right8
                                c.fz(c.reg8["A"],1)

                                if prev < c.reg8["A"] {
                                    c.reg8["FL"] |= 0x10
                                }

                                c.do_instr(desc,ticks,args)
                            }
        case "CP":
            lambda = func(c *CPU)  {
                                val_right8,val_right16 = c.addr_type(type_right,reg_right)
                                i:= c.reg8["A"] - val_right8
                                c.fz(i,1)

                                if i > c.reg8["A"] {
                                    c.reg8["FL"] |= 0x10
                                }

                                c.do_instr(desc,ticks,args)
                            }



    

        case "SBC":
            lambda = func(c *CPU)  {
                                val_right8,val_right16 = c.addr_type(type_right,reg_right)
                                prev:= c.reg8["A"]
                                c.reg8["A"] -=val_right8
                                //subtract carray flag
                                c.reg8["A"] -=0x10 &c.reg8["FL"] 
                                c.fz(c.reg8["A"],1)

                                if prev < c.reg8["A"] {
                                    c.reg8["FL"] |= 0x10
                                }

                                c.do_instr(desc,ticks,args)
                            }
        case "AND":
            lambda = func(c *CPU)  {
                                val_right8,val_right16 = c.addr_type(type_right,reg_right)
                                c.reg8["A"] &=val_right8
                                c.fz(c.reg8["A"],1)
                                c.do_instr(desc,ticks,args)
                            }

        case "OR":
            lambda = func(c *CPU)  {
                                val_right8,val_right16 = c.addr_type(type_right,reg_right)
                                c.reg8["A"] |=val_right8
                                c.fz(c.reg8["A"],0)
                                c.do_instr(desc,ticks,args)
                            }

        case "XOR":
            lambda = func(c *CPU)  {
                                val_right8,val_right16 = c.addr_type(type_right,reg_right)
                                c.reg8["A"] ^=val_right8
                                c.fz(c.reg8["A"],0)
                                c.do_instr(desc,ticks,args)
                            }


       case "INC":
            lambda = func(c *CPU)  {
                                val_right8,val_right16 = c.addr_type(type_left,reg_left)

                                if len(reg_left) != 2{

                                    c.reg8[reg_left]++
                                    c.fz(c.reg8[reg_left],0)
                                    c.do_instr(desc,ticks,args)
                                } else {
                                    var high string = string(reg_left[0])
                                    var low string =string(reg_left[1])
                                    val :=(uint16(c.reg8[high]) <<8) |uint16(c.reg8[low])
                                    
                                    val++
                                    c.reg8[high] = uint8(val & 0xff00 >> 8)
                                    c.reg8[low] = uint8(val & 0xff)
                                    c.do_instr(desc,ticks,args)
                                   }
                                 }
      case "DEC":
            lambda = func(c *CPU)  {
                                val_right8,val_right16 = c.addr_type(type_right,reg_right)
       
                                if len(reg_left) != 2{
                                    c.reg8[reg_left]--
                                    c.fz(uint8(c.reg8[reg_left]),1)
                                    c.do_instr(desc,ticks,args)
                                }else {
                                    var high string = string(reg_left[0])
                                    var low string =string(reg_left[1])
                                    val :=uint16(c.reg8[high]) <<8 |uint16(c.reg8[low])
                                    val--
                                    
                                    c.reg8[high] = uint8(val & 0xff00 >> 8)
                                    c.reg8[low] = uint8(val & 0xff)
                                    c.do_instr(desc,ticks,args)
                                 }
                               }


    }




    if lambda != nil {
        //fmt.Println("Created "+left+" "+reg_right)
        return lambda
    }else{
        lambda = undefined
        fmt.Println("UNDEFINED PUSH/POP " +reg_left+ " " +reg_right)
    }
    return lambda



}

func gen_push_pop(left string ,reg_right string) (Action)  {
    type_right:=get_ld_type(reg_right)
       
    desc := left +","+reg_right

    var lambda Action = nil
    var val_right8 uint8
    var val_right16 uint16
    undefined :=  func(c *CPU) {fmt.Println("Undefined PUSH/POP op" +left+ ","+reg_right)}   

    if left == "PUSH" {
       
            lambda = func(c *CPU)  {
                    val_right8,val_right16 = c.addr_type(type_right,reg_right)
                    c.mmu.write_w(c.reg16["SP"],val_right16)
                    c.reg16["SP"]-=2
                    c.do_instr(desc,20,1)
                   
            }
    } else if left == "POP" {
             lambda = func(c *CPU)  {
  
                    c.reg8[string(reg_right[0])] = uint8(0xff00 & c.mmu.read_w(c.reg16["SP"]))
                    c.reg8[string(reg_right[1])] = uint8(0x00ff & c.mmu.read_w(c.reg16["SP"]))
                    c.do_instr(desc,20,1)
                    c.reg16["SP"]+=2
            }
    }
    
    if lambda != nil {
        //fmt.Println("Created "+left+" "+reg_right)
        return lambda
    }else{
        lambda = undefined
        fmt.Println("UNDEFINED PUSH/POP " +left+ " " +reg_right)
    }
    return lambda

}

func gen_ld(reg_left string, reg_right string,ticks uint16,args uint16) (Action)  {
    
    type_left:=get_ld_type(reg_left)
    type_right:=get_ld_type(reg_right)

    undefined :=  func(c *CPU) {fmt.Println("Undefined LD op" +reg_left+ ","+reg_right)}   
    var lambda Action = nil

    desc := "LD "+reg_left+","+reg_right
    
    
    //var val_right16 uint16;
    //var val_right8 uint8;
 

    f_get_val :=gen_get_val(type_right,reg_right)
    f_set_val :=gen_set_val(type_left,reg_left)

    lambda = func(c *CPU)  {
                f_set_val(c,f_get_val(c))
                c.do_instr(desc,(ticks),(args))
    }
/*
    } else if type_left == memreg {
        //parse reg_right
        var reg_high string = string(reg_left[1])
        var reg_low string =  string(reg_left[2])
        lambda =  func(c *CPU)  {
                val_right8,val_right16 = c.addr_type(type_right,reg_right)
                addr := (uint16(c.reg8[reg_high]) <<8) | uint16(c.reg8[reg_low])        
                c.mmu.write_w(addr,val_right16)
	 
		c.do_instr(desc,ticks,args)
        } 
       
    } else if type_left == memnn {
        lambda = func(c *CPU) {
                
                val_right8,val_right16 = c.addr_type(type_right,reg_right)
                addr :=  uint16(c.mmu.read_b(c.reg16["PC"]+2)) <<8 | uint16(c.mmu.read_b(c.reg16["PC"]+1))
                
                c.mmu.write_w(addr, val_right16)
                c.do_instr(desc,ticks,args)
        }
    }else if type_left == memn {
        lambda = func(c *CPU) {
                val_right8,val_right16 = c.addr_type(type_right,reg_right)
                addr :=0xff00|uint16(c.mmu.read_b(c.reg16["PC"]+1))
            
                c.mmu.write_b(addr,val_right8)
                c.do_instr(desc,ticks,args)
        }


    } else if type_left == regc {
        lambda = func(c *CPU) {
                val_right8,val_right16 = c.addr_type(type_right,reg_right)
    
                c.mmu.write_w(uint16(0xff00 |uint16(c.reg8["C"])),val_right16)

                c.do_instr(desc,ticks,args)
        }
    } else if type_left == reghld {
        var reg_high string = string(reg_left[1])
        var reg_low string =  string(reg_left[2])
        f:= gen_alu("DEC","HL","",0,0)
       lambda = func(c *CPU) {
                val_right8,val_right16 = c.addr_type(type_right,reg_right)
                
                addr := (uint16(c.reg8[reg_high]) <<8) | uint16(c.reg8[reg_low])        
                c.mmu.write_b(addr,val_right8)

                c.do_instr(desc,ticks,args)
                f(c)
        }
           
    } else if type_left == reghli { 
        var reg_high string = string(reg_left[1])
        var reg_low string =  string(reg_left[2])
      
       lambda = func(c *CPU) {
                val_right8,val_right16 = c.addr_type(type_right,reg_right)
                
                addr := (uint16(c.reg8[reg_high]) <<8) | uint16(c.reg8[reg_low])        
                c.mmu.write_b(addr,val_right8)

                if uint16(c.reg8[reg_high]) <<8 |  uint16(c.reg8[reg_low])  > 255 {
                    c.reg8[reg_right]++
                } else {
                    c.reg8[reg_low]++
                } 
                c.do_instr(desc,ticks,args)
        }


     } else if type_left == reg16_combo { 
        var reg_high string = string(reg_left[0])
        var reg_low string =  string(reg_left[1])
      
       lambda = func(c *CPU) {
                
                val_right8,val_right16 = c.addr_type(type_right,reg_right)
                
                c.reg8[reg_high] = uint8(val_right16 & 0xff00>>8)
                c.reg8[reg_low] = uint8(val_right16 &0x00ff)
                c.do_instr(desc,ticks,args)


        }
    } else if reg_left == "SP" && reg_right =="n" { 
        lambda = func(c *CPU) {
                
                val_right8,val_right16 = c.addr_type(type_right,reg_right)
                if(val_right16>127){ 
                    val_right16=-((^val_right16+1)&0xff); 
                }

                c.reg16["SP"] += val_right16
                c.reg8["H"] =  uint8(val_right16 & 0xff00>>8)
                c.reg8["L"] = uint8(val_right16 &0x00ff)
        
                c.do_instr(desc+" special",ticks,args)
        }
    }  else if type_left == reg16  {
        lambda = func(c *CPU)  {
                val_right8,val_right16 = c.addr_type(type_right,reg_right)
                c.reg16[reg_left] = val_right16
                c.do_instr(desc,(ticks),(args))
        }
    }

*/

   if lambda != nil {
        //fmt.Println("Created LD "+reg_left+","+reg_right)
    	return lambda
    }else{
        lambda = undefined
        fmt.Println("UNDEFINED LD "+reg_left+","+reg_right)
    }
    return lambda
}



func NewCpu() *CPU{
    return BuildCpu()
}

func BuildCpu() *CPU{
    c := new (CPU)

    c.reg8 = make(RegMap8)
    c.reg16 = make(RegMap16)
    c.ops = make(OpMap)
 
    //Init registers
    /////////////////
    c.reg8["A"] = 0
    c.reg8["B"] = 0
    c.reg8["C"] = 0
    c.reg8["D"] = 0
    c.reg8["E"] = 0
    c.reg8["F"] = 0 
    c.reg8["H"] = 0 
    c.reg8["L"] = 0
    c.reg8["FL"] = 0
    c.reg16["SP"]=0
    c.reg16["PC"]=0
    /////////////////

    //Generate opcode Map
    c.ops[0x7f] = gen_ld("A","A",4,1)
    c.ops[0x78] = gen_ld("A","B",4,1)
    c.ops[0x77] = gen_ld("A","C",4,1)
    c.ops[0x7A] = gen_ld("A","D",4,1)
    c.ops[0x7B] = gen_ld("A","E",4,1)
    c.ops[0x7C] = gen_ld("A","H",4,1)
    c.ops[0x7D] = gen_ld("A","L",4,1)
    c.ops[0x0A] = gen_ld("A","(BC)",8,1)
    c.ops[0x1A] = gen_ld("A","(DE)",8,1)
    c.ops[0x7E] = gen_ld("A","(HL)",8,1)
    c.ops[0xFA] = gen_ld("A","(nn)",8,3)
    c.ops[0x3E] = gen_ld("A","n",8,2)
    c.ops[0xF2] = gen_ld("A","(C)",8,1)
    
    
    c.ops[0x06] = gen_ld("B","n",8,2)
    c.ops[0x40] = gen_ld("B","B",4,1)
    c.ops[0x41] = gen_ld("B","C",4,1)
    c.ops[0x42] = gen_ld("B","D",4,1)    
    c.ops[0x43] = gen_ld("B","E",4,1)
    c.ops[0x44] = gen_ld("B","H",4,1)
    c.ops[0x45] = gen_ld("B","L",4,1)
    c.ops[0x46] = gen_ld("B","(HL)",8,1)
    c.ops[0x47] = gen_ld("B","A",4,1)    

    c.ops[0x48] = gen_ld("C","B",4,1)
    c.ops[0x49] = gen_ld("C","C",4,1)
    c.ops[0x4A] = gen_ld("C","D",4,1)
    c.ops[0x4B] = gen_ld("C","E",4,1)
    c.ops[0x4C] = gen_ld("C","H",4,1)
    c.ops[0x4D] = gen_ld("C","L",4,1)
    c.ops[0x4E] = gen_ld("C","(HL)",8,1)    
    c.ops[0x4F] =  gen_ld("C","A",4,1)    
    c.ops[0x0E] =  gen_ld("C","n",8,2)    
    c.ops[0x50] = gen_ld("D","B",4,1)
    c.ops[0x51] = gen_ld("D","C",4,1)
    c.ops[0x52] = gen_ld("D","D",4,1)
    c.ops[0x53] = gen_ld("D","E",4,1)
    c.ops[0x54] = gen_ld("D","H",4,1)
    c.ops[0x55] = gen_ld("D","L",4,1)
    c.ops[0x56] = gen_ld("D","(HL)",8,1)    
    c.ops[0x57] = gen_ld("D","A",4,1)    
    c.ops[0x16] =  gen_ld("D","E",4,1)    
    
    c.ops[0x58] = gen_ld("E","B",4,1)
    c.ops[0x59] = gen_ld("E","C",4,1)
    c.ops[0x5A] = gen_ld("E","D",4,1)
    c.ops[0x5B] = gen_ld("E","E",4,1)
    c.ops[0x5C] = gen_ld("E","H",4,1)
    c.ops[0x5D] = gen_ld("E","L",4,1)
    c.ops[0x5E] = gen_ld("E","(HL)",8,1)
    c.ops[0x5F] = gen_ld("E","A",4,1)        
    c.ops[0x1E] = gen_ld("E","n",8,2)        
    
    c.ops[0x60] = gen_ld("H","B",4,1)
    c.ops[0x61] = gen_ld("H","C",4,1)
    c.ops[0x62] = gen_ld("H","D",4,1)
    c.ops[0x63] = gen_ld("H","E",4,1)
    c.ops[0x64] = gen_ld("H","H",4,1)
    c.ops[0x65] = gen_ld("H","L",4,1)
    c.ops[0x66] = gen_ld("H","(HL)",8,1)
    c.ops[0x67] = gen_ld("H","A",4,1)
    c.ops[0x26] = gen_ld("H","n",8,2)        
    
    c.ops[0x68] = gen_ld("L","B",4,1)
    c.ops[0x69] = gen_ld("L","C",4,1)
    c.ops[0x6A] = gen_ld("L","D",4,1)
    c.ops[0x6B] = gen_ld("L","E",4,1)
    c.ops[0x6C] = gen_ld("L","H",4,1)
    c.ops[0x6D] = gen_ld("L","L",4,1)
    c.ops[0x6E] = gen_ld("L","(HL)",8,1)
    c.ops[0x6F] = gen_ld("L","A",4,1)
    c.ops[0x2E] = gen_ld("L","n",8,2)   
 
    c.ops[0x70] = gen_ld("(HL)","B",8,1)
    c.ops[0x71] = gen_ld("(HL)","C",8,1)
    c.ops[0x72] = gen_ld("(HL)","D",8,1)
    c.ops[0x73] = gen_ld("(HL)","E",8,1)
    c.ops[0x74] = gen_ld("(HL)","H",8,1)
    c.ops[0x75] = gen_ld("(HL)","L",8,1)
    c.ops[0x77] = gen_ld("(HL)","A",8,1)
    c.ops[0x36] = gen_ld("(HL)","n",8,1)
    
    c.ops[0x02] = gen_ld("(BC)","A",8,1)
    c.ops[0x12] = gen_ld("(DE)","A",8,1)
    c.ops[0xEA] = gen_ld("(nn)","A",16,3)
    c.ops[0xE2] = gen_ld("(C)","A",8,1)
    c.ops[0x3A] = gen_ld("A","(HLD)",12,1)
    c.ops[0x32] = gen_ld("(HLD)","A",12,1)
    c.ops[0x2A] = gen_ld("A","(HLI)",12,1)
    c.ops[0x22] = gen_ld("(HLI)","A",12,1)
    c.ops[0xE0] = gen_ld("(n)","A",8,2)
    c.ops[0xF0] = gen_ld("A","(n)",8,2)

    c.ops[0x01] = gen_ld("BC","nn",12,3)
    c.ops[0x11] = gen_ld("DE","nn",12,3)
    c.ops[0x21] = gen_ld("HL","nn",12,3)
    c.ops[0x31] = gen_ld("SP","nn",12,3)
    c.ops[0xf8] = gen_ld("SP","n",12,2) //signed
    c.ops[0xf9] = gen_ld("SP","HL",8,1)
    c.ops[0x08] = gen_ld("(nn)","SP",20,3)
    c.ops[0xf5] = gen_push_pop("PUSH","AF") 
    c.ops[0xC5] = gen_push_pop("PUSH","BC") 
    c.ops[0xD5] = gen_push_pop("PUSH","DE") 
    c.ops[0xE5] = gen_push_pop("PUSH","HL") 
    
    c.ops[0xF1] = gen_push_pop("POP","AF") 
    c.ops[0xC1] = gen_push_pop("POP","BC") 
    c.ops[0xD1] = gen_push_pop("POP","DE") 
    c.ops[0xE1] = gen_push_pop("POP","HL") 

    c.ops[0x87] = gen_alu("ADD","A","A",4,1)
    c.ops[0x80] = gen_alu("ADD","A","B",4,1)
    c.ops[0x81] = gen_alu("ADD","A","C",4,1)
    c.ops[0x82] = gen_alu("ADD","A","D",4,1)
    c.ops[0x83] = gen_alu("ADD","A","E",4,1)
    c.ops[0x84] = gen_alu("ADD","A","H",4,1)
    c.ops[0x85] = gen_alu("ADD","A","L",4,1)
    c.ops[0x86] = gen_alu("ADD","A","(HL)",8,1)
    c.ops[0xc6] = gen_alu("ADD","A","n",8,2)
    
    c.ops[0x97] = gen_alu("SUB","A","A",4,1)
    c.ops[0x90] = gen_alu("SUB","A","B",4,1)
    c.ops[0x91] = gen_alu("SUB","A","C",4,1)
    c.ops[0x92] = gen_alu("SUB","A","D",4,1)
    c.ops[0x93] = gen_alu("SUB","A","E",4,1)
    c.ops[0x94] = gen_alu("SUB","A","H",4,1)
    c.ops[0x95] = gen_alu("SUB","A","L",4,1)
    c.ops[0x96] = gen_alu("SUB","A","(HL)",8,1)
    c.ops[0xD6] = gen_alu("SUB","A","n",8,2)
    
    //No tests for these
    c.ops[0x9F] = gen_alu("SBC","A","A",4,1)
    c.ops[0x99] = gen_alu("SBC","A","B",4,1)
    c.ops[0x99] = gen_alu("SBC","A","C",4,1)
    c.ops[0x9A] = gen_alu("SBC","A","D",4,1)
    c.ops[0x9B] = gen_alu("SBC","A","E",4,1)
    c.ops[0x9C] = gen_alu("SBC","A","H",4,1)
    c.ops[0x9D] = gen_alu("SBC","A","L",4,1)
    c.ops[0x9E] = gen_alu("SBC","A","(HL)",8,1)
    //c.ops[0xD6] = gen_alu("SBC","A","n",8,2)
    
    c.ops[0xA7] = gen_alu("AND","A","A",4,1)
    c.ops[0xA0] = gen_alu("AND","A","B",4,1)
    c.ops[0xA1] = gen_alu("AND","A","C",4,1)
    c.ops[0xA2] = gen_alu("AND","A","D",4,1)
    c.ops[0xA3] = gen_alu("AND","A","E",4,1)
    c.ops[0xA4] = gen_alu("AND","A","H",4,1)
    c.ops[0xA5] = gen_alu("AND","A","L",4,1)
    c.ops[0xA6] = gen_alu("AND","A","(HL)",8,1)
    c.ops[0xE6] = gen_alu("AND","A","n",8,2)

    c.ops[0xB7] = gen_alu("OR","A","A",4,1)
    c.ops[0xB0] = gen_alu("OR","A","B",4,1)
    c.ops[0xB1] = gen_alu("OR","A","C",4,1)
    c.ops[0xB2] = gen_alu("OR","A","D",4,1)
    c.ops[0xB3] = gen_alu("OR","A","E",4,1)
    c.ops[0xB4] = gen_alu("OR","A","H",4,1)
    c.ops[0xB5] = gen_alu("OR","A","L",4,1)
    c.ops[0xB6] = gen_alu("OR","A","(HL)",8,1)
    c.ops[0xF6] = gen_alu("OR","A","n",8,2)

    c.ops[0xAF] = gen_alu("XOR","A","A",4,1)
    c.ops[0xA8] = gen_alu("XOR","A","B",4,1)
    c.ops[0xA9] = gen_alu("XOR","A","C",4,1)
    c.ops[0xAA] = gen_alu("XOR","A","D",4,1)
    c.ops[0xAB] = gen_alu("XOR","A","E",4,1)
    c.ops[0xAC] = gen_alu("XOR","A","H",4,1)
    c.ops[0xAD] = gen_alu("XOR","A","L",4,1)
    c.ops[0xAE] = gen_alu("XOR","A","(HL)",8,1)
    c.ops[0xEE] = gen_alu("XOR","A","n",8,2)
    
    c.ops[0xBF] = gen_alu("CP","A","A",4,1)
    c.ops[0xB8] = gen_alu("CP","A","B",4,1)
    c.ops[0xB9] = gen_alu("CP","A","C",4,1)
    c.ops[0xBA] = gen_alu("CP","A","D",4,1)
    c.ops[0xBB] = gen_alu("CP","A","E",4,1)
    c.ops[0xBC] = gen_alu("CP","A","H",4,1)
    c.ops[0xBD] = gen_alu("CP","A","L",4,1)
    c.ops[0xBE] = gen_alu("CP","A","(HL)",8,1)
    c.ops[0xFE] = gen_alu("CP","A","n",8,2)
    
    c.ops[0x3D] = gen_alu("DEC","A","",4,1)
    c.ops[0x05] = gen_alu("DEC","B","",4,1)
    c.ops[0x0D] = gen_alu("DEC","C","",4,1)
    c.ops[0x15] = gen_alu("DEC","D","",4,1)
    c.ops[0x1D] = gen_alu("DEC","E","",4,1)
    c.ops[0x25] = gen_alu("DEC","H","",4,1)
    c.ops[0x2D] = gen_alu("DEC","L","",4,1)
    c.ops[0x35] = gen_alu("DEC","(HL)","",12,1)
    
    

    
    c.ops[0x3c] = gen_alu("INC","A","",4,1)
    c.ops[0x04] = gen_alu("INC","B","",4,1)
    c.ops[0x0c] = gen_alu("INC","C","",4,1)
    c.ops[0x14] = gen_alu("INC","D","",4,1)
    c.ops[0x1c] = gen_alu("INC","E","",4,1)
    c.ops[0x24] = gen_alu("INC","H","",4,1)
    c.ops[0x2c] = gen_alu("INC","L","",4,1)
    c.ops[0x33] = gen_alu("INC","SP","",12,1)
    c.ops[0x13] = gen_alu("INC","DE","",12,1)

    c.ops[0x34] = gen_alu("INC","(HL)","",12,1)
    c.ops[0x23] = gen_alu("INC","HL","",12,1)
    
    

    
    c.ops[0xcb7c] = func(c *CPU) {
                                                
                                  if ((c.reg8["H"] &0x80)==0){
                                    c.reg8["FL"]|= 0x80  
                                  }else{
                                   c.reg8["FL"] &= 0x7f
                                  }
                                  c.do_instr("SET B1 A",1,2)//2 for cb cmds  
}   
    c.ops[0x20] = func(c*CPU) {
                                if((c.reg8["FL"]&0x80) ==0) {
                                   n:=c.mmu.read_b(c.reg16["PC"]+1)
                                   fmt.Println(n)
                                  if ( n > 127) { n=^(n+1) }   
                                   fmt.Printf("JMP ADDR:0x%X\n",n)                            
                                   
                                   c.reg16["PC"] -= uint16(n)
                                   c.do_instr("JNZ",12,0)  

                                } else{
                                  c.do_instr("JNZ",12,2)
                                
                                }
                            }

    p_call := gen_push_pop("PUSH","PC")
    c.ops[0xCD] = func(c*CPU) {                  
                                c.reg16["PC"]+=3
                                p_call(c)
                                 c.reg16["PC"] =uint16(c.mmu.read_b(c.reg16["PC"]-2))<<8 |uint16(c.mmu.read_b(c.reg16["PC"]-1))
                                c.do_instr("CALL",4,0) 
}
     c.ops[0xCC] = func(c*CPU) {
                            if((c.reg8["FL"]&0x80) ==1) {

                                c.reg16["PC"]+=3
                                p_call(c)
                               c.reg16["PC"] =uint16(c.mmu.read_b(c.reg16["PC"]-2))<<8 |uint16(c.mmu.read_b(c.reg16["PC"]-1))
                                c.do_instr("CALL",4,0)

                            }else{
                               c.do_instr("CALL",12,3)
                            }

                              }


                                                
return c
}




