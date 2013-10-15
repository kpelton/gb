package dram

import "testing"


func Test_lo_ram(t *testing.T) {
	c :=  NewDRAM()

   for i:= 0; i<0x1000; i++  {
        c.Write_b(0xc000 +uint16(i),0x55)
    }
    for i:= 0; i<0x1000; i++  {
        if c.Read_b(0xc000 +uint16(i)) != 0x55  {
            t.Error("Failed to readback addr ",0xc000 +uint16(i))
        }
    }
}


func Test_z_ram(t *testing.T) {
	c :=  NewDRAM()

   for i:= 0; i<0x7f; i++  {
        c.Write_b(0xff80+uint16(i),0x55)
    }
    for i:= 0; i<0x7f; i++  {
        if c.Read_b(0xff80 +uint16(i)) != 0x55  {
            t.Error("Failed to readback Z_ram addr ",0xc000 +uint16(i))
        }
    }
}



func Test_hi_ram_banks(t *testing.T) {
  	c :=  NewDRAM()

 
    for i:= 0; i<0xff; i++  {
        test_hi_ram(c,uint8(i),0x55+uint8(i),t) 
    }
}

func test_hi_ram(c *DRAM,bank uint8, val uint8,t *testing.T) {
   // c.mmu.SVBK = bank
	c.Write_mmio(0xff70,bank)
    for i:= 0; i<0x1000; i++  {
        c.Write_b(0xd000 +uint16(i),val)
    }

    for i:= 0; i<0x1000; i++  {
        if c.Read_b(0xd000 +uint16(i)) != val  {
            t.Error("Failed to readback addr ",0xc000 +uint16(i))
        }
    }

}

func Test_mmio(t *testing.T) {
  	c :=  NewDRAM()

	for i:= 1; i<0x8; i++  {
	  c.Write_mmio(0xff70,uint8(i))
	  val := c.Read_mmio(0xff70)
	  if val == 0 {
		  t.Error("SVBK can never be 1",i)
	  } else if val != uint8(i) {
		  t.Error("invalid SVBK address",i %7,val)
	  }
    }
}

func Test_lo_ram_from_echo(t *testing.T) {
	c :=  NewDRAM()

   for i:= 0; i<0x1000; i++  {
        c.Write_b(0xe000 +uint16(i),0x55)
    }
    for i:= 0; i<0x1000; i++  {
        if c.Read_b(0xc000 +uint16(i)) != 0x55  {
            t.Error("Failed to readback addr ",0xc000 +uint16(i))
        }
    }
}
func Test_lo_ram_to_echo(t *testing.T) {
	c :=  NewDRAM()

   for i:= 0; i<0x1000; i++  {
        c.Write_b(0xc000 +uint16(i),0x55)
    }
    for i:= 0; i<0x1000; i++  {
        if c.Read_b(0xe000 +uint16(i)) != 0x55  {
            t.Error("Failed to readback addr ",0xc000 +uint16(i))
        }
    }
}


func Test_hi_ram_to_echo(t *testing.T) {
	c :=  NewDRAM()

   for i:= 0; i<0xe00; i++  {
        c.Write_b(0xd000 +uint16(i),0x55)
    }
    for i:= 0; i<0xe00; i++  {
        if c.Read_b(0xf000 +uint16(i)) != 0x55  {
            t.Error("Failed to readback addr ",i)
        }
    }
}
