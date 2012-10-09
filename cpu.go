package main

import "fmt"

 const (
        reg8  = iota
        reg16
        regc
        memreg
        memnn
        nn
        n
        invalid
    )    


type Action func(*CPU)
type OpMap map[uint16] Action 
type RegMap8 map[string] uint8
type RegMap16 map[string] uint16
type Memory [0x10000]uint16


type CPU struct {
    ops OpMap
    reg8 RegMap8
    reg16 RegMap16
    mem Memory

}

func (c *CPU)exec() () {

    fmt.Println("Test!!!");
}

func (c *CPU)tick(val uint16) () {
    fmt.Println("tick",val);
}



func get_ld_type(arg string) (int) {
    //possible types that we can accOBept
   
    var arg_type int
    
    switch  {
           
        case arg == "(nn)":
            arg_type = memnn
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
        case arg == "(C)":
            arg_type = regc
        default:
            arg_type = invalid
    }   
    
    return arg_type

}

func (c *CPU)do_instr (desc string ,ticks uint16,args uint16) {
    fmt.Println(desc)
    c.tick(ticks)
    c.reg16["PC"]+=args
}

func (c *CPU) addr_type(a_type int,reg_right string) (uint8,uint16) {

    var val_right8 uint8=0
    var val_right16 uint16=0

    switch (a_type) {
        case reg8: 
            val_right8 = c.reg8[reg_right]
        case memreg: 
            //parse reg_right
            var reg_high uint8 = reg_right[1]
            var reg_low uint8 = reg_right[2]

            val_right16=c.mem[((reg_high <<8) | reg_low)]
            val_right8=uint8(val_right16)
            
        case memnn:
            val_right16 = c.mem[((c.reg16["PC"]+2)<<8) | c.mem[(c.reg16["PC"]+1)]]
            val_right8 = uint8(val_right16)            
        
        case n:
            val_right16 = c.mem[(c.reg16["PC"]+1)]
            val_right8 = uint8(val_right16)
                        
        case regc:
            val_right16 = c.mem[uint16(0xff00 |c.reg16["C"])]
            val_right8 = uint8(val_right16)
        }
        return val_right8,val_right16

}

