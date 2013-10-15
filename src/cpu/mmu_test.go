package cpu

import (
	"testing"
	"dram")


func Test_lo_ram(t *testing.T) {
	c := new(CPU)
	c.mmu = NewMMU(c)
	c.dram = dram.NewDRAM()

   for i:= 0; i<0x1000; i++  {
        c.mmu.write_b(0xc000 +uint16(i),0x55)
    }
    for i:= 0; i<0x1000; i++  {
        if c.mmu.read_b(0xc000 +uint16(i)) != 0x55  {
            t.Error("Failed to readback addr ",0xc000 +uint16(i))
        }
    }
}
func Test_z_ram(t *testing.T) {
	c := new(CPU)
	c.mmu = NewMMU(c)
	c.dram = dram.NewDRAM()

   for i:= 0; i<0x7f; i++  {
        c.mmu.write_b(0xff80+uint16(i),0x55)
    }
    for i:= 0; i<0x7f; i++  {
        if c.mmu.read_b(0xff80 +uint16(i)) != 0x55  {
            t.Error("Failed to readback Z_ram addr ",0xc000 +uint16(i))
        }
    }
}

func Test_hi_ram_banks(t *testing.T) {
    c := new(CPU)
    c.mmu = NewMMU(c)
 	c.dram = dram.NewDRAM()

    for i:= 0; i<0x1000; i++  {
        test_hi_ram(c,uint8(i),0x55+uint8(i),t) 
    }
}
func test_hi_ram(c *CPU,bank uint8, val uint8,t *testing.T) {
   // c.mmu.SVBK = bank
    c.mmu.write_b(0xff70,bank)

    for i:= 0; i<0x1000; i++  {
        c.mmu.write_b(0xd000 +uint16(i),val)
    }

    for i:= 0; i<0x1000; i++  {
        if c.mmu.read_b(0xd000 +uint16(i)) != val  {
            t.Error("Failed to readback addr ",0xc000 +uint16(i))
        }
    }

}
