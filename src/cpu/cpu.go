package cpu

import (
	"clock"
	"component"
	"dmac"
	"dram"
	"fmt"
	"gp"
	"gpu"
	"ic"
	"mmu"
	"os"
	"serial"
	"sound"
	"timer"
	//"runtime/pprof"
	//"time"
	//"bufio"
)

const (
	PC = iota
	SP
)
const (
	A = iota
	B
	C
	D
	E
	F
	H
	L
	FL_Z
	FL_C
	FL_H
	FL_N
	EI
)

const (
	reg8 = iota
	reg16
	reg16_combo
	regc
	reghld
	reghli
	regldh
	memreg16
	memreg
	memnn
	memnn8
	memn
	nn
	n
	invalid
)
const (
	DIV_MMIO  = 0xff04
	KEY1_MMIO = 0xff4d
)

type Action func(*CPU)
type SetVal func(*CPU, uint16)
type GetVal func(*CPU) uint16
type OpMap [0xffff]Action
type RegMap8 [13]uint8  //12 registers one uint for each of the
type RegMap16 [2]uint16 //2 registers
type OpCall map[uint16]uint32

type CPU struct {
	ops        OpMap
	reg8       RegMap8
	reg16      RegMap16
	mmu        *mmu.MMU
	gpu        *gpu.GPU
	gp         *gp.GP
	serial     serial.Serial
	timer      *timer.Timer
	ic         *ic.IC
	sound      *sound.Sound
	dram       *dram.DRAM
	dmac       *dmac.DMAC
	is_halted  bool
	DIV        uint8
	KEY1       uint8
	last_instr uint16
	push_pc    Action
	sswitch    bool
	clk_mul    uint16
	reg_list   component.RegList
	clock      *clock.Clock
	bt_debug_buffer   [1000]uint16
	bt_count int
	ei_wait_instr bool
}

func (c *CPU) Ready_sswitch() {
	c.sswitch = true
}

func (c *CPU) Read_mmio(addr uint16) uint8 {

	var val uint8

	switch addr {
	case DIV_MMIO: //DIV counter register ... needs to go in timer
		val = c.DIV
	case KEY1_MMIO: //KEY1 clock register
		val = c.KEY1
	default:
		panic("unhandled cpu mmio read")
	}
	return val
}

func (c *CPU) Write_mmio(addr uint16, val uint8) {

	switch addr {
	case KEY1_MMIO:

		//get ready to switch speed
		if val&0x1 == 1 {
			c.Ready_sswitch()
		}
	default:
		fmt.Printf("unhandled cpu mmio write addr:%x val:%x\n",addr,val)
	}
}

func (c *CPU) set_sswitch() {
	if c.sswitch == true {
		c.sswitch = false
		if c.clk_mul == 2 {
			c.clk_mul = 1
			fmt.Println("CLK to 4mhz")
			c.KEY1 = 0
		} else {
			c.clk_mul = 2
			fmt.Println("CLK to 8mhz")
			c.KEY1 = 0xFE

		}
	}

}

func (c *CPU) Reset() {
	c.reg16[PC] = 0x100
	c.reg16[SP] = 0xfffe
	c.reg8[A] = 0x11
	c.reg8[B] = 0x00
	c.reg8[C] = 0x13
	c.reg8[D] = 0x00
	c.reg8[E] = 0xd8
	c.reg8[F] = 0x90
	c.reg8[H] = 0x01
	c.reg8[L] = 0x4d
	c.reg8[FL_Z] = 0x1
	c.reg8[FL_C] = 0x1
	c.reg8[FL_H] = 0x0
	c.reg8[FL_N] = 0x0
	c.reg8[EI] = 0x0
	c.gpu.STAT = 0x85
	c.gpu.LCDC = 0x91
}

func (c *CPU) load_bios() {

	c.mmu.Create_new_cart(os.Args[len(os.Args)-1])

	c.reg16[PC] = 0x100
	c.reg16[SP] = 0xfffe
	c.reg8[A] = 0x11
	c.reg8[B] = 0x00
	c.reg8[C] = 0x13
	c.reg8[D] = 0x00
	c.reg8[E] = 0xd8
	c.reg8[F] = 0x90
	c.reg8[H] = 0x01
	c.reg8[L] = 0x4d
	c.reg8[FL_Z] = 0x1
	c.reg8[FL_C] = 0x1
	c.reg8[FL_H] = 0x0
	c.reg8[FL_N] = 0x0
	c.reg8[EI] = 0x0
	c.gpu.STAT = 0x85
	c.gpu.LCDC = 0x91

}

func get_reg_id(reg string) int {
	var val int = -1
	switch reg {
	case "PC":
		val = PC
	case "SP":
		val = SP
	case "A":
		val = A
	case "B":
		val = B
	case "C":
		val = C
	case "D":
		val = D
	case "E":
		val = E
	case "F":
		val = F
	case "H":
		val = H
	case "L":
		val = L
	case "FL_Z":
		val = FL_Z
	case "FL_C":
		val = FL_C
	case "FL_H":
		val = FL_H
	case "FL_N":
		val = FL_N

	}
	return val
}
func (c *CPU) handleInterrupts() {

	if c.is_halted && c.ic.IF&c.ic.IE != 0 {
		c.is_halted = false
		//fmt.Println("CORE UNHALTED",c.ic.IF,c.ic.IE)

	}

	if c.reg8[EI] == 1 {
		if c.ei_wait_instr == true {
			c.ei_wait_instr = false
			return
		}
		vector := c.ic.Handle()
		//Handle will Dissassert interrupt
		if vector > 0 {
			c.reg8[EI] = 0
			c.push_pc(c) //push pc on stack
			c.is_halted = false
			c.last_instr += 20
			//fmt.Println("Handled at at LY",c.gpu.LY,c.gpu.LYC,vector)

			c.reg16[PC] = vector

		}

	}
}
func (c *CPU) Dump() {
	fmt.Printf("Cnt:%d PC:%04x SP:%04x A:%02x B:%02x C:%02x D:%02x E:%02x H:%02x L:%02x FL_Z:%01x FL_C:%01x FL_H:%01x LY:%02x  STAT:%x \n",c.clock.Cycles, c.reg16[PC], c.reg16[SP], c.reg8[A], c.reg8[B], c.reg8[C], c.reg8[D], c.reg8[E], c.reg8[H], c.reg8[L], c.reg8[FL_Z], c.reg8[FL_C], c.reg8[FL_H], c.gpu.LY, c.gpu.STAT) //,c.reg8[FL_N]);
}
func (c *CPU) Exec() {

	c.load_bios()
	var op uint16

	//	fo, err := os.Create("output.txt")
	//	if err != nil { panic(err) }
	////		  defer fo.Close()
	//		pprof.StartCPUProfile(fo)
	//	last_update := time.Now()
	count := uint(0)
	for {
		dma_clocks := c.dmac.Update()

		c.last_instr +=dma_clocks
        cycles := c.last_instr>>(c.clk_mul-1)
		in_oam := false
		if dma_clocks > 0 {
			in_oam = true
		}
		//c.last_instr = 4
		if c.ic.IF > 0 {
			c.handleInterrupts()
		}
		c.sound.Update(cycles)
		//gameboy color executes oam dma in 76 cycles not 80
		//run op
		//c.last_instr/c.clk_mul
		c.sound.Update(cycles)
		//for i:=uint16(0); i<cycles; i++ {
		//	c.gpu.Update(c.last_instr/c.clk_mul,in_oam)
		c.gpu.Update(cycles,in_oam)
			
		//}

		c.mmu.Update(c.reg16[PC],c.gpu.LY,c.clock.Cycles)
		if !c.is_halted {

			op = uint16(c.mmu.Read_w(c.reg16[PC]))
			//c.Dump()
			//fmt.Println(c.gpu.LY)
			if op&0x00ff != 0xcb {
				op &= 0xff

			} else {
				op = 0xcb00 | ((op & 0xff00) >> 8)
			}
//			c.Dump()
			action := c.ops[op]
			if action == nil {
				fmt.Printf("Undefined opcode %x \n",op)
				fmt.Println("MMU  State")
				c.mmu.Dump()
				fmt.Println("CPU State")
				c.Dump()
				fmt.Println(c.bt_debug_buffer)
				fmt.Println(c.bt_count)
				panic("Intruction would fail quitting...")
			
			}
			action(c)
			//c.last_instr /=2
			count++

		}

		//fmt.Println(count)
		//Update gamepad/buttons
		if count >= 2000 {
			raise_int := c.gp.Update()
			count = 0
			if raise_int == 0xff {
				fmt.Println("GLOBAL RESET")
				c.Reset()
			} else if raise_int > 0 {
				//c.ic.Assert(raise_int)
			}
		}
		c.serial.Update(cycles)

		//for i:=0; i< int(c.last_instr); i++ {
		//	}
		raise_int := c.timer.Update(uint64(c.last_instr))
		if raise_int > 0 {
			c.ic.Assert(raise_int)
		}
		c.DIV++
		//	 if time.Since(last_update) > 20 *time.Second {

		//	pprof.StopCPUProfile()
		//	 fmt.Println("STOPPED")
		//	}
		c.clock.Cycles += uint64(c.last_instr >>2)

	}
}

