package gpu

import "testing"

func TestGpuTiming(t *testing.T) {
	 := NewCpu()
	c.gpu.LCDC = 0x81
	c.gpu.LYC = 0xff
	for i := 0; i < 70224; i++ {
		c.gpu.Update(c.mmu, uint16(i%20))
		t.Log("AFTER STAT ", c.gpu.STAT, i, c.gpu.LY, c.gpu.cycle_count)

	}

	t.Log("AFTER LY ", c.gpu.LY)
	t.Log("AFTER STAT ", c.gpu.STAT)
	t.Log("AFTER cycle_count ", c.gpu.cycle_count)
}
/*
func TestGpuTimingOAM(t *testing.T) {
	c := NewCpu()
	c.gpu.LCDC = 0x81
	for i := 0; i < 220; i++ {
		c.gpu.Update(c.mmu, 1)
	}

	t.Log("AFTER LY ", c.gpu.LY)
	t.Log("AFTER STAT ", c.gpu.STAT)
	t.Log("AFTER cycle_count ", c.gpu.cycle_count)
}
/*
func TestGpuTimingHbl(t *testing.T) {
	c := NewCpu()
	c.gpu.LCDC = 0x81
	for i := 0; i < 203; i++ {
		c.gpu.Update(c.mmu, 1)
	}

	t.Log("AFTER LY ", c.gpu.LY)
	t.Log("AFTER STAT ", c.gpu.STAT)
	t.Log("AFTER cycle_count ", c.gpu.cycle_count)
}


func TestTimer(t *testing.T) {
    c := NewCpu()
    //Set TMA to on
    //check 262 hz
    //timer on with 262 clock
    c.timer.TAC = 0x5
    c.timer.TIMA = 0xEC
    t.Log("BEFORE",c.timer.TIMA)

    c.timer.Update(79*4)    
    t.Log(c.timer.last_update)
    t.Log("AFTER",c.timer.TIMA)


}
*/
