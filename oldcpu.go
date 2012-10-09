package main

import "fmt"

type IntVal interface {
  Get(CPU) int
  Set(int)
}

type Action  func(*CPU)

type OpMap map[uint16] Action 




type CPU struct {
    //General registers
    A  uint8
    B  uint8
    C  uint8
    D  uint8
    E  uint8
    F  uint8
    H  uint8
    L  uint8
    
    //stack pointer/program counter
    PC  uint16
    SP  uint16
    
    //flags
    FL uint16
 
    //16 bits of addr space
    mem [0x8000]uint16

}


func (c *CPU) gen_ld(reg_left string, reg_right string, desc string) (Action)  {
    return func(c *CPU)  {
        

    }
}








//copy memory location into register given a pointer to a register
func (c *CPU) ld_8_mem_reg(reg *uint8)  () {
    *reg = uint8(c.mem[c.PC+1])
    tick(8)
    //Increment PC 2 for 1 argument+op
    c.PC+=2
}
//copy a register to a register given two registers
func (c *CPU) ld_8_reg_reg(reg *uint8,src *uint8)  () {

    *reg = uint8(*reg)
    tick(4)
    //Increment PC 1 for op
    c.PC+=1
}


//copy a memory location from HL to register 
func (c *CPU) ld_16_mem_hl_reg(reg *uint8)  () {

    *reg = uint8(c.mem[ (c.H<<8) | c.L])
    tick(8)
    //Increment PC 1 for op
    c.PC+=1
}
//BC
func (c *CPU) ld_16_mem_bc_reg(reg *uint8)  () {

    *reg = uint8(c.mem[ (c.B<<8) | c.C])
    tick(8)
    //Increment PC 1 for op
    c.PC+=1
}
//DE
func (c *CPU) ld_16_mem_de_reg(reg *uint8)  () {

    *reg = uint8(c.mem[ (c.D<<8) | c.E])
    tick(8)
    //Increment PC 1 for op
    c.PC+=1
}
//nn
func (c *CPU) ld_16_mem_n_n_reg(reg *uint8)  () {

    *reg =  uint8(c.mem[(c.PC+2<<8) |c.PC+1 ])
    tick(16)
    //Increment PC 1 for op
    c.PC+=3

//#
}
func (c *CPU) ld_16_mem_n_reg(reg *uint8)  () {

    *reg =  uint8(c.mem[c.PC+1])
    tick(8)
    //Increment PC 1 for op
    c.PC+=2
}


//copy a value from register to (HL) 
func (c *CPU) ld_16_reg_mem(reg *uint8)  () {

    c.mem[ (c.H<<8) | c.L] = uint16(*reg)
    tick(8)
    //Increment PC 1 for op
    c.PC+=1
}

//copy a value from register to (HL) 
func (c *CPU) ld_16_n_mem()  () {

    c.mem[ (c.H<<8) | c.L] =  c.mem[c.PC+1]
    tick(12)
    //Increment PC 1 for op
    c.PC+=1
}