func gen_ld(reg_left string, reg_right string) (Action)  {
    
    type_left:=get_ld_type(reg_left)
    type_right:=get_ld_type(reg_right)

    undefined :=  func(c *CPU) {fmt.Println("Undefined LD op" +reg_left+ ","+reg_right)}   
    var lambda Action = nil

    desc := "LD "+reg_left+","+reg_right
    
    
    var val_right16 uint16;
    var val_right8 uint8;
 
    var ticks uint16 = 0 
    var args uint16 =1
   
    if  type_right == memnn {
        ticks+=12
        args+=2

    }  else if type_right == memreg {
        ticks+=4

    } else if type_right == n {
        ticks+=4
        args+=1
    }

    if type_left == reg8  {
        ticks+=4
        lambda = func(c *CPU)  {
                val_right8,val_right16 = c.addr_type(type_right,reg_right)
                c.reg8[reg_left] = val_right8
                c.do_instr(desc,(ticks),(args))
        }
    } else if type_left == memreg {
        //parse reg_right
        var reg_high uint8 = reg_left[1]
        var reg_low uint8 = reg_left[2]
        ticks+=8
        lambda =  func(c *CPU)  {
                val_right8,val_right16 = c.addr_type(type_right,reg_right)
                c.mem[((reg_high <<8) | reg_low)] = val_right16
                c.do_instr(desc,ticks,args)
        } 
       
    } else if type_left == memnn {
        ticks+=16
        lambda = func(c *CPU) {
                val_right8,val_right16 = c.addr_type(type_right,reg_right)
                c.mem[(c.mem[(c.reg16["PC"]+2)]<<8) | c.mem[(c.reg16["PC"]+1)]] =
                                         val_right16
                c.do_instr(desc,ticks,args)
        }
    } else if type_left == regc {
        ticks+=8
        lambda = func(c *CPU) {
                val_right8,val_right16 = c.addr_type(type_right,reg_right)
                c.mem[uint16(0xff00 |c.reg16["C"])] = val_right16
                c.do_instr(desc,ticks,args)
         }
    }


   if lambda != nil {
        fmt.Println("Created LD "+reg_left+","+reg_right)
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
    c.ops[0x7f] = gen_ld("A","A")
    c.ops[0x78] = gen_ld("A","B")
    c.ops[0x77] = gen_ld("A","C")
    c.ops[0x7A] = gen_ld("A","D")
    c.ops[0x7B] = gen_ld("A","E")
    c.ops[0x7C] = gen_ld("A","H")
    c.ops[0x7D] = gen_ld("A","L")
    c.ops[0x0A] = gen_ld("A","(BC)")
    c.ops[0x1A] = gen_ld("A","(DE)")
    c.ops[0x7E] = gen_ld("A","(HL)")
    c.ops[0xFA] = gen_ld("A","(nn)")
    c.ops[0x3E] = gen_ld("A","n")
    c.ops[0xF2] = gen_ld("A","(C)")
    
    
    c.ops[0x40] = gen_ld("B","B")
    c.ops[0x41] = gen_ld("B","C")
    c.ops[0x42] = gen_ld("B","D")    
    c.ops[0x43] = gen_ld("B","E")
    c.ops[0x44] = gen_ld("B","H")
    c.ops[0x45] = gen_ld("B","L")
    c.ops[0x46] = gen_ld("B","(HL)")
    c.ops[0x47] = gen_ld("B","A")    

    c.ops[0x48] = gen_ld("C","B")
    c.ops[0x49] = gen_ld("C","C")
    c.ops[0x4A] = gen_ld("C","D")
    c.ops[0x4B] = gen_ld("C","E")
    c.ops[0x4C] = gen_ld("C","H")
    c.ops[0x4D] = gen_ld("C","L")
    c.ops[0x4E] = gen_ld("C","(HL)")    
    c.ops[0x4F] =  gen_ld("C","A")    

    c.ops[0x50] = gen_ld("D","B")
    c.ops[0x51] = gen_ld("D","C")
    c.ops[0x52] = gen_ld("D","D")
    c.ops[0x53] = gen_ld("D","E")
    c.ops[0x54] = gen_ld("D","H")
    c.ops[0x55] = gen_ld("D","L")
    c.ops[0x56] = gen_ld("D","(HL)")    
    c.ops[0x57] = gen_ld("D","A")    

    c.ops[0x58] = gen_ld("E","B")
    c.ops[0x59] = gen_ld("E","C")
    c.ops[0x5A] = gen_ld("E","D")
    c.ops[0x5B] = gen_ld("E","E")
    c.ops[0x5C] = gen_ld("E","H")
    c.ops[0x5D] = gen_ld("E","L")
    c.ops[0x5E] = gen_ld("E","(HL)")
    c.ops[0x5F] = gen_ld("E","A")        

    c.ops[0x60] = gen_ld("H","B")
    c.ops[0x61] = gen_ld("H","C")
    c.ops[0x62] = gen_ld("H","D")
    c.ops[0x63] = gen_ld("H","E")
    c.ops[0x64] = gen_ld("H","H")
    c.ops[0x65] = gen_ld("H","L")
    c.ops[0x66] = gen_ld("H","(HL)")
    c.ops[0x67] = gen_ld("H","A")
    
    c.ops[0x68] = gen_ld("L","B")
    c.ops[0x69] = gen_ld("L","C")
    c.ops[0x6A] = gen_ld("L","D")
    c.ops[0x6B] = gen_ld("L","E")
    c.ops[0x6C] = gen_ld("L","H")
    c.ops[0x6D] = gen_ld("L","L")
    c.ops[0x6E] = gen_ld("L","(HL)")
    c.ops[0x6F] = gen_ld("L","A")

    c.ops[0x70] = gen_ld("(HL)","B")
    c.ops[0x71] = gen_ld("(HL)","C")
    c.ops[0x72] = gen_ld("(HL)","D")
    c.ops[0x73] = gen_ld("(HL)","E")
    c.ops[0x74] = gen_ld("(HL)","H")
    c.ops[0x75] = gen_ld("(HL)","L")
    c.ops[0x77] = gen_ld("(HL)","A")
    c.ops[0x36] = gen_ld("(HL)","n")
    
    c.ops[0x02] = gen_ld("(BC)","A")
    c.ops[0x12] = gen_ld("(DE)","A")
    c.ops[0xEA] = gen_ld("(nn)","A")
    c.ops[0xE2] = gen_ld("(C)","A")


    
    return c
}




func main() {
    var c = NewCpu()

    fmt.Println(c.reg8)
    c.reg8["B"] = 3 
    for k := range c.ops {
        c.ops[k](c)
    }
    c.reg8["B"] = 3
    fmt.Println(c.reg8) 
    
    c.exec()
}