func get_ld_type(arg string) int {
	//turn type into token

	var arg_type int

	switch {
	case arg == "(HLD)":
		arg_type = reghld
	case (arg == "(PC)" || arg == "(SP)"):
		arg_type = memreg16
	case arg == "(HLI)":
		arg_type = reghli
	case arg == "(LDH)":
		arg_type = regldh
	case arg == "(nn)":
		arg_type = memnn
	case arg == "(nn)8":
		arg_type = memnn8
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

func (c *CPU) do_instr(desc string, ticks uint16, args uint16) {

	//c.tick(ticks)
	//time.Sleep(time.Microsecond)
	//if !c.mmu.inbios   {
	//fmt.Println(desc,ticks)
	//fmt.Printf("PC:%04",c.reg16[PC])
	//  c.Print_dump()
	//	}
	c.last_instr = ticks
	c.reg16[PC] += args
	c.bt_debug_buffer[c.bt_count] = c.reg16[PC]
	c.bt_count +=1
	if c.bt_count == 1000 {
		c.bt_count = 0
	}

}

func gen_set_val(a_type int, reg string) SetVal {
	lambda := func(c *CPU, val uint16) {}
	reg_id := get_reg_id(reg)
	switch a_type {
	case reg8:
		lambda = func(c *CPU, val uint16) { c.reg8[reg_id] = uint8(val) }
	case regc:
		lambda = func(c *CPU, val uint16) {
			c.mmu.Write_b(uint16(0xff00|uint16(c.reg8[C])), uint8(val))
		}
	case reg16_combo:
		reg_0 := get_reg_id(string(reg[0]))
		reg_1 := get_reg_id(string(reg[1]))
		lambda = func(c *CPU, val uint16) {
			c.reg8[reg_0] = uint8((val & 0xff00) >> 8)
			c.reg8[reg_1] = uint8(val & 0x00ff)
		}
	case reg16:
		lambda = func(c *CPU, val uint16) {
			c.reg16[reg_id] = val
		}

	case memreg:
		reg_0 := get_reg_id(string(reg[1]))
		reg_1 := get_reg_id(string(reg[2]))
		lambda = func(c *CPU, val uint16) {
			reg_high := c.reg8[reg_0]
			reg_low := c.reg8[reg_1]
			addr := (uint16(reg_high) << 8) | uint16(reg_low)
			c.mmu.Write_b(addr, uint8(val))
		}

	case memn:
		lambda = func(c *CPU, val uint16) {
			addr := 0xff00 | uint16(c.mmu.Read_b(c.reg16[PC]+1))
			c.mmu.Write_b(addr, uint8(val))

		}
	case memnn:
		lambda = func(c *CPU, val uint16) {
			addr := uint16(c.mmu.Read_b(c.reg16[PC]+2))<<8 | uint16(c.mmu.Read_b(c.reg16[PC]+1))

			c.mmu.Write_w(addr, val)
		}
	case memnn8:
		lambda = func(c *CPU, val uint16) {
			addr := uint16(c.mmu.Read_b(c.reg16[PC]+2))<<8 | uint16(c.mmu.Read_b(c.reg16[PC]+1))

			c.mmu.Write_b(addr, uint8(val))
		}

	case reghli:
		f := gen_alu("INC", "HL", "", 0, 0)
		reg_0 := get_reg_id(string(reg[1]))
		reg_1 := get_reg_id(string(reg[2]))
		lambda = func(c *CPU, val uint16) {
			reg_high := c.reg8[reg_0]
			reg_low := c.reg8[reg_1]

			addr := (uint16(reg_high))<<8 | uint16(reg_low)
			c.reg8[reg_0] = uint8(addr & 0xff00 >> 8)
			c.reg8[reg_1] = uint8(addr & 0x00ff)

			c.mmu.Write_b(addr, uint8(val))
			f(c)
		}
	case reghld:
		f := gen_alu("DEC", "HL", "", 0, 0)
		reg_0 := get_reg_id(string(reg[1]))
		reg_1 := get_reg_id(string(reg[2]))
		lambda = func(c *CPU, val uint16) {
			reg_high := c.reg8[reg_0]
			reg_low := c.reg8[reg_1]
			addr := (uint16(reg_high))<<8 | uint16(reg_low)

			c.reg8[reg_0] = uint8(addr & 0xff00 >> 8)
			c.reg8[reg_1] = uint8(addr & 0x00ff)
			f(c)
			c.mmu.Write_b(addr, uint8(val))
		}
	default:
		lambda = func(c *CPU, val uint16) {
			fmt.Println("UNHANDLED Set  ERROR", reg)
		}

	}

	return lambda
}

func gen_get_val(a_type int, reg string) GetVal {

	lambda := func(c *CPU) uint16 { fmt.Println("UNDEFINED !!!!!!!!!!!!!"); return 0 }
	reg_id := get_reg_id(reg)

	switch a_type {
	case reg8:
		lambda = func(c *CPU) uint16 { return uint16(c.reg8[reg_id]) }
	case memn:
		lambda = func(c *CPU) uint16 {
			addr := uint16(0xff00 | uint16(c.mmu.Read_b(c.reg16[PC]+1)))
			return c.mmu.Read_w(addr)
		}

	case regc:
		lambda = func(c *CPU) uint16 {
			addr := uint16(0xff00 | uint16(c.reg8[C]))
			return uint16(c.mmu.Read_b(addr))
		}

	case memreg:
		reg_0 := get_reg_id(string(reg[1]))
		reg_1 := get_reg_id(string(reg[2]))
		lambda = func(c *CPU) uint16 {
			reg_high := c.reg8[reg_0]
			reg_low := c.reg8[reg_1]
			addr := (uint16(reg_high) << 8) | uint16(reg_low)

			return uint16(c.mmu.Read_b(addr))
		}
	case reg16:
		lambda = func(c *CPU) uint16 {
			return c.reg16[reg_id]
		}

	case memreg16:
		lambda = func(c *CPU) uint16 {
			//	fmt.Println("CALLEDDD!!!")
			return c.mmu.Read_w(c.reg16[reg_id])
		}

	case reg16_combo:
		reg_0 := get_reg_id(string(reg[0]))
		reg_1 := get_reg_id(string(reg[1]))
		lambda = func(c *CPU) uint16 {
			var reg_high uint8 = c.reg8[reg_0]
			var reg_low uint8 = c.reg8[reg_1]
			return uint16(reg_high)<<8 | uint16(reg_low)

		}

	case memnn:
		lambda = func(c *CPU) uint16 {
			addr := uint16(c.mmu.Read_b(c.reg16[PC]+2))<<8 | uint16(c.mmu.Read_b(c.reg16[PC]+1))
			//		fmt.Printf("%04x,%04x\n",addr,c.mmu.Read_w(addr))
			return c.mmu.Read_w(addr)
		}

	case n:
		lambda = func(c *CPU) uint16 {
			return uint16(c.mmu.Read_b(c.reg16[PC] + 1))
		}
	case nn:
		lambda = func(c *CPU) uint16 {
			return uint16(c.mmu.Read_b(c.reg16[PC]+2))<<8 | uint16(c.mmu.Read_b(c.reg16[PC]+1))
		}

	case reghld:

		f := gen_alu("DEC", "HL", "", 0, 0)
		reg_0 := get_reg_id(string(reg[1]))
		reg_1 := get_reg_id(string(reg[2]))
		lambda = func(c *CPU) uint16 {
			//parse reg_right
			reg_high := c.reg8[reg_0]
			reg_low := c.reg8[reg_1]

			addr := (uint16(reg_high) << 8) | uint16(reg_low)

			c.reg8[reg_0] = uint8(addr & 0xff00 >> 8)
			c.reg8[reg_1] = uint8(addr & 0x00ff)
			f(c)
			return uint16(c.mmu.Read_w(addr))

		}

	case reghli:
		//parse reg_right
		f := gen_alu("INC", "HL", "", 0, 0)
		reg_0 := get_reg_id(string(reg[1]))
		reg_1 := get_reg_id(string(reg[2]))
		lambda = func(c *CPU) uint16 {
			reg_high := c.reg8[reg_0]
			reg_low := c.reg8[reg_1]
			addr := (uint16(reg_high) << 8) | uint16(reg_low)
			c.reg8[reg_0] = uint8(addr & 0xff00 >> 8)
			c.reg8[reg_1] = uint8(addr & 0x00ff)
			f(c)
			return uint16(c.mmu.Read_b(addr))
		}

	}
	return lambda
}

func gen_alu(op_type string, reg_left string, reg_right string, ticks uint16, args uint16) Action {
	type_right := get_ld_type(reg_right)
	type_left := get_ld_type(reg_left)

	desc := op_type + " " + reg_left + "," + reg_right

	lambda := func(c *CPU) { fmt.Println("Undefined ALU op", op_type) }

	f_right_get_val := gen_get_val(type_right, reg_right)
	f_left_get_val := gen_get_val(type_left, reg_left)
	f_set_val := gen_set_val(type_left, reg_left)

	switch op_type {
	case "DAA":
		lambda = func(c *CPU) {
			val := uint16(f_right_get_val(c))
			if c.reg8[FL_N] != 1 {
				if c.reg8[FL_H] == 1 || ((val & 0x0f) > 0x09) {
					val += 0x06
				}
				if c.reg8[FL_C] == 1 || val > 0x9f {
					val += 0x60
				}

			} else {
				if c.reg8[FL_H] == 1 {
					val = (val - 6) & 0xff
				}
				if c.reg8[FL_C] == 1 {
					val -= 0x60
				}

			}
			if val&0x100 > 0 {
				c.reg8[FL_C] = 1

			}
			c.reg8[FL_H] = 0
			val &= 0xff
			if val == 0x0 {
				c.reg8[FL_Z] = 1
			} else {
				c.reg8[FL_Z] = 0
			}
			f_set_val(c, val)
			c.do_instr(desc, (ticks), (args))

		}

	case "ADD":
		if reg_left == "SP" {

			lambda = func(c *CPU) {
				n := f_right_get_val(c)
				prev := f_left_get_val(c)
				//fmt.Println("before",n)

				if n > 127 {

					val := ^uint8((n & 0x00ff))
					val += 1

					//	fmt.Println(val)
					f_set_val(c, prev-uint16(val))

				} else {
					f_set_val(c, prev+n)
				}
				//	fmt.Println(n,f_right_get_val(c))
				if 0xff-(prev&0x00ff) < n {
					c.reg8[FL_C] = 1
				} else {
					c.reg8[FL_C] = 0

				}
				if 0x000f-(prev&0x000f) < (n & 0x000f) {
					c.reg8[FL_H] = 1
				} else {
					c.reg8[FL_H] = 0

				}

				c.reg8[FL_Z] = 0
				c.reg8[FL_N] = 0

				c.do_instr(desc, (ticks), (args))

			}

		} else {
			lambda = func(c *CPU) {
				prev := f_left_get_val(c)
				right := f_right_get_val(c)

				f_set_val(c, prev+right)

				if uint32((prev&0xf)+(right&0xf)) > 15 {
					c.reg8[FL_H] = 1
				} else {
					c.reg8[FL_H] = 0
				}
				if uint32(prev+right) > 0xff {
					c.reg8[FL_C] = 1
				} else {
					c.reg8[FL_C] = 0
				}
				if uint8(prev)+uint8(right) == 0x0 {
					c.reg8[FL_Z] = 1
				} else {
					c.reg8[FL_Z] = 0
				}
				c.reg8[FL_N] = 0

				c.do_instr(desc, ticks, args)
			}
		}
	case "ADD16":
		lambda = func(c *CPU) {
			prev := f_left_get_val(c)
			right := f_right_get_val(c)
			if reg_left != "SP" {
				//we don't set flags when regleft is sp
				if 0x0fff-(prev&0x0fff) < (right & 0x0fff) {
					c.reg8[FL_H] = 1
				} else {
					c.reg8[FL_H] = 0
				}
				if (0xffff - prev) < right {
					c.reg8[FL_C] = 1
				} else {
					c.reg8[FL_C] = 0
				}

				c.reg8[FL_N] = 0
			}

			f_set_val(c, prev+right)
			//			fmt.Printf("Add16:%x+%x=%x\n",prev,right, f_left_get_val(c))

			c.do_instr(desc, ticks, args)
		}

	case "SUB":
		lambda = func(c *CPU) {
			prev := f_left_get_val(c)
			right := f_right_get_val(c)
			f_set_val(c, f_left_get_val(c)-f_right_get_val(c))

			if (prev & 0xf) < (right & 0xf) {
				c.reg8[FL_H] = 1
			} else {
				c.reg8[FL_H] = 0
			}
			if prev < right {
				c.reg8[FL_C] = 1

			} else {
				c.reg8[FL_C] = 0
			}
			if prev == right {
				c.reg8[FL_Z] = 1
			} else {
				c.reg8[FL_Z] = 0
			}
			c.reg8[FL_N] = 1

			c.do_instr(desc, ticks, args)
		}
	case "CP":
		lambda = func(c *CPU) {
			right := uint8(f_right_get_val(c))
			left := uint8(f_left_get_val(c))
			if (left & 0xf) < (right & 0xf) {
				c.reg8[FL_H] = 1
			} else {
				c.reg8[FL_H] = 0
			}
			if left < right {
				c.reg8[FL_C] = 1
			} else {
				c.reg8[FL_C] = 0
			}
			if left-right == 0 {
				c.reg8[FL_Z] = 1
			} else {
				c.reg8[FL_Z] = 0
			}
			c.reg8[FL_N] = 1
			c.do_instr(desc, ticks, args)
		}

	case "SBC":
		lambda = func(c *CPU) {
			a := f_left_get_val(c)
			b := f_right_get_val(c)
			carry_set := 0
			half_set := 0
			fl := uint16(c.reg8[FL_C])
			temp := a - fl

			if a < fl {
				carry_set = 1
			}

			if (a & 0x0F) < (fl & 0x0F) {
				half_set = 1
			}

			temp2 := temp - b

			// will there be a borrow? if so, no carry
			if temp < b {
				carry_set = 1
			}

			// will the lower nibble borrow? if so, no carry
			if (temp & 0x0F) < (b & 0x0F) {
				half_set = 1
			}

			f_set_val(c, uint16(temp2))

			if (half_set) == 1 {
				c.reg8[FL_H] = 1
			} else {
				c.reg8[FL_H] = 0
			}
			if carry_set == 1 {
				c.reg8[FL_C] = 1
			} else {
				c.reg8[FL_C] = 0
			}
			if f_left_get_val(c) == 0 {
				c.reg8[FL_Z] = 1
			} else {
				c.reg8[FL_Z] = 0
			}

			c.reg8[FL_N] = 1

			c.do_instr(desc, ticks, args)
		}
	case "ADC":
		lambda = func(c *CPU) {
			a := f_left_get_val(c)
			b := f_right_get_val(c)
			carry_set := 0
			half_set := 0
			fl := uint16(c.reg8[FL_C])
			temp := a + fl

			if 0xff-a < fl {
				carry_set = 1
			}

			if 0x0f-(a&0x0F) < (fl & 0x0F) {
				half_set = 1
			}

			temp2 := temp + b

			// will there be a borrow? if so, no carry
			if 0xff-temp < b {
				carry_set = 1
			}

			// will the lower nibble borrow? if so, no carry
			if 0x0f-(temp&0x0F) < (b & 0x0F) {
				half_set = 1
			}

			f_set_val(c, uint16(temp2))

			if (half_set) == 1 {
				c.reg8[FL_H] = 1
			} else {
				c.reg8[FL_H] = 0
			}
			if carry_set == 1 {
				c.reg8[FL_C] = 1
			} else {
				c.reg8[FL_C] = 0
			}
			if f_left_get_val(c) == 0 {
				c.reg8[FL_Z] = 1
			} else {
				c.reg8[FL_Z] = 0
			}

			c.reg8[FL_N] = 0

			c.do_instr(desc, ticks, args)

		}

	case "AND":
		lambda = func(c *CPU) {
			f_set_val(c, f_left_get_val(c)&f_right_get_val(c))
			if f_left_get_val(c) == 0 {
				c.reg8[FL_Z] = 1
			} else {
				c.reg8[FL_Z] = 0
			}
			c.reg8[FL_C] = 0
			c.reg8[FL_H] = 1
			c.reg8[FL_N] = 0

			c.do_instr(desc, ticks, args)
		}

	case "OR":
		lambda = func(c *CPU) {
			f_set_val(c, f_left_get_val(c)|f_right_get_val(c))
			if f_left_get_val(c) == 0 {
				c.reg8[FL_Z] = 1
			} else {
				c.reg8[FL_Z] = 0
			}
			c.reg8[FL_C] = 0
			c.reg8[FL_H] = 0
			c.reg8[FL_N] = 0
			c.do_instr(desc, ticks, args)
		}
	case "XOR":
		lambda = func(c *CPU) {
			f_set_val(c, f_left_get_val(c)^f_right_get_val(c))
			if f_left_get_val(c) == 0 {
				c.reg8[FL_Z] = 1
			} else {
				c.reg8[FL_Z] = 0
			}
			c.reg8[FL_C] = 0
			c.reg8[FL_H] = 0
			c.reg8[FL_N] = 0
			c.do_instr(desc, ticks, args)
		}

	case "INC":
		lambda = func(c *CPU) {

			f_set_val(c, f_left_get_val(c)+1)

			if len(reg_left) != 2 {
				val := f_left_get_val(c)
				if (val & 0x0f) == 0 {
					c.reg8[FL_H] = 1
				} else {
					c.reg8[FL_H] = 0
				}
				c.reg8[FL_N] = 0
				if val == 0 {
					c.reg8[FL_Z] = 1
				} else {
					c.reg8[FL_Z] = 0
				}
			}

			c.do_instr(desc, ticks, args)
		}
	case "DEC":
		lambda = func(c *CPU) {

			f_set_val(c, f_left_get_val(c)-1)

			if len(reg_left) != 2 {
				val := f_left_get_val(c)
				if (val & 0x0f) == 0x0f {
					c.reg8[FL_H] = 1
				} else {
					c.reg8[FL_H] = 0
				}
				c.reg8[FL_N] = 1
				if val == 0 {
					c.reg8[FL_Z] = 1
				} else {
					c.reg8[FL_Z] = 0
				}
			}

			c.do_instr(desc, ticks, args)
		}

	}

	return lambda

}

func gen_push_pop(left string, reg_right string) Action {
	type_right := get_ld_type(reg_right)

	desc := left + "," + reg_right

	lambda := func(c *CPU) { fmt.Println("Undefined PUSH op" + left) }

	if left == "PUSH" {
		f_get_val := gen_get_val(type_right, reg_right)
		lambda = func(c *CPU) {
			//write word to mem
			c.reg16[SP] -= 2
			c.mmu.Write_w(c.reg16[SP], f_get_val(c))
			c.do_instr(desc, 16, 1)
		}
	} else if left == "PUSHAF" {
		f_get_val := gen_get_val(type_right, reg_right)
		lambda = func(c *CPU) {
			//write word to mem
			c.reg8[F] = 0

			if c.reg8[FL_C] == 1 {
				c.reg8[F] |= 0x10
			}
			if c.reg8[FL_H] == 1 {
				c.reg8[F] |= 0x20
			}
			if c.reg8[FL_N] == 1 {
				c.reg8[F] |= 0x40
			}
			if c.reg8[FL_Z] == 1 {
				c.reg8[F] |= 0x80
			}

			c.reg16[SP] -= 2

			c.mmu.Write_w(c.reg16[SP], f_get_val(c))

			c.do_instr(desc, 16, 1)

		}
	} else if left == "POP" {

		f_set_val := gen_set_val(type_right, reg_right)
		lambda = func(c *CPU) {
			val := c.mmu.Read_w(c.reg16[SP])
			c.reg16[SP] += 2
			f_set_val(c, val)

			c.do_instr(desc, 12, 1)

		}
	} else if left == "POPAF" {

		f_set_val := gen_set_val(type_right, reg_right)
		lambda = func(c *CPU) {

			val := c.mmu.Read_w(c.reg16[SP])
			//fmt.Printf("read 0x%04x to 0x%04x",val,c.reg16[SP])
			//fmt.Println(val)
			new_val := uint8(val & 0x00ff)
			c.reg8[FL_C] = 0
			c.reg8[FL_H] = 0
			c.reg8[FL_N] = 0
			c.reg8[FL_Z] = 0

			if new_val&0x10 == 0x10 {
				c.reg8[FL_C] = 1
				//	fmt.Printf("Set C\n")

			}

			if new_val&0x20 == 0x20 {
				c.reg8[FL_H] = 1
				//	fmt.Printf("Set H\n")

			}
			if new_val&0x40 == 0x40 {
				c.reg8[FL_N] = 1
				//	fmt.Printf("Set N\n")
			}
			if new_val&0x80 == 0x80 {
				c.reg8[FL_Z] = 1
				//	fmt.Printf("Set Z\n")
			}
			//	fmt.Printf("Pop->Write->new_val 0x%04x\n",new_val)

			//	fmt.Printf("Pop->Write 0x%04x\n",val)

			c.reg16[SP] += 2
			f_set_val(c, val)

			c.do_instr(desc, 12, 1)

		}
	}
	return lambda

}

func gen_jmp(left string, reg_right string, skip_flags uint8, signed uint8, mask uint8, check uint8, ticks uint16, ticks_taken uint16, args uint16) Action {

	type_right := get_ld_type(reg_right)

	desc := left + "," + reg_right
	f_get_val := gen_get_val(type_right, reg_right)

	//create func to set PC
	if reg_right == "(HL)" {
		reg_0 := get_reg_id(string(reg_right[1]))
		reg_1 := get_reg_id(string(reg_right[2]))

		f_get_val = func(c *CPU) uint16 {
			fmt.Println("FFFFIIIIXXX")
			reg_high := c.reg8[reg_0]
			reg_low := c.reg8[reg_1]
			addr := (uint16(reg_high) << 8) | uint16(reg_low)
			//only case that reads word from (HL)
			return uint16(c.mmu.Read_w(addr))
		}
	}

	lambda := func(c *CPU) {
		n := f_get_val(c)
		//	fmt.Printf("JMP ADDR:0x%X,mask:0x%X,check:0x%X\n",n,mask,skip_flags)
		reg := FL_Z
		switch mask {
		case 0x10:
			reg = FL_C

		case 0x80:
			reg = FL_Z

		}

		if skip_flags == 1 || c.reg8[reg] == check {

			//relative jump
			if signed == 1 {
				if n > 127 {
					n = (^(n + 1)) & 0x00ff
					//	fmt.Printf("%02x\n",n)
					c.reg16[PC] -= n

				} else {
					//	fmt.Printf("%02x\n",n)

					c.reg16[PC] += n + args

				}
				//	fmt.Println("s Jump")

			} else {
				c.reg16[PC] = n
			}
			//	fmt.Println("o Jump")

			c.do_instr(desc, ticks_taken, 0) //skip args if we are acutally doing jump
		} else {
			//	fmt.Println("Mo Jump")
			c.do_instr(desc, ticks, args)

		}
	}

	return lambda
}
func gen_call(left string, reg_right string, skip_flags uint8, signed uint8, mask uint8, check uint8, ticks uint16, ticks_taken uint16, args uint16) Action {

	p_func := gen_push_pop("PUSH", "PC")
	jmp_func := gen_jmp("CJMP", reg_right, skip_flags, 0, mask, check, 8, 16, 3) //signed

	lambda := func(c *CPU) {
		prev := c.reg16[PC]

		jmp_func(c)
		if c.reg16[PC]-prev != 3 { //We actually did the jump 3 is for the arg count for this instruction
			jmp_val := c.reg16[PC]
			//Only push PC on stack if we take the jump
			c.reg16[PC] = 3 + prev
			p_func(c)
			c.reg16[PC] = jmp_val
			c.do_instr("CALL", ticks_taken, 0)
		} else {
			c.do_instr("CALL", ticks, 0)

		}
	}
	//hack to fix ticks

	return lambda
}

func gen_ret(left string, skip_flag uint8, mask uint8, check uint8, ticks uint16, ticks_taken uint16) Action {

	f_get_val := gen_get_val(memreg16, "SP")
	f_set_val := gen_set_val(reg16, "PC")
	reg := FL_Z
	switch mask {
	case 0x10:
		reg = FL_C

	case 0x80:
		reg = FL_Z
	}

	lambda := func(c *CPU) {
		if skip_flag == 1 || c.reg8[reg] == check {
			//fmt.Println("RETURNING","z",c.reg8[reg])
			val := f_get_val(c)

			f_set_val(c, val)
			c.reg16[SP] += 2
			if left == "RETI" {
				c.reg8[EI] = 1
				c.Dump()
			}
			c.do_instr(left, ticks_taken, 0)
		} else {
			c.do_instr(left, ticks, 1)
		}

	}
	return lambda
}

func gen_rotate_shift(left string, reg_right string, ticks uint16, args uint16) Action {
	type_right := get_ld_type(reg_right)

	desc := left + "," + reg_right

	lambda := func(c *CPU) { fmt.Println("Undefined SHIFT op" + left + reg_right) }
	f_left_get_val := gen_get_val(type_right, reg_right)
	f_set_val := gen_set_val(type_right, reg_right)
	switch left {

	case "RLC":
		lambda = func(c *CPU) {
			prev := uint8(f_left_get_val(c))
			c.reg8[FL_C] = (prev & 0x80) >> 7
			//fmt.Println("Before",prev,(prev << 1))

			f_set_val(c, uint16((prev<<1)+c.reg8[FL_C]))

			if f_left_get_val(c) == 0 {
				c.reg8[FL_Z] = 1
			} else {
				c.reg8[FL_Z] = 0
			}
			//fmt.Println("After",f_left_get_val(c))
			c.reg8[FL_H] = 0
			c.reg8[FL_N] = 0

			c.do_instr(desc, ticks, args)
		}
	case "RL":
		lambda = func(c *CPU) {
			prev := uint8(f_left_get_val(c))
			temp := c.reg8[FL_C]
			c.reg8[FL_C] = (prev & 0x80) >> 7
			f_set_val(c, uint16((prev<<1)+temp))

			if f_left_get_val(c) == 0 {
				c.reg8[FL_Z] = 1
			} else {
				c.reg8[FL_Z] = 0
			}
			c.reg8[FL_H] = 0
			c.reg8[FL_N] = 0

			c.do_instr(desc, ticks, args)
		}
	case "RRC":
		lambda = func(c *CPU) {
			prev := uint8(f_left_get_val(c))
			c.reg8[FL_C] = (prev & 0x01)
			f_set_val(c, uint16((prev>>1)+(c.reg8[FL_C]<<7)))

			if f_left_get_val(c) == 0 {
				c.reg8[FL_Z] = 1
			} else {
				c.reg8[FL_Z] = 0
			}
			c.reg8[FL_H] = 0
			c.reg8[FL_N] = 0

			c.do_instr(desc, ticks, args)
		}

	case "SRA":
		lambda = func(c *CPU) {
			prev := uint8(f_left_get_val(c))
			c.reg8[FL_C] = (prev & 0x01)
			prev >>= 1
			f_set_val(c, uint16(prev|((prev&0x40)<<1)))
			if f_left_get_val(c) == 0 {
				c.reg8[FL_Z] = 1
			} else {
				c.reg8[FL_Z] = 0
			}
			c.reg8[FL_H] = 0
			c.reg8[FL_N] = 0

			c.do_instr(desc, ticks, args)
		}
	case "SRL":
		lambda = func(c *CPU) {
			prev := uint8(f_left_get_val(c))
			c.reg8[FL_C] = (prev & 0x01)
			f_set_val(c, uint16(prev>>1))
			if f_left_get_val(c) == 0 {
				c.reg8[FL_Z] = 1
			} else {
				c.reg8[FL_Z] = 0
			}
			c.reg8[FL_H] = 0
			c.reg8[FL_N] = 0

			c.do_instr(desc, ticks, args)
		}

	case "SLA":
		lambda = func(c *CPU) {

			prev := uint8(f_left_get_val(c))
			c.reg8[FL_C] = (prev & 0x80) >> 7
			f_set_val(c, uint16(prev<<1))

			if f_left_get_val(c) == 0 {
				c.reg8[FL_Z] = 1
			} else {
				c.reg8[FL_Z] = 0
			}
			c.reg8[FL_H] = 0
			c.reg8[FL_N] = 0

			c.do_instr(desc, ticks, args)
		}

	case "RRCA":
		lambda = func(c *CPU) {
			prev := uint8(f_left_get_val(c))
			c.reg8[FL_C] = prev & 0x01
			f_set_val(c, uint16((prev>>1)+(c.reg8[FL_C]<<7)))

			if f_left_get_val(c) == 0 {
				c.reg8[FL_Z] = 1
			} else {
				c.reg8[FL_Z] = 0
			}
			c.reg8[FL_H] = 0
			c.reg8[FL_N] = 0

			c.do_instr(desc, ticks, args)
		}

	case "RR":
		lambda = func(c *CPU) {
			prev := uint8(f_left_get_val(c))
			temp := c.reg8[FL_C]
			c.reg8[FL_C] = (prev & 0x01)
			f_set_val(c, uint16((prev>>1)+(temp<<7)))

			if f_left_get_val(c) == 0 {
				c.reg8[FL_Z] = 1
			} else {
				c.reg8[FL_Z] = 0
			}
			c.reg8[FL_H] = 0
			c.reg8[FL_N] = 0

			c.do_instr(desc, ticks, args)
		}

	case "SCF":

		lambda = func(c *CPU) {
			c.reg8[FL_C] = 1
			c.reg8[FL_H] = 0
			c.reg8[FL_N] = 0
			c.do_instr(desc, ticks, args)

		}

	case "CCF":
		lambda = func(c *CPU) {
			c.reg8[FL_C] = ^c.reg8[FL_C] & 0x1
			c.reg8[FL_H] = 0
			c.reg8[FL_N] = 0

			c.do_instr(desc, ticks, args)

		}

	case "SWAP":
		lambda = func(c *CPU) {
			prev := f_left_get_val(c)
			lower := (prev & 0x0f)
			upper := (prev & 0xf0)
			val := (lower << 4) | (upper >> 4)
			f_set_val(c, val)

			if f_left_get_val(c) == 0 {
				c.reg8[FL_Z] = 1
			} else {
				c.reg8[FL_Z] = 0
			}
			c.reg8[FL_C] = 0
			c.reg8[FL_H] = 0
			c.reg8[FL_N] = 0

			c.do_instr(desc, ticks, args)
		}

	}

	return lambda

}

func gen_test_bit(bit uint8, reg_left string, ticks uint16, args uint16) Action {

	type_left := get_ld_type(reg_left)
	f_get_val := gen_get_val(type_left, reg_left)

	lambda := func(c *CPU) {
		val := f_get_val(c)
		if (val & (1 << bit)) == 0 {
			c.reg8[FL_Z] = 1
		} else {
			//not sure if we need to clear
			c.reg8[FL_Z] = 0x0
		}
		c.reg8[FL_N] = 0x0
		c.reg8[FL_H] = 0x1

		//+2 for cb cmds
		c.do_instr("TBIT "+string(bit)+" "+reg_left, ticks, args)
	}
	return lambda
}
func gen_set_bit(bit uint8, reg_left string, ticks uint16, args uint16) Action {

	type_left := get_ld_type(reg_left)
	f_get_val := gen_get_val(type_left, reg_left)
	f_set_val := gen_set_val(type_left, reg_left)
	desc := "SETB " + string(bit) + "," + reg_left

	lambda := func(c *CPU) {
		f_set_val(c, f_get_val(c)|(1<<bit))
		c.do_instr(desc, (ticks), (args))
	}
	return lambda
}

func gen_res_bit(bit uint8, reg_left string, ticks uint16, args uint16) Action {

	type_left := get_ld_type(reg_left)
	f_get_val := gen_get_val(type_left, reg_left)
	f_set_val := gen_set_val(type_left, reg_left)
	desc := "RESB " + string(bit) + "," + reg_left

	lambda := func(c *CPU) {
		f_set_val(c, f_get_val(c)&^(1<<bit))
		c.do_instr(desc, (ticks), (args))
	}
	return lambda
}

func gen_ld(reg_left string, reg_right string, ticks uint16, args uint16) Action {

	type_left := get_ld_type(reg_left)
	type_right := get_ld_type(reg_right)

	lambda := func(c *CPU) { fmt.Println("Undefined LD op ", reg_right, " ", reg_left) }

	desc := "LD " + reg_left + "," + reg_right

	f_get_val := gen_get_val(type_right, reg_right)
	f_set_val := gen_set_val(type_left, reg_left)

	if reg_left == "SP" && reg_right == "n" {

		lambda = func(c *CPU) {
			f_set_val(c, f_get_val(c))
			c.do_instr(desc, (ticks), (args))
		}

	} else {

		lambda = func(c *CPU) {
			f_set_val(c, f_get_val(c))
			c.do_instr(desc, (ticks), (args))
		}
	}

	return lambda
}

func gen_ldh(reg_left string, reg_right string, ticks uint16, args uint16) Action {

	lambda := func(c *CPU) { fmt.Println("Undefined LD op ", reg_right, " ", reg_left) }
	desc := "LDH" + reg_left + "," + reg_right

	if reg_left == "(n)" {

		lambda = func(c *CPU) {
			valn := c.mmu.Read_b(c.reg16[PC] + 1)
			c.mmu.Write_b(0xff00+uint16(valn), c.reg8[A])
			c.do_instr(desc, (ticks), (args))
			//	fmt.Printf("->%04x,%04x\n",0xff00+uint16(valn),valn)

		}
	} else if reg_left == "A" {

		lambda = func(c *CPU) {

			valn := c.mmu.Read_b(c.reg16[PC] + 1)
			valff := c.mmu.Read_b(0xff00 + uint16(valn))
			c.reg8[A] = valff
			c.do_instr(desc, (ticks), (args))
			//fmt.Printf("<-%04x,%04x\n",0xff00+uint16(valn),valn)

		}
	}
	return lambda
}

func createOps(c *CPU) {
	//Init registers
	/////////////////
	c.reg8[A] = 0
	c.reg8[B] = 0
	c.reg8[C] = 0
	c.reg8[D] = 0
	c.reg8[E] = 0
	c.reg8[F] = 0
	c.reg8[H] = 0
	c.reg8[L] = 0
	c.reg8[F] = 0
	c.reg8[FL_Z] = 0
	c.reg8[FL_C] = 0
	c.reg8[FL_N] = 0
	c.reg8[FL_H] = 0
	c.push_pc = gen_push_pop("PUSH", "PC")

	/////////////////

	//Generate opcode Map
	c.ops[0x7f] = gen_ld("A", "A", 4, 1)
	c.ops[0x78] = gen_ld("A", "B", 4, 1)
	c.ops[0x77] = gen_ld("A", "C", 4, 1)
	c.ops[0x7A] = gen_ld("A", "D", 4, 1)
	c.ops[0x7B] = gen_ld("A", "E", 4, 1)
	c.ops[0x7C] = gen_ld("A", "H", 4, 1)
	c.ops[0x7D] = gen_ld("A", "L", 4, 1)
	c.ops[0x0A] = gen_ld("A", "(BC)", 8, 1)
	c.ops[0x1A] = gen_ld("A", "(DE)", 8, 1)
	c.ops[0x7E] = gen_ld("A", "(HL)", 8, 1)
	c.ops[0xFA] = gen_ld("A", "(nn)", 16, 3)
	c.ops[0x3E] = gen_ld("A", "n", 8, 2)
	c.ops[0xF2] = gen_ld("A", "(C)", 8, 1)
	c.ops[0x79] = gen_ld("A", "C", 4, 1)

	c.ops[0x06] = gen_ld("B", "n", 8, 2)
	c.ops[0x40] = gen_ld("B", "B", 4, 1)
	c.ops[0x41] = gen_ld("B", "C", 4, 1)
	c.ops[0x42] = gen_ld("B", "D", 4, 1)
	c.ops[0x43] = gen_ld("B", "E", 4, 1)
	c.ops[0x44] = gen_ld("B", "H", 4, 1)
	c.ops[0x45] = gen_ld("B", "L", 4, 1)
	c.ops[0x46] = gen_ld("B", "(HL)", 8, 1)
	c.ops[0x47] = gen_ld("B", "A", 4, 1)
	c.ops[0x48] = gen_ld("C", "B", 4, 1)
	c.ops[0x49] = gen_ld("C", "C", 4, 1)
	c.ops[0x4A] = gen_ld("C", "D", 4, 1)
	c.ops[0x4B] = gen_ld("C", "E", 4, 1)
	c.ops[0x4C] = gen_ld("C", "H", 4, 1)
	c.ops[0x4D] = gen_ld("C", "L", 4, 1)
	c.ops[0x4E] = gen_ld("C", "(HL)", 8, 1)
	c.ops[0x4F] = gen_ld("C", "A", 4, 1)
	c.ops[0x0E] = gen_ld("C", "n", 8, 2)
	c.ops[0x50] = gen_ld("D", "B", 4, 1)
	c.ops[0x51] = gen_ld("D", "C", 4, 1)
	c.ops[0x52] = gen_ld("D", "D", 4, 1)
	c.ops[0x53] = gen_ld("D", "E", 4, 1)
	c.ops[0x54] = gen_ld("D", "H", 4, 1)
	c.ops[0x55] = gen_ld("D", "L", 4, 1)
	c.ops[0x56] = gen_ld("D", "(HL)", 8, 1)
	c.ops[0x57] = gen_ld("D", "A", 4, 1)
	c.ops[0x16] = gen_ld("D", "n", 8, 2)

	c.ops[0x58] = gen_ld("E", "B", 4, 1)
	c.ops[0x59] = gen_ld("E", "C", 4, 1)
	c.ops[0x5A] = gen_ld("E", "D", 4, 1)
	c.ops[0x5B] = gen_ld("E", "E", 4, 1)
	c.ops[0x5C] = gen_ld("E", "H", 4, 1)
	c.ops[0x5D] = gen_ld("E", "L", 4, 1)
	c.ops[0x5E] = gen_ld("E", "(HL)", 8, 1)
	c.ops[0x5F] = gen_ld("E", "A", 4, 1)
	c.ops[0x1E] = gen_ld("E", "n", 8, 2)

	c.ops[0x60] = gen_ld("H", "B", 4, 1)
	c.ops[0x61] = gen_ld("H", "C", 4, 1)
	c.ops[0x62] = gen_ld("H", "D", 4, 1)
	c.ops[0x63] = gen_ld("H", "E", 4, 1)
	c.ops[0x64] = gen_ld("H", "H", 4, 1)
	c.ops[0x65] = gen_ld("H", "L", 4, 1)
	c.ops[0x66] = gen_ld("H", "(HL)", 8, 1)
	c.ops[0x67] = gen_ld("H", "A", 4, 1)
	c.ops[0x26] = gen_ld("H", "n", 8, 2)

	c.ops[0x68] = gen_ld("L", "B", 4, 1)
	c.ops[0x69] = gen_ld("L", "C", 4, 1)
	c.ops[0x6A] = gen_ld("L", "D", 4, 1)
	c.ops[0x6B] = gen_ld("L", "E", 4, 1)
	c.ops[0x6C] = gen_ld("L", "H", 4, 1)
	c.ops[0x6D] = gen_ld("L", "L", 4, 1)
	c.ops[0x6E] = gen_ld("L", "(HL)", 8, 1)
	c.ops[0x6F] = gen_ld("L", "A", 4, 1)
	c.ops[0x2E] = gen_ld("L", "n", 8, 2)

	c.ops[0x70] = gen_ld("(HL)", "B", 8, 1)
	c.ops[0x71] = gen_ld("(HL)", "C", 8, 1)
	c.ops[0x72] = gen_ld("(HL)", "D", 8, 1)
	c.ops[0x73] = gen_ld("(HL)", "E", 8, 1)
	c.ops[0x74] = gen_ld("(HL)", "H", 8, 1)
	c.ops[0x75] = gen_ld("(HL)", "L", 8, 1)
	c.ops[0x77] = gen_ld("(HL)", "A", 8, 1)
	c.ops[0x36] = gen_ld("(HL)", "n", 12, 2)

	c.ops[0x02] = gen_ld("(BC)", "A", 8, 1)
	c.ops[0x12] = gen_ld("(DE)", "A", 8, 1)
	c.ops[0xEA] = gen_ld("(nn)8", "A", 16, 3)
	c.ops[0xE2] = gen_ld("(C)", "A", 8, 1)
	c.ops[0x3A] = gen_ld("A", "(HLD)", 8, 1)
	c.ops[0x32] = gen_ld("(HLD)", "A", 8, 1)
	c.ops[0x2A] = gen_ld("A", "(HLI)", 8, 1)
	c.ops[0x22] = gen_ld("(HLI)", "A", 8, 1)
	c.ops[0xE0] = gen_ldh("(n)", "A", 12, 2)
	c.ops[0xF0] = gen_ldh("A", "(n)", 12, 2)

	c.ops[0x01] = gen_ld("BC", "nn", 12, 3)
	c.ops[0x11] = gen_ld("DE", "nn", 12, 3)
	c.ops[0x21] = gen_ld("HL", "nn", 12, 3)
	c.ops[0x31] = gen_ld("SP", "nn", 12, 3)

	//special case for f8 op
	add_sp := gen_alu("ADD", "SP", "n", 12, 2)
	s_val := gen_set_val(reg16_combo, "HL")
	c.ops[0xf8] = func(c *CPU) {
		before := c.reg16[SP]
		add_sp(c)
		s_val(c, c.reg16[SP])
		c.reg16[SP] = before

	}

	c.ops[0xf9] = gen_ld("SP", "HL", 8, 1)
	c.ops[0x08] = gen_ld("(nn)", "SP", 20, 3)
	c.ops[0xf5] = gen_push_pop("PUSHAF", "AF")
	c.ops[0xC5] = gen_push_pop("PUSH", "BC")
	c.ops[0xD5] = gen_push_pop("PUSH", "DE")
	c.ops[0xE5] = gen_push_pop("PUSH", "HL")

	c.ops[0xF1] = gen_push_pop("POPAF", "AF")
	c.ops[0xC1] = gen_push_pop("POP", "BC")
	c.ops[0xD1] = gen_push_pop("POP", "DE")
	c.ops[0xE1] = gen_push_pop("POP", "HL")

	c.ops[0x27] = gen_alu("DAA", "A", "A", 4, 1)

	c.ops[0x87] = gen_alu("ADD", "A", "A", 4, 1)
	c.ops[0x80] = gen_alu("ADD", "A", "B", 4, 1)
	c.ops[0x81] = gen_alu("ADD", "A", "C", 4, 1)
	c.ops[0x82] = gen_alu("ADD", "A", "D", 4, 1)
	c.ops[0x83] = gen_alu("ADD", "A", "E", 4, 1)
	c.ops[0x84] = gen_alu("ADD", "A", "H", 4, 1)
	c.ops[0x85] = gen_alu("ADD", "A", "L", 4, 1)
	c.ops[0x86] = gen_alu("ADD", "A", "(HL)", 8, 1)
	c.ops[0xc6] = gen_alu("ADD", "A", "n", 8, 2)
	c.ops[0xe8] = gen_alu("ADD", "SP", "n", 16, 2)

	c.ops[0x8F] = gen_alu("ADC", "A", "A", 4, 1)
	c.ops[0x88] = gen_alu("ADC", "A", "B", 4, 1)
	c.ops[0x89] = gen_alu("ADC", "A", "C", 4, 1)
	c.ops[0x8A] = gen_alu("ADC", "A", "D", 4, 1)
	c.ops[0x8B] = gen_alu("ADC", "A", "E", 4, 1)
	c.ops[0x8C] = gen_alu("ADC", "A", "H", 4, 1)
	c.ops[0x8D] = gen_alu("ADC", "A", "L", 4, 1)
	c.ops[0x8E] = gen_alu("ADC", "A", "(HL)", 8, 1)
	c.ops[0xCE] = gen_alu("ADC", "A", "n", 8, 2)

	c.ops[0x9] = gen_alu("ADD16", "HL", "BC", 8, 1)
	c.ops[0x19] = gen_alu("ADD16", "HL", "DE", 8, 1)
	c.ops[0x29] = gen_alu("ADD16", "HL", "HL", 8, 1)
	c.ops[0x39] = gen_alu("ADD16", "HL", "SP", 8, 1)

	c.ops[0x97] = gen_alu("SUB", "A", "A", 4, 1)
	c.ops[0x90] = gen_alu("SUB", "A", "B", 4, 1)
	c.ops[0x91] = gen_alu("SUB", "A", "C", 4, 1)
	c.ops[0x92] = gen_alu("SUB", "A", "D", 4, 1)
	c.ops[0x93] = gen_alu("SUB", "A", "E", 4, 1)
	c.ops[0x94] = gen_alu("SUB", "A", "H", 4, 1)
	c.ops[0x95] = gen_alu("SUB", "A", "L", 4, 1)
	c.ops[0x96] = gen_alu("SUB", "A", "(HL)", 8, 1)
	c.ops[0xD6] = gen_alu("SUB", "A", "n", 8, 2)

	c.ops[0x9F] = gen_alu("SBC", "A", "A", 4, 1)
	c.ops[0x98] = gen_alu("SBC", "A", "B", 4, 1)
	c.ops[0x99] = gen_alu("SBC", "A", "C", 4, 1)
	c.ops[0x9A] = gen_alu("SBC", "A", "D", 4, 1)
	c.ops[0x9B] = gen_alu("SBC", "A", "E", 4, 1)
	c.ops[0x9C] = gen_alu("SBC", "A", "H", 4, 1)
	c.ops[0x9D] = gen_alu("SBC", "A", "L", 4, 1)
	c.ops[0x9E] = gen_alu("SBC", "A", "(HL)", 8, 1)
	c.ops[0xDE] = gen_alu("SBC", "A", "n", 8, 2)

	c.ops[0xA7] = gen_alu("AND", "A", "A", 4, 1)
	c.ops[0xA0] = gen_alu("AND", "A", "B", 4, 1)
	c.ops[0xA1] = gen_alu("AND", "A", "C", 4, 1)
	c.ops[0xA2] = gen_alu("AND", "A", "D", 4, 1)
	c.ops[0xA3] = gen_alu("AND", "A", "E", 4, 1)
	c.ops[0xA4] = gen_alu("AND", "A", "H", 4, 1)
	c.ops[0xA5] = gen_alu("AND", "A", "L", 4, 1)
	c.ops[0xA6] = gen_alu("AND", "A", "(HL)", 8, 1)
	c.ops[0xE6] = gen_alu("AND", "A", "n", 8, 2)

	c.ops[0xB7] = gen_alu("OR", "A", "A", 4, 1)
	c.ops[0xB0] = gen_alu("OR", "A", "B", 4, 1)
	c.ops[0xB1] = gen_alu("OR", "A", "C", 4, 1)
	c.ops[0xB2] = gen_alu("OR", "A", "D", 4, 1)
	c.ops[0xB3] = gen_alu("OR", "A", "E", 4, 1)
	c.ops[0xB4] = gen_alu("OR", "A", "H", 4, 1)
	c.ops[0xB5] = gen_alu("OR", "A", "L", 4, 1)
	c.ops[0xB6] = gen_alu("OR", "A", "(HL)", 8, 1)
	c.ops[0xF6] = gen_alu("OR", "A", "n", 8, 2)

	c.ops[0xAF] = gen_alu("XOR", "A", "A", 4, 1)
	c.ops[0xA8] = gen_alu("XOR", "A", "B", 4, 1)
	c.ops[0xA9] = gen_alu("XOR", "A", "C", 4, 1)
	c.ops[0xAA] = gen_alu("XOR", "A", "D", 4, 1)
	c.ops[0xAB] = gen_alu("XOR", "A", "E", 4, 1)
	c.ops[0xAC] = gen_alu("XOR", "A", "H", 4, 1)
	c.ops[0xAD] = gen_alu("XOR", "A", "L", 4, 1)
	c.ops[0xAE] = gen_alu("XOR", "A", "(HL)", 8, 1)
	c.ops[0xEE] = gen_alu("XOR", "A", "n", 8, 2)

	c.ops[0xBF] = gen_alu("CP", "A", "A", 4, 1)
	c.ops[0xB8] = gen_alu("CP", "A", "B", 4, 1)
	c.ops[0xB9] = gen_alu("CP", "A", "C", 4, 1)
	c.ops[0xBA] = gen_alu("CP", "A", "D", 4, 1)
	c.ops[0xBB] = gen_alu("CP", "A", "E", 4, 1)
	c.ops[0xBC] = gen_alu("CP", "A", "H", 4, 1)
	c.ops[0xBD] = gen_alu("CP", "A", "L", 4, 1)
	c.ops[0xBE] = gen_alu("CP", "A", "(HL)", 8, 1)
	c.ops[0xFE] = gen_alu("CP", "A", "n", 8, 2)

	c.ops[0x3D] = gen_alu("DEC", "A", "", 4, 1)
	c.ops[0x05] = gen_alu("DEC", "B", "", 4, 1)
	c.ops[0x0D] = gen_alu("DEC", "C", "", 4, 1)
	c.ops[0x15] = gen_alu("DEC", "D", "", 4, 1)
	c.ops[0x1D] = gen_alu("DEC", "E", "", 4, 1)
	c.ops[0x25] = gen_alu("DEC", "H", "", 4, 1)
	c.ops[0x2D] = gen_alu("DEC", "L", "", 4, 1)
	c.ops[0x35] = gen_alu("DEC", "(HL)", "", 12, 1)
	c.ops[0x3b] = gen_alu("DEC", "SP", "", 8, 1)
	c.ops[0x1B] = gen_alu("DEC", "DE", "", 8, 1)
	c.ops[0x2B] = gen_alu("DEC", "HL", "", 8, 1)
	c.ops[0x0B] = gen_alu("DEC", "BC", "", 8, 1)

	c.ops[0x3c] = gen_alu("INC", "A", "", 4, 1)
	c.ops[0x04] = gen_alu("INC", "B", "", 4, 1)
	c.ops[0x0c] = gen_alu("INC", "C", "", 4, 1)
	c.ops[0x14] = gen_alu("INC", "D", "", 4, 1)
	c.ops[0x1c] = gen_alu("INC", "E", "", 4, 1)
	c.ops[0x24] = gen_alu("INC", "H", "", 4, 1)
	c.ops[0x2c] = gen_alu("INC", "L", "", 4, 1)
	c.ops[0x33] = gen_alu("INC", "SP", "", 8, 1)
	c.ops[0x13] = gen_alu("INC", "DE", "", 8, 1)
	c.ops[0x23] = gen_alu("INC", "HL", "", 8, 1)
	c.ops[0x03] = gen_alu("INC", "BC", "", 8, 1)
	c.ops[0x34] = gen_alu("INC", "(HL)", "", 12, 1)

	c.ops[0xCB40] = gen_test_bit(0, "B", 8, 2)
	c.ops[0xCB41] = gen_test_bit(0, "C", 8, 2)
	c.ops[0xCB42] = gen_test_bit(0, "D", 8, 2)
	c.ops[0xCB43] = gen_test_bit(0, "E", 8, 2)
	c.ops[0xCB44] = gen_test_bit(0, "H", 8, 2)
	c.ops[0xCB45] = gen_test_bit(0, "L", 8, 2)
	c.ops[0xCB46] = gen_test_bit(0, "(HL)", 12, 2)
	c.ops[0xCB47] = gen_test_bit(0, "A", 8, 2)
	c.ops[0xCB48] = gen_test_bit(1, "B", 8, 2)
	c.ops[0xCB49] = gen_test_bit(1, "C", 8, 2)
	c.ops[0xCB4A] = gen_test_bit(1, "D", 8, 2)
	c.ops[0xCB4B] = gen_test_bit(1, "E", 8, 2)
	c.ops[0xCB4C] = gen_test_bit(1, "H", 8, 2)
	c.ops[0xCB4D] = gen_test_bit(1, "L", 8, 2)
	c.ops[0xCB4E] = gen_test_bit(1, "(HL)", 12, 2)
	c.ops[0xCB4F] = gen_test_bit(1, "A", 8, 2)
	c.ops[0xCB50] = gen_test_bit(2, "B", 8, 2)
	c.ops[0xCB51] = gen_test_bit(2, "C", 8, 2)
	c.ops[0xCB52] = gen_test_bit(2, "D", 8, 2)
	c.ops[0xCB53] = gen_test_bit(2, "E", 8, 2)
	c.ops[0xCB54] = gen_test_bit(2, "H", 8, 2)
	c.ops[0xCB55] = gen_test_bit(2, "L", 8, 2)
	c.ops[0xCB56] = gen_test_bit(2, "(HL)", 12, 2)
	c.ops[0xCB57] = gen_test_bit(2, "A", 8, 2)
	c.ops[0xCB58] = gen_test_bit(3, "B", 8, 2)
	c.ops[0xCB59] = gen_test_bit(3, "C", 8, 2)
	c.ops[0xCB5A] = gen_test_bit(3, "D", 8, 2)
	c.ops[0xCB5B] = gen_test_bit(3, "E", 8, 2)
	c.ops[0xCB5C] = gen_test_bit(3, "H", 8, 2)
	c.ops[0xCB5D] = gen_test_bit(3, "L", 8, 2)
	c.ops[0xCB5E] = gen_test_bit(3, "(HL)", 12, 2)
	c.ops[0xCB5F] = gen_test_bit(3, "A", 8, 2)
	c.ops[0xCB60] = gen_test_bit(4, "B", 8, 2)
	c.ops[0xCB61] = gen_test_bit(4, "C", 8, 2)
	c.ops[0xCB62] = gen_test_bit(4, "D", 8, 2)
	c.ops[0xCB63] = gen_test_bit(4, "E", 8, 2)
	c.ops[0xCB64] = gen_test_bit(4, "H", 8, 2)
	c.ops[0xCB65] = gen_test_bit(4, "L", 8, 2)
	c.ops[0xCB66] = gen_test_bit(4, "(HL)", 12, 2)
	c.ops[0xCB67] = gen_test_bit(4, "A", 8, 2)
	c.ops[0xCB68] = gen_test_bit(5, "B", 8, 2)
	c.ops[0xCB69] = gen_test_bit(5, "C", 8, 2)
	c.ops[0xCB6A] = gen_test_bit(5, "D", 8, 2)
	c.ops[0xCB6B] = gen_test_bit(5, "E", 8, 2)
	c.ops[0xCB6C] = gen_test_bit(5, "H", 8, 2)
	c.ops[0xCB6D] = gen_test_bit(5, "L", 8, 2)
	c.ops[0xCB6E] = gen_test_bit(5, "(HL)", 12, 2)
	c.ops[0xCB6F] = gen_test_bit(5, "A", 8, 2)
	c.ops[0xCB70] = gen_test_bit(6, "B", 8, 2)
	c.ops[0xCB71] = gen_test_bit(6, "C", 8, 2)
	c.ops[0xCB72] = gen_test_bit(6, "D", 8, 2)
	c.ops[0xCB73] = gen_test_bit(6, "E", 8, 2)
	c.ops[0xCB74] = gen_test_bit(6, "H", 8, 2)
	c.ops[0xCB75] = gen_test_bit(6, "L", 8, 2)
	c.ops[0xCB76] = gen_test_bit(6, "(HL)", 12, 2)
	c.ops[0xCB77] = gen_test_bit(6, "A", 8, 2)
	c.ops[0xCB78] = gen_test_bit(7, "B", 8, 2)
	c.ops[0xCB79] = gen_test_bit(7, "C", 8, 2)
	c.ops[0xCB7A] = gen_test_bit(7, "D", 8, 2)
	c.ops[0xCB7B] = gen_test_bit(7, "E", 8, 2)
	c.ops[0xCB7C] = gen_test_bit(7, "H", 8, 2)
	c.ops[0xCB7D] = gen_test_bit(7, "L", 8, 2)
	c.ops[0xCB7E] = gen_test_bit(7, "(HL)", 12, 2)
	c.ops[0xCB7F] = gen_test_bit(7, "A", 8, 2)
	c.ops[0xCB80] = gen_res_bit(0, "B", 8, 2)
	c.ops[0xCB81] = gen_res_bit(0, "C", 8, 2)
	c.ops[0xCB82] = gen_res_bit(0, "D", 8, 2)
	c.ops[0xCB83] = gen_res_bit(0, "E", 8, 2)
	c.ops[0xCB84] = gen_res_bit(0, "H", 8, 2)
	c.ops[0xCB85] = gen_res_bit(0, "L", 8, 2)
	c.ops[0xCB86] = gen_res_bit(0, "(HL)", 16, 2)
	c.ops[0xCB87] = gen_res_bit(0, "A", 8, 2)
	c.ops[0xCB88] = gen_res_bit(1, "B", 8, 2)
	c.ops[0xCB89] = gen_res_bit(1, "C", 8, 2)
	c.ops[0xCB8A] = gen_res_bit(1, "D", 8, 2)
	c.ops[0xCB8B] = gen_res_bit(1, "E", 8, 2)
	c.ops[0xCB8C] = gen_res_bit(1, "H", 8, 2)
	c.ops[0xCB8D] = gen_res_bit(1, "L", 8, 2)
	c.ops[0xCB8E] = gen_res_bit(1, "(HL)", 16, 2)
	c.ops[0xCB8F] = gen_res_bit(1, "A", 8, 2)
	c.ops[0xCB90] = gen_res_bit(2, "B", 8, 2)
	c.ops[0xCB91] = gen_res_bit(2, "C", 8, 2)
	c.ops[0xCB92] = gen_res_bit(2, "D", 8, 2)
	c.ops[0xCB93] = gen_res_bit(2, "E", 8, 2)
	c.ops[0xCB94] = gen_res_bit(2, "H", 8, 2)
	c.ops[0xCB95] = gen_res_bit(2, "L", 8, 2)
	c.ops[0xCB96] = gen_res_bit(2, "(HL)", 16, 2)
	c.ops[0xCB97] = gen_res_bit(2, "A", 8, 2)
	c.ops[0xCB98] = gen_res_bit(3, "B", 8, 2)
	c.ops[0xCB99] = gen_res_bit(3, "C", 8, 2)
	c.ops[0xCB9A] = gen_res_bit(3, "D", 8, 2)
	c.ops[0xCB9B] = gen_res_bit(3, "E", 8, 2)
	c.ops[0xCB9C] = gen_res_bit(3, "H", 8, 2)
	c.ops[0xCB9D] = gen_res_bit(3, "L", 8, 2)
	c.ops[0xCB9E] = gen_res_bit(3, "(HL)", 16, 2)
	c.ops[0xCB9F] = gen_res_bit(3, "A", 8, 2)
	c.ops[0xCBA0] = gen_res_bit(4, "B", 8, 2)
	c.ops[0xCBA1] = gen_res_bit(4, "C", 8, 2)
	c.ops[0xCBA2] = gen_res_bit(4, "D", 8, 2)
	c.ops[0xCBA3] = gen_res_bit(4, "E", 8, 2)
	c.ops[0xCBA4] = gen_res_bit(4, "H", 8, 2)
	c.ops[0xCBA5] = gen_res_bit(4, "L", 8, 2)
	c.ops[0xCBA6] = gen_res_bit(4, "(HL)", 16, 2)
	c.ops[0xCBA7] = gen_res_bit(4, "A", 8, 2)
	c.ops[0xCBA8] = gen_res_bit(5, "B", 8, 2)
	c.ops[0xCBA9] = gen_res_bit(5, "C", 8, 2)
	c.ops[0xCBAA] = gen_res_bit(5, "D", 8, 2)
	c.ops[0xCBAB] = gen_res_bit(5, "E", 8, 2)
	c.ops[0xCBAC] = gen_res_bit(5, "H", 8, 2)
	c.ops[0xCBAD] = gen_res_bit(5, "L", 8, 2)
	c.ops[0xCBAE] = gen_res_bit(5, "(HL)", 16, 2)
	c.ops[0xCBAF] = gen_res_bit(5, "A", 8, 2)
	c.ops[0xCBB0] = gen_res_bit(6, "B", 8, 2)
	c.ops[0xCBB1] = gen_res_bit(6, "C", 8, 2)
	c.ops[0xCBB2] = gen_res_bit(6, "D", 8, 2)
	c.ops[0xCBB3] = gen_res_bit(6, "E", 8, 2)
	c.ops[0xCBB4] = gen_res_bit(6, "H", 8, 2)
	c.ops[0xCBB5] = gen_res_bit(6, "L", 8, 2)
	c.ops[0xCBB6] = gen_res_bit(6, "(HL)", 16, 2)
	c.ops[0xCBB7] = gen_res_bit(6, "A", 8, 2)
	c.ops[0xCBB8] = gen_res_bit(7, "B", 8, 2)
	c.ops[0xCBB9] = gen_res_bit(7, "C", 8, 2)
	c.ops[0xCBBA] = gen_res_bit(7, "D", 8, 2)
	c.ops[0xCBBB] = gen_res_bit(7, "E", 8, 2)
	c.ops[0xCBBC] = gen_res_bit(7, "H", 8, 2)
	c.ops[0xCBBD] = gen_res_bit(7, "L", 8, 2)
	c.ops[0xCBBE] = gen_res_bit(7, "(HL)", 16, 2)
	c.ops[0xCBBF] = gen_res_bit(7, "A", 8, 2)
	c.ops[0xCBC0] = gen_set_bit(0, "B", 8, 2)
	c.ops[0xCBC1] = gen_set_bit(0, "C", 8, 2)
	c.ops[0xCBC2] = gen_set_bit(0, "D", 8, 2)
	c.ops[0xCBC3] = gen_set_bit(0, "E", 8, 2)
	c.ops[0xCBC4] = gen_set_bit(0, "H", 8, 2)
	c.ops[0xCBC5] = gen_set_bit(0, "L", 8, 2)
	c.ops[0xCBC6] = gen_set_bit(0, "(HL)", 16, 2)
	c.ops[0xCBC7] = gen_set_bit(0, "A", 8, 2)
	c.ops[0xCBC8] = gen_set_bit(1, "B", 8, 2)
	c.ops[0xCBC9] = gen_set_bit(1, "C", 8, 2)
	c.ops[0xCBCA] = gen_set_bit(1, "D", 8, 2)
	c.ops[0xCBCB] = gen_set_bit(1, "E", 8, 2)
	c.ops[0xCBCC] = gen_set_bit(1, "H", 8, 2)
	c.ops[0xCBCD] = gen_set_bit(1, "L", 8, 2)
	c.ops[0xCBCE] = gen_set_bit(1, "(HL)", 16, 2)
	c.ops[0xCBCF] = gen_set_bit(1, "A", 8, 2)
	c.ops[0xCBD0] = gen_set_bit(2, "B", 8, 2)
	c.ops[0xCBD1] = gen_set_bit(2, "C", 8, 2)
	c.ops[0xCBD2] = gen_set_bit(2, "D", 8, 2)
	c.ops[0xCBD3] = gen_set_bit(2, "E", 8, 2)
	c.ops[0xCBD4] = gen_set_bit(2, "H", 8, 2)
	c.ops[0xCBD5] = gen_set_bit(2, "L", 8, 2)
	c.ops[0xCBD6] = gen_set_bit(2, "(HL)", 16, 2)
	c.ops[0xCBD7] = gen_set_bit(2, "A", 8, 2)
	c.ops[0xCBD8] = gen_set_bit(3, "B", 8, 2)
	c.ops[0xCBD9] = gen_set_bit(3, "C", 8, 2)
	c.ops[0xCBDA] = gen_set_bit(3, "D", 8, 2)
	c.ops[0xCBDB] = gen_set_bit(3, "E", 8, 2)
	c.ops[0xCBDC] = gen_set_bit(3, "H", 8, 2)
	c.ops[0xCBDD] = gen_set_bit(3, "L", 8, 2)
	c.ops[0xCBDE] = gen_set_bit(3, "(HL)", 16, 2)
	c.ops[0xCBDF] = gen_set_bit(3, "A", 8, 2)
	c.ops[0xCBE0] = gen_set_bit(4, "B", 8, 2)
	c.ops[0xCBE1] = gen_set_bit(4, "C", 8, 2)
	c.ops[0xCBE2] = gen_set_bit(4, "D", 8, 2)
	c.ops[0xCBE3] = gen_set_bit(4, "E", 8, 2)
	c.ops[0xCBE4] = gen_set_bit(4, "H", 8, 2)
	c.ops[0xCBE5] = gen_set_bit(4, "L", 8, 2)
	c.ops[0xCBE6] = gen_set_bit(4, "(HL)", 16, 2)
	c.ops[0xCBE7] = gen_set_bit(4, "A", 8, 2)
	c.ops[0xCBE8] = gen_set_bit(5, "B", 8, 2)
	c.ops[0xCBE9] = gen_set_bit(5, "C", 8, 2)
	c.ops[0xCBEA] = gen_set_bit(5, "D", 8, 2)
	c.ops[0xCBEB] = gen_set_bit(5, "E", 8, 2)
	c.ops[0xCBEC] = gen_set_bit(5, "H", 8, 2)
	c.ops[0xCBED] = gen_set_bit(5, "L", 8, 2)
	c.ops[0xCBEE] = gen_set_bit(5, "(HL)", 16, 2)
	c.ops[0xCBEF] = gen_set_bit(5, "A", 8, 2)
	c.ops[0xCBF0] = gen_set_bit(6, "B", 8, 2)
	c.ops[0xCBF1] = gen_set_bit(6, "C", 8, 2)
	c.ops[0xCBF2] = gen_set_bit(6, "D", 8, 2)
	c.ops[0xCBF3] = gen_set_bit(6, "E", 8, 2)
	c.ops[0xCBF4] = gen_set_bit(6, "H", 8, 2)
	c.ops[0xCBF5] = gen_set_bit(6, "L", 8, 2)
	c.ops[0xCBF6] = gen_set_bit(6, "(HL)", 16, 2)
	c.ops[0xCBF7] = gen_set_bit(6, "A", 8, 2)
	c.ops[0xCBF8] = gen_set_bit(7, "B", 8, 2)
	c.ops[0xCBF9] = gen_set_bit(7, "C", 8, 2)
	c.ops[0xCBFA] = gen_set_bit(7, "D", 8, 2)
	c.ops[0xCBFB] = gen_set_bit(7, "E", 8, 2)
	c.ops[0xCBFC] = gen_set_bit(7, "H", 8, 2)
	c.ops[0xCBFD] = gen_set_bit(7, "L", 8, 2)
	c.ops[0xCBFE] = gen_set_bit(7, "(HL)", 16, 2)
	c.ops[0xCBFF] = gen_set_bit(7, "A", 8, 2)
	f_rlc := gen_rotate_shift("RLC", "A", 4, 1)
	f_rl := gen_rotate_shift("RL", "A", 4, 1)

	c.ops[0x07] = func(c *CPU) { f_rlc(c); c.reg8[FL_Z] = 0 }
	c.ops[0x17] = func(c *CPU) { f_rl(c); c.reg8[FL_Z] = 0 }
	c.ops[0xCB17] = gen_rotate_shift("RL", "A", 8, 2)
	c.ops[0xCB10] = gen_rotate_shift("RL", "B", 8, 2)
	c.ops[0xCB11] = gen_rotate_shift("RL", "C", 8, 2)
	c.ops[0xCB12] = gen_rotate_shift("RL", "D", 8, 2)
	c.ops[0xCB13] = gen_rotate_shift("RL", "E", 8, 2)
	c.ops[0xCB14] = gen_rotate_shift("RL", "H", 8, 2)
	c.ops[0xCB15] = gen_rotate_shift("RL", "L", 8, 2)
	c.ops[0xCB16] = gen_rotate_shift("RL", "(HL)", 16, 2)
	c.ops[0xCB19] = gen_rotate_shift("RR", "C", 8, 2)

	f_rrc := gen_rotate_shift("RRC", "A", 4, 1)
	f_rr := gen_rotate_shift("RR", "A", 4, 1)

	c.ops[0x0f] = func(c *CPU) { f_rrc(c); c.reg8[FL_Z] = 0 }
	c.ops[0x1F] = func(c *CPU) { f_rr(c); c.reg8[FL_Z] = 0 }
	c.ops[0xCB0f] = gen_rotate_shift("RRC", "A", 8, 2)
	c.ops[0xCB08] = gen_rotate_shift("RRC", "B", 8, 2)
	c.ops[0xCB09] = gen_rotate_shift("RRC", "C", 8, 2)
	c.ops[0xCB0A] = gen_rotate_shift("RRC", "D", 8, 2)
	c.ops[0xCB0B] = gen_rotate_shift("RRC", "E", 8, 2)
	c.ops[0xCB0C] = gen_rotate_shift("RRC", "H", 8, 2)
	c.ops[0xCB0D] = gen_rotate_shift("RRC", "L", 8, 2)
	c.ops[0xCB0E] = gen_rotate_shift("RRC", "(HL)", 16, 2)

	c.ops[0xCB1F] = gen_rotate_shift("RR", "A", 8, 2)
	c.ops[0xCB18] = gen_rotate_shift("RR", "B", 8, 2)
	c.ops[0xCB19] = gen_rotate_shift("RR", "C", 8, 2)
	c.ops[0xCB1A] = gen_rotate_shift("RR", "D", 8, 2)
	c.ops[0xCB1B] = gen_rotate_shift("RR", "E", 8, 2)
	c.ops[0xCB1C] = gen_rotate_shift("RR", "H", 8, 2)
	c.ops[0xCB1D] = gen_rotate_shift("RR", "L", 8, 2)
	c.ops[0xCB1E] = gen_rotate_shift("RR", "(HL)", 16, 2)

	c.ops[0xCB27] = gen_rotate_shift("SLA", "A", 8, 2)
	c.ops[0xCB20] = gen_rotate_shift("SLA", "B", 8, 2)
	c.ops[0xCB21] = gen_rotate_shift("SLA", "C", 8, 2)
	c.ops[0xCB22] = gen_rotate_shift("SLA", "D", 8, 2)
	c.ops[0xCB23] = gen_rotate_shift("SLA", "E", 8, 2)
	c.ops[0xCB24] = gen_rotate_shift("SLA", "H", 8, 2)
	c.ops[0xCB25] = gen_rotate_shift("SLA", "L", 8, 2)
	c.ops[0xCB26] = gen_rotate_shift("SLA", "(HL)", 16, 2)

	c.ops[0xCB07] = gen_rotate_shift("RLC", "A", 8, 2)
	c.ops[0xCB00] = gen_rotate_shift("RLC", "B", 8, 2)
	c.ops[0xCB01] = gen_rotate_shift("RLC", "C", 8, 2)
	c.ops[0xCB02] = gen_rotate_shift("RLC", "D", 8, 2)
	c.ops[0xCB03] = gen_rotate_shift("RLC", "E", 8, 2)
	c.ops[0xCB04] = gen_rotate_shift("RLC", "H", 8, 2)
	c.ops[0xCB05] = gen_rotate_shift("RLC", "L", 8, 2)
	c.ops[0xCB06] = gen_rotate_shift("RLC", "(HL)", 16, 2)

	c.ops[0xCB2F] = gen_rotate_shift("SRA", "A", 8, 2)
	c.ops[0xCB28] = gen_rotate_shift("SRA", "B", 8, 2)
	c.ops[0xCB29] = gen_rotate_shift("SRA", "C", 8, 2)
	c.ops[0xCB2A] = gen_rotate_shift("SRA", "D", 8, 2)
	c.ops[0xCB2B] = gen_rotate_shift("SRA", "E", 8, 2)
	c.ops[0xCB2C] = gen_rotate_shift("SRA", "H", 8, 2)
	c.ops[0xCB2D] = gen_rotate_shift("SRA", "L", 8, 2)
	c.ops[0xCB2E] = gen_rotate_shift("SRA", "(HL)", 16, 2)

	c.ops[0xCB3F] = gen_rotate_shift("SRL", "A", 8, 2)
	c.ops[0xCB38] = gen_rotate_shift("SRL", "B", 8, 2)
	c.ops[0xCB39] = gen_rotate_shift("SRL", "C", 8, 2)
	c.ops[0xCB3A] = gen_rotate_shift("SRL", "D", 8, 2)
	c.ops[0xCB3B] = gen_rotate_shift("SRL", "E", 8, 2)
	c.ops[0xCB3C] = gen_rotate_shift("SRL", "H", 8, 2)
	c.ops[0xCB3D] = gen_rotate_shift("SRL", "L", 8, 2)
	c.ops[0xCB3E] = gen_rotate_shift("SRL", "(HL)", 16, 2)

	c.ops[0xCB37] = gen_rotate_shift("SWAP", "A", 8, 2)
	c.ops[0xCB30] = gen_rotate_shift("SWAP", "B", 8, 2)
	c.ops[0xCB31] = gen_rotate_shift("SWAP", "C", 8, 2)
	c.ops[0xCB32] = gen_rotate_shift("SWAP", "D", 8, 2)
	c.ops[0xCB33] = gen_rotate_shift("SWAP", "E", 8, 2)
	c.ops[0xCB34] = gen_rotate_shift("SWAP", "H", 8, 2)
	c.ops[0xCB35] = gen_rotate_shift("SWAP", "L", 8, 2)
	c.ops[0xCB36] = gen_rotate_shift("SWAP", "(HL)", 16, 2)

	//ABS JUMPS
	c.ops[0xc3] = gen_jmp("JPA", "nn", 1, 0, 0, 0, 16, 16, 3) //no args always jump
	c.ops[0xc2] = gen_jmp("JPNZ", "nn", 0, 0, 0x80, 0, 12, 16, 3)
	c.ops[0xcA] = gen_jmp("JPZ", "nn", 0, 0, 0x80, 1, 12, 16, 3)
	c.ops[0xD2] = gen_jmp("JPNC", "nn", 0, 0, 0x10, 0, 12, 16, 3)
	c.ops[0xDA] = gen_jmp("JPC", "nn", 0, 0, 0x10, 1, 12, 16, 3)
	c.ops[0xE9] = gen_jmp("JP", "HL", 1, 0, 0, 0, 4, 4, 1)
	//Relative jmp
	c.ops[0x20] = gen_jmp("JPNZ REL", "n", 0, 1, 0x80, 0, 8, 12, 2)
	c.ops[0x28] = gen_jmp("JPZ REL", "n", 0, 1, 0x80, 1, 8, 12, 2)
	c.ops[0x30] = gen_jmp("JPNC REL", "n", 0, 1, 0x10, 0, 8, 12, 2)
	c.ops[0x38] = gen_jmp("JPC REL", "n", 0, 1, 0x10, 1, 8, 12, 2)
	c.ops[0x18] = gen_jmp("JP", "n", 1, 1, 0, 0, 12, 12, 2) //signed
	c.ops[0xCD] = gen_call("CALL", "nn", 1, 0, 0, 0, 24, 24, 3)
	c.ops[0xC4] = gen_call("CALLNZ", "nn", 0, 0, 0x80, 0, 12, 24, 3)
	c.ops[0xCC] = gen_call("CALLZ", "nn", 0, 0, 0x80, 1, 12, 24, 3)
	c.ops[0xD4] = gen_call("CALLNC", "nn", 0, 0, 0x10, 0, 12, 24, 3)
	c.ops[0xDC] = gen_call("CALLC", "nn", 0, 0, 0x10, 1, 12, 24, 3)

	c.ops[0xC9] = gen_ret("RET", 1, 0, 0, 16, 16)
	c.ops[0xC0] = gen_ret("RET NZ", 0, 0x80, 0, 8, 20)
	c.ops[0xC8] = gen_ret("RET Z", 0, 0x80, 1, 8, 20)
	c.ops[0xD0] = gen_ret("RET NC", 0, 0x10, 0, 8, 20)
	c.ops[0xD8] = gen_ret("RET C", 0, 0x10, 1, 8, 20)

	f_reti := gen_ret("RETI", 1, 0, 0, 16, 16)
	//enable EI and return
	c.ops[0xD9] = func(c *CPU) { fmt.Println("EI"); c.Dump();  c.reg8[EI] = 1; f_reti(c); c.do_instr("RST", 16, 0) }

	c.ops[0x0] = func(c *CPU) { c.do_instr("NOP", 4, 1) }
	c.ops[0x10] = func(c *CPU) { c.set_sswitch(); c.do_instr("STOP", 4, 1) }
	c.ops[0xFB] = func(c *CPU) {c.ei_wait_instr=true; c.reg8[EI] = 1; c.do_instr("EI", 4, 1) }

	c.ops[0xF3] = func(c *CPU) { c.reg8[EI] = 0; c.do_instr("DI", 4, 1) }

	c.ops[0x2F] = func(c *CPU) {
		c.reg8[FL_H] = 1
		c.reg8[FL_N] = 1
		c.reg8[A] = ^c.reg8[A]
		c.do_instr("CPL", 4, 1)
	}

	f := gen_push_pop("PUSH", "PC")

	c.ops[0xc7] = func(c *CPU) { c.reg16[PC]++; f(c); c.reg16[PC] = 0; c.do_instr("RST", 16, 0) }
	c.ops[0xCF] = func(c *CPU) { c.reg16[PC]++; f(c); c.reg16[PC] = 0x8; c.do_instr("RST", 16, 0) }
	c.ops[0xD7] = func(c *CPU) { c.reg16[PC]++; f(c); c.reg16[PC] = 0x10; c.do_instr("RST", 16, 0) }
	c.ops[0xDF] = func(c *CPU) { c.reg16[PC]++; f(c); c.reg16[PC] = 0x18; c.do_instr("RST", 16, 0) }
	c.ops[0xE7] = func(c *CPU) { c.reg16[PC]++; f(c); c.reg16[PC] = 0x20; c.do_instr("RST", 16, 0) }
	c.ops[0xEF] = func(c *CPU) { c.reg16[PC]++; f(c); c.reg16[PC] = 0x28; c.do_instr("RST", 16, 0) }
	c.ops[0xF7] = func(c *CPU) { c.reg16[PC]++; f(c); c.reg16[PC] = 0x30; c.do_instr("RST", 16, 0) }
	c.ops[0xFF] = func(c *CPU) { c.reg16[PC]++; f(c); c.reg16[PC] = 0x38; c.do_instr("RST", 16, 0) }

	c.ops[0x76] = func(c *CPU) { c.is_halted = true; c.do_instr("HALT", 4, 1) }
	c.ops[0x37] = gen_rotate_shift("SCF", "A", 4, 1)
	c.ops[0x3f] = gen_rotate_shift("CCF", "A", 4, 1)

	//c.ops[0x27   ] =  func(c* CPU) {c.do_instr("NOOP", 4, 1)}
}

func (c *CPU) Get_reg_list() component.RegList {
	return c.reg_list
}
func setup_mmu_conn(c component.MMIOComponent, m *mmu.MMU) {
	reg_list := c.Get_reg_list()
	for i := range reg_list {
		m.Connect_mmio(reg_list[i].Addr, reg_list[i].Name, c)
	}
}
func setup_mmu_range_conn(c component.MemComponent, m *mmu.MMU) {
	range_list := c.Get_range_list()
	for i := range range_list {
		m.Connect_range(range_list[i], c)
	}
}

func NewCpu(listen bool, connect string, scale int, serial_p string,debug int,maxfps bool) *CPU {
	c := new(CPU)
	c.reg_list = component.RegList{
		{Name: "KEY1", Addr: KEY1_MMIO},
		{Name: "DIV", Addr: DIV_MMIO},
	}

	c.mmu = mmu.NewMMU(debug)
	c.timer = timer.NewTimer()
	c.ic = ic.NewIC()
	c.gp = gp.NewGP()
	c.sound = sound.NewSound()

	c.dmac = dmac.NewDMAC(c.mmu)
	c.gpu = gpu.NewGPU(c.ic, c.dmac,int16(scale),maxfps)
	c.dram = dram.NewDRAM()

	c.clock = clock.NewClock()
	c.clk_mul = 1

	if serial_p != "" {
		c.serial = serial.NewRealSerial(c.ic, serial_p)
	} else if listen != false || connect != "" {
		c.serial = serial.NewNetSerial(c.ic, listen, connect)
	} else {
		c.serial = serial.NewFakeSerial(c.ic)
	}

	///setup mmu conns
	setup_mmu_conn(c.timer, c.mmu)
	setup_mmu_conn(c.ic, c.mmu)
	setup_mmu_conn(c.gp, c.mmu)
	setup_mmu_conn(c.dram, c.mmu)
	setup_mmu_range_conn(c.dram, c.mmu)
	setup_mmu_conn(c.dmac, c.mmu)
	setup_mmu_conn(c.serial, c.mmu)
	setup_mmu_conn(c, c.mmu)
	setup_mmu_conn(c.sound, c.mmu)

	//gpu
	setup_mmu_conn(c.gpu, c.mmu)
	setup_mmu_range_conn(c.gpu, c.mmu)

	createOps(c)

	return c
}