func (c *CPU) exec() ()  {
    
    switch c.PC {
                //SPECIAL Cycles
                //Noop 4 cycles,

                default:
                    fmt.Println("Unsupported opcode:" +string(c.PC))

                case 0x0:
                    tick(4)
    
                //LOAD

                ////Handle all 8 bit loads////
                ///////////mem->reg///////////
                //LD B,n 
                case 0x06:
                    c.ld_8_mem_reg(&c.B)
                //LD C,n 
                case 0x0E:
                    c.ld_8_mem_reg(&c.C)
                //LD D,n 
                case 0x16:
                    c.ld_8_mem_reg(&c.D)
                //LD E,n 
                case  0x1E:
                    c.ld_8_mem_reg(&c.E)
                //LD H,n 
                case 0x26:
                    c.ld_8_mem_reg(&c.H)
                //LD L,n 
                case 0x2E:
                    c.ld_8_mem_reg(&c.L)
        
                ///////////reg->reg///////////
                //LD A,A
                case 0x7F:
                    c.ld_8_reg_reg(&c.A,&c.A)
                //LD A,B
                case 0x78:
                    c.ld_8_reg_reg(&c.A,&c.B)
                //LD A,C
                case 0x79:
                    c.ld_8_reg_reg(&c.A,&c.C)
                //LD A,D
                case 0x7A:
                    c.ld_8_reg_reg(&c.A,&c.D)
                //LD A,E
                case 0x7B:
                    c.ld_8_reg_reg(&c.A,&c.E)
                //LD A,H
                case 0x7C:
                    c.ld_8_reg_reg(&c.A,&c.H)
                //LD A,L
                case 0x7D:
                    c.ld_8_reg_reg(&c.A,&c.L)
                //LD A,(HL)
                case 0x7E:
                    c.ld_16_mem_hl_reg(&c.A)
                
                
                //LD B,A
                case 0x47:
                    c.ld_8_reg_reg(&c.B,&c.A)
                //LD B,B
                case 0x40:
                    c.ld_8_reg_reg(&c.B,&c.B)
                //LD B,C
                case 0x41:
                    c.ld_8_reg_reg(&c.B,&c.C)
                //LD B,D
                case 0x42:
                    c.ld_8_reg_reg(&c.B,&c.D)   
                //LD B,E
                case 0x43:
                    c.ld_8_reg_reg(&c.B,&c.E)
                //LD B,H
                case 0x44:
                    c.ld_8_reg_reg(&c.B,&c.H)    
                //LD B,L
                case 0x45:
                    c.ld_8_reg_reg(&c.B,&c.L)
                //LD B,(HL)
                case 0x46:
                    c.ld_16_mem_hl_reg(&c.B)


                //LD C,A
                case 0x4F:
                    c.ld_8_reg_reg(&c.C,&c.A)
                //LD C,B
                case 0x48:
                    c.ld_8_reg_reg(&c.C,&c.B)
                //LD C,C
                case 0x49:
                    c.ld_8_reg_reg(&c.C,&c.C)
                //LD C,D
                case 0x4A:
                    c.ld_8_reg_reg(&c.C,&c.D)   
                //LD C,E
                case 0x4B:
                    c.ld_8_reg_reg(&c.C,&c.E)
                //LD C,H
                case 0x4C:
                    c.ld_8_reg_reg(&c.C,&c.H)    
                //LD C,L
                case 0x4D:
                    c.ld_8_reg_reg(&c.C,&c.L)
                //LD C,(HL)
                case 0x4E:
                    c.ld_16_mem_hl_reg(&c.C)

                //LD D,A
                case 0x57:
                    c.ld_8_reg_reg(&c.D,&c.A)  
                //LD D,B
                case 0x50:
                    c.ld_8_reg_reg(&c.D,&c.B)
                //LD D,C
                case 0x51:
                    c.ld_8_reg_reg(&c.D,&c.C)
                //LD D,D
                case 0x52:
                    c.ld_8_reg_reg(&c.D,&c.D)   
                //LD D,E
                case 0x53:
                    c.ld_8_reg_reg(&c.D,&c.E)
                //LD D,H
                case 0x54:
                    c.ld_8_reg_reg(&c.D,&c.H)    
                //LD D,L
                case 0x55:
                    c.ld_8_reg_reg(&c.D,&c.L)
                //LD D,(HL)
                case 0x56:
                    c.ld_16_mem_hl_reg(&c.D)

                
                //LD E,A
                case 0x5F:
                    c.ld_8_reg_reg(&c.E,&c.A)
                //LD E,B
                case 0x58:
                    c.ld_8_reg_reg(&c.E,&c.B)
                //LD E,C
                case 0x59:
                    c.ld_8_reg_reg(&c.E,&c.C)
                //LD E,D
                case 0x5A:
                    c.ld_8_reg_reg(&c.E,&c.D)   
                //LD E,E
                case 0x5B:
                    c.ld_8_reg_reg(&c.E,&c.E)
                //LD E,F
                case 0x5C:
                    c.ld_8_reg_reg(&c.E,&c.H)    
                //LD E,L
                case 0x5D:
                    c.ld_8_reg_reg(&c.E,&c.L)
                //LD E,(HL)
                case 0x5E:
                    c.ld_16_mem_hl_reg(&c.E)


                //LD H,A
                case 0x67:
                    c.ld_8_reg_reg(&c.H,&c.A)
                
                //LD H,B
                case 0x60:
                    c.ld_8_reg_reg(&c.H,&c.B)
                //LD H,C
                case 0x61:
                    c.ld_8_reg_reg(&c.H,&c.C)
                //LD H,D
                case 0x62:
                    c.ld_8_reg_reg(&c.H,&c.D)   
                //LD H,E
                case 0x63:
                    c.ld_8_reg_reg(&c.H,&c.E)
                //LD H,F
                case 0x64:
                    c.ld_8_reg_reg(&c.H,&c.H)    
                //LD H,L
                case 0x65:
                    c.ld_8_reg_reg(&c.H,&c.L)
                //LD H,(HL)
                case 0x66:
                    c.ld_16_mem_hl_reg(&c.H)

                //LD L,A
                case 0x6F:
                    c.ld_8_reg_reg(&c.L,&c.A)
                //LD L,B
                case 0x68:
                    c.ld_8_reg_reg(&c.L,&c.B)
                //LD L,C
                case 0x69:
                    c.ld_8_reg_reg(&c.L,&c.C)
                //LD L,D
                case 0x6A:
                    c.ld_8_reg_reg(&c.L,&c.D)   
                //LD L,E
                case 0x6B:
                    c.ld_8_reg_reg(&c.L,&c.E)
                //LD L,H
                case 0x6C:
                    c.ld_8_reg_reg(&c.L,&c.H)    
                //LD L,L
                case 0x6D:
                    c.ld_8_reg_reg(&c.L,&c.L)
                //LD L,(HL)
                case 0x6E:
                    c.ld_16_mem_hl_reg(&c.D)
        
                //LD (HL),B
                case 0x70:
                    c.ld_16_reg_mem(&c.B)
                //LD (HL),C
                case 0x71:
                    c.ld_16_reg_mem(&c.C)
                //LD (HL),D
                case 0x72:
                    c.ld_16_reg_mem(&c.D)   
                //LD (HL),E
                case 073:
                    c.ld_16_reg_mem(&c.E)
                //LD (HL),H
                case 0x74:
                    c.ld_16_reg_mem(&c.H)    
                //LD (HL),L
                case 0x75:
                    c.ld_16_reg_mem(&c.L)
                //LD (HL),n
                case 0x36:
                    c.ld_16_n_mem()

                //A ops
                case 0x0A:
                   c.ld_16_mem_bc_reg(&c.A)
                case 0x1A:
                   c.ld_16_mem_de_reg(&c.A)
                case 0xFA:
                   c.ld_16_mem_n_n_reg(&c.A)
                case 0x3E:
                   c.ld_16_mem_n_reg(&c.A)

            
           }   
}

func tick(count int) {
    //Add code for delying execution
    for i:=0; i<count; i++ {
        fmt.Println("Tick")
    }
}



func main() {


    var x CPU

    x.PC=0x36
    x.exec()
}