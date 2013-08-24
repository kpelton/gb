package cpu

import (
	"banthar/sdl"
	"constants"
	"fmt"
	"time"
)

type Screen struct {
	screen *sdl.Surface
	rect   sdl.Rect
}

const (
	DARKEST  = 0x000000
	DARK     = 0x55555555
	LIGHT    = 0xaaaaaaaa
	LIGHTEST = 0xffffffff

	DARKEST_SEL   = 3
	DARK_SEL      = 2
	LIGHT_SEL     = 1
	LIGHTEST_SEL  = 0
	HBLANK_CYCLES = 204
	OAM_CYCLES    = 80
	RAM_CYCLES    = 172
    SCALE = 2
)

func newScreen() *Screen {
	s := new(Screen)
	s.initSDL()
	return s
}

func (s *Screen) initSDL() {

	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		fmt.Println(sdl.GetError())
	}

	s.screen = sdl.SetVideoMode(160*SCALE, 144*SCALE, 32, sdl.HWSURFACE|sdl.DOUBLEBUF|sdl.ASYNCBLIT)
	if s.screen == nil {
		fmt.Println(sdl.GetError())
	}

	sdl.WM_SetCaption("Monko's Gameboy", "")

}
func (s *Screen) PutPixel(x int16, y int16, color uint32) {
	s.rect.H = SCALE
	s.rect.W = SCALE
	s.rect.X = x * SCALE
	s.rect.Y = y * SCALE
	s.screen.FillRect(&s.rect, color)
}

type Palette [4]uint32
type Line [160]uint8

type GPU struct {
	screen             *Screen
	t_screen           *Screen
	LCDC               uint8
	STAT               uint8
	SCY                uint8
	SCX                uint8
	LYC                uint8
	LY                 uint8
	WX                 uint8
	WY                 uint8
	BGP                uint8
	OBP0               uint8
	OBP1               uint8
	mem_written        bool
	oam_cycle_count    uint16
	hblank_cycle_count uint16
	vblank_cycle_count uint16
	cycle_count        int16
	last_update        time.Time
	frame_time         time.Time
	rect   sdl.Rect

	currline  uint8
	bg_tmap   TileMap
	w_tmap    TileMap
	line_done uint8
	frames    uint16
	//palettes
	win_palette  Palette
	bg_palette   Palette
	obp0_palette Palette
	obp1_palette Palette

	last_lcdc uint8
}

type sprite struct {
	y         uint8
	x         uint8
	num       uint8
	fl_pal    uint8
	fl_x_flip uint8
	fl_y_flip uint8
	fl_pri    uint8
}

func (g *GPU) get_palette_color(selection uint8) uint32 {
	var retval uint32 = 0

	switch selection {
	case DARKEST_SEL:
		retval = DARKEST
	case DARK_SEL:
		retval = DARK
	case LIGHT_SEL:
		retval = LIGHT
	case LIGHTEST_SEL:
		retval = LIGHTEST
	default:
		panic("Can't convert this color!!!!!")
	}
	return retval
}
func (g *GPU) update_palette(p *Palette, val uint8) {

	p[0] = g.get_palette_color(val & 0x03)
	p[1] = g.get_palette_color((val & 0x0C) >> 2)
	p[2] = g.get_palette_color((val & 0x30) >> 4)
	p[3] = g.get_palette_color((val & 0xC0) >> 6)
}

func NewGPU() *GPU {
	g := new(GPU)
	g.screen = newScreen()
	g.t_screen = newScreen()
	g.mem_written = false
	g.last_update = time.Now()
	g.frame_time = time.Now()

	g.bg_palette[0] = LIGHTEST
	g.bg_palette[1] = LIGHT
	g.bg_palette[2] = DARK
	g.bg_palette[3] = DARKEST

	g.win_palette = g.bg_palette

	g.obp0_palette = g.bg_palette
	g.obp1_palette = g.bg_palette
	g.cycle_count = 456
	return g
}

type Tile [8][8]uint8
type Tile16 [8][16]uint8
type TileMap [32][32]Tile

func (g *GPU) get_tile_val(m *MMU, addr uint16) Tile {

	var k int16
	var j uint16
	var i uint8
	var tile Tile

	for k = 0; k < 8; k++ {
		var off uint16 = addr + uint16(k*2)
		bl := m.vm[off&0x1fff]
		bh := m.vm[(off+1)&0x1fff]
		i = 7
		for j = 0; j < 8; j++ {
			val := uint8((bh>>(j)&0x1)<<1) | uint8((bl >> j & 0x1))
			tile[i][k] = val
			i--
		}
	}
	return tile
}

func (g *GPU) get_tile_val16(m *MMU, addr uint16) Tile16 {

	var k int16
	var j uint16
	var i uint8
	var tile Tile16

	for k = 0; k < 16; k++ {
		var off uint16 = addr + uint16(k*2)
		bl := m.vm[off&0x1fff]
		bh := m.vm[(off+1)&0x1fff]
		i = 7
		for j = 0; j < 8; j++ {
			val := uint8((bh>>(j)&0x1)<<1) | uint8((bl >> j & 0x1))
			tile[i][k] = val
			i--
		}
	}
	return tile
}

func (g *GPU) output_pixel(val uint8, x uint16, y uint16) {

	g.screen.PutPixel(int16(x), int16(y), g.bg_palette[val])

}

func (g *GPU) output_pixel_sprite(val uint8, x uint16, y uint16, pal *Palette) {
	if val != 0 {
		g.screen.PutPixel(int16(x), int16(y), pal[val])
	}
}

func (g *GPU) print_tile(m *MMU, addr uint16, xoff uint16, yoff uint16, ytoff uint16, xflip bool) {
	var i uint8
	var j uint8
	tile := g.get_tile_val(m, addr)

	if xflip {
		j = 0
		for i = 7; i > 0; i-- {
           
			g.output_pixel(tile[i][ytoff], uint16(uint8(i)+uint8(xoff)), yoff)
			j++
		}
	} else {
		for i = 0; i < 8; i++ {

			g.output_pixel(tile[i][ytoff], uint16(uint8(i)+uint8(xoff)), yoff)
		}
	}
}

func (g *GPU) print_tile_sprite16(m *MMU, addr uint16, xoff uint16, yoff uint16, ytoff uint16, xflip bool,pri uint8, pal *Palette,line *Line) {
	var i int16
	var j uint8
	tile := g.get_tile_val16(m, addr)
    var x uint16
	if xflip {
		j = 0
		for i = 7; i >= 0; i-- {
            x = uint16(uint8(j)+uint8(xoff))
            if x <160 &&(pri == 0 || line[x] ==0){ 

			    g.output_pixel_sprite(tile[i][ytoff], x, yoff, pal)
            }
			j++
		}
	} else {
		for i = 0; i < 8; i++ {
            x = uint16(uint8(i)+uint8(xoff))
            if x <160 &&(pri == 0 || line[x] ==0){ 

			    g.output_pixel_sprite(tile[i][ytoff], x, yoff, pal)
            }

		}
	}
}

func (g *GPU) print_tile_sprite(m *MMU, addr uint16, xoff uint16, yoff uint16, ytoff uint16, xflip bool, pri uint8,pal *Palette,line *Line) {
	var i int16
	var j uint16
    var x uint16
	tile := g.get_tile_val(m, addr)

	if xflip {
		j = 0
		for i = 7; i >= 0; i-- {
            x = uint16(uint8(j)+uint8(xoff))
            if x <160 &&(pri == 0 || line[x] ==0){ 

			    g.output_pixel_sprite(tile[i][ytoff], uint16(uint8(j)+uint8(xoff)), yoff, pal)
            }
			j++
		}
	} else {
		for i = 0; i < 8; i++ {
            x = uint16(uint8(i)+uint8(xoff))
            if x <160 &&(pri == 0 || line[x] ==0){ 
			    g.output_pixel_sprite(tile[i][ytoff],x, yoff, pal)
            }
		}
	}
}

func (g *GPU) get_tile_map(m *MMU) {

	var map_base uint16
	var map_limit uint16
	var w_map_base uint16
	var w_map_limit uint16
	var i int
	var j int
	var tile_base uint16
	//var tile_limit uint16

	var tile uint16

	//Bit3 Tile map base
	if g.LCDC&0x08 == 0x08 {
		map_base = 0x9c00
		map_limit = 0x9fff
	} else {
		map_base = 0x9800
		map_limit = 0x9Bff
	}
	//Bit4 Tile data select
	if g.LCDC&0x10 == 0x10 {
		tile_base = 0x8000

		//tile_limit = 0x8FFF
	} else {
		tile_base = 0x8800
		// fmt.Println("WARNING!!!!!!!!!!!!!!!\n")
		//fmt.Println("!!!!!!!!!!!!!NOT IMPLEMENTED!!!!!!!!!!!!!!!\n")
		//tile_limit = 0x97FF
	}

	//Bit3 Tile map base
	if g.LCDC&0x40 == 0x40 {
		w_map_base = 0x9c00
		w_map_limit = 0x9fff
	} else {
		w_map_base = 0x9800
		w_map_limit = 0x9Bff
	}

	//b:=0
	for offset := map_base; offset <= map_limit; offset++ {
		b := m.vm[offset&0x1fff]
		if tile_base == 0x8800 {
			//signed case
			if int8(b) >= 0 {
				tile = 0x9000 + uint16(int(int8(b))*16)
				//fmt.Println(int8(b))
				//fmt.Printf("%x,%x,\n",b,tile)
			} else {
				tile = tile_base + uint16((128+int(int8(b)))*16)
				//fmt.Println(int8(b))
				//fmt.Println((128+int(int8(b))) )
				//fmt.Printf("%x,%x,\n",b,tile)
			}

		} else {

			//unsigned

			tile = tile_base + (uint16(b) * 16)
		}
		g.bg_tmap[i][j] = g.get_tile_val(m, tile)

		i++
		if i == 32 {
			i = 0
			j++
		}
		//.b++	
	}

	i = 0
	j = 0
	//fmt.Printf("0x%x\n",g.LCDC)
	if g.LCDC&0x20 == 0x20 {

		for offset := w_map_base; offset <= w_map_limit; offset++ {
		    b := m.vm[offset&0x1fff]

			//fmt.Printf("0x%x:0x%x\n",offset,b)
			if tile_base == 0x8800 {
				//signed case

				if int8(b) >= 0 {
					tile = 0x9000 + uint16(int(int8(b))*16)
				} else {
					tile = tile_base + uint16((128+int(int8(b)))*16)
				}

			} else {

				//unsigned
				tile = tile_base + (uint16(b) * 16)
			}
			g.w_tmap[i][j] = g.get_tile_val(m, tile)
			//fmt.Printf("0x%x:0x%x:0x%x\n",offset,b,tile)

			i++
			if i == 32 {
				i = 0
				j++
			}
		}
	}

}

func (g *GPU) print_tile_line(line uint, scanline *Line) {
	tile_line := (uint8(line) + g.SCY) & 7
	map_line := (uint8(line) + g.SCY) >> 3
	//tile_line := (uint8(line)) & 7
	//map_line := (uint8(line)) >>3 

	j := g.SCX & 7
	i := g.SCX >> 3
	for x := 0; x < 160; {
		for j < 8 {
			//fmt.Println(i,map_line,j,tile_line)

			val := (g.bg_tmap[i][map_line][j][tile_line])
			//g.screen.PutPixel(int16(x), int16(line), g.bg_palette[val])
            if x <160 {
                scanline[x] = val
            }
			j++
			x++
		}
		j = 0
		i = (1 + i) & 31

	}

}
func (g *GPU) print_tile_line_w(line uint, scanline *Line){
	var x uint8
	tile_line := (uint8(line) - g.WY) & 7
	map_line := (uint8(line) - g.WY) >> 3
	j := 0
	i := 0
	if g.WX < 7 {
		x = 0
	} else {
		x = g.WX - 7
	}
	for x < 166 {
		for j < 8 {
			val := g.w_tmap[i][map_line][j][tile_line]
			//g.screen.PutPixel(int16(x), int16(line), g.bg_palette[val])
                if x <160 {
                scanline[x] = val
            }
			j++
			x++
		}
		j = 0
		i = (1 + i) & 31

	}
}
func (g *GPU) print_sprites(m *MMU,line  *Line) {

	var sp sprite
	var j uint8
	var yoff uint8
	var ytoff uint8
	var xflip bool = false
	var mask uint8 = 0xff
	var yflip_mask uint8 = 0x7
	var pal *Palette = &g.obp0_palette

	var size uint8 = 8

	if g.LCDC&0x04 == 0x04 {
		mask = 0xfe
		yflip_mask = 0xf
		//on 8x16 least sig bit of num is 0

	}

	for i := 0x9f; i > 0; i -= 4 {
		//Main attributes
		sp.y = m.oam[i-3]
		pal = &g.obp0_palette

		if sp.y > 155 {
			continue
		}
		sp.x = m.oam[i-2]

		//tile number is uint

		if g.LCDC&0x04 == 0x04 {
			mask = 0xfe
			size = 16
		}

		sp.num = m.oam[i-1] & mask

		//Flags
		sp.fl_pri = m.oam[i] >> 7
		sp.fl_y_flip = (m.oam[i] & 0x40) >> 6
		sp.fl_x_flip = (m.oam[i] & 0x20) >> 5
		sp.fl_pal = (m.oam[i] & 0x10) >> 4

		yoff = sp.y - 16
		ytoff = (g.LY - yoff)
		//fmt.Println(sp,yoff,g.LY,ytoff)	

		xflip = false
		if ytoff < size {
			//fmt.Println(sp,yoff,g.LY,ytoff)	

			if sp.fl_y_flip == 1 {
				ytoff = (^ytoff & yflip_mask)
			}
			if sp.fl_pal == 1 {
				pal = &g.obp1_palette
			}

			if sp.fl_x_flip == 1 {
				xflip = true
			}
			if g.LCDC&0x04 == 0x04 {

				//	fmt.Println(sp)

				g.print_tile_sprite16(m, 0x8000+(uint16(sp.num)*16), uint16((sp.x-8)+j), uint16(g.LY), uint16(ytoff), xflip, sp.fl_pri, pal,line)
			} else {
				g.print_tile_sprite(m, 0x8000+(uint16(sp.num)*16), uint16((sp.x-8)+j), uint16(g.LY), uint16(ytoff), xflip,sp.fl_pri, pal,line)

			}
		}

	}

}
func (g *GPU) display_line(y int16,line *Line, pal *Palette) {
    g.rect.H = SCALE
	g.rect.W = SCALE
	g.rect.Y = y * SCALE
	var x int16
    for x=0; x<160; x++ {
	    g.rect.X = x * SCALE
        col:=pal[line[x]]
        for j:=x+1; j<160; j++ {
            col2:=pal[line[j]]
            if col2 != col {
               break
            }
            g.rect.W+=SCALE
            x++
        }
        g.screen.screen.FillRect(&g.rect, pal[line[x]])
    }
}
func (g *GPU) hblank(m *MMU, clocks uint16) {
	var line Line	


    if g.LY == 0 {
		g.get_tile_map(m)

	}
    
	if g.LCDC&0x81 == 0x81 {
		if g.last_lcdc&0x58 != g.LCDC&0x58 {
			g.get_tile_map(m)
			//	fmt.Println("REFRESH")
		}
		g.print_tile_line(uint(g.LY),&line)
		if g.LCDC&0x20 == 0x20 {
			if g.WX < 166 && g.LY >= g.WY {
				g.print_tile_line_w(uint(g.LY),&line)
			}
		}
            g.display_line(int16(g.LY),&line,&g.bg_palette)


		if g.LCDC&0x82 == 0x82 {
			g.print_sprites(m,&line)

		}

	}
	g.line_done = 1
	g.last_lcdc = g.LCDC
}

func (g *GPU) check_stat_int(m *MMU) {

	//Check LYC FLAg
	if g.LY == g.LYC {
		if g.STAT&0x04 != 0x04 && g.STAT&0x40 == 0x40 {
			m.cpu.ic.Assert(constants.LCDC)
			g.STAT |= 0x04

			//fmt.Println("Asserted lyc",g.LY,g.LYC)
		}

	} else {
		g.STAT &= ^uint8(0x4)
	}

	//	if g.STAT &0x010 == 0x10 && g.STAT &0x03 == 0 {
	//			m.cpu.ic.Assert(constants.LCDC)
	//         fmt.Println("Asserted hblank")

	//	}

}
func (g *GPU) check_stat_int_hblank(m *MMU) {

	if g.STAT&0x8 == 0x8 {
		m.cpu.ic.Assert(constants.LCDC)
		fmt.Println("Asserted hblank")

	}
}

func (g *GPU) vblank(m *MMU, clocks uint16) {

	if g.LY == 144 && g.STAT&0x1 == 0 {
		//V-BLANK

		g.STAT = (g.STAT & 0xfc) | 0x01
		//ASSERT vblank int
		m.cpu.ic.Assert(constants.V_BLANK)
		g.screen.screen.Flip()
		g.frames += 1
		if time.Since(g.frame_time) < time.Duration(17)*time.Millisecond {
			 time.Sleep((time.Duration(16700) * time.Microsecond) - time.Since(g.frame_time))
	//		 time.Sleep((time.Duration(1) * time.Microsecond) - time.Since(g.frame_time))

		}
		if time.Since(g.last_update) > time.Second {
			fmt.Println("FPS", int(g.frames))
			g.frames = 0
			g.last_update = time.Now()
		}
		g.frame_time = time.Now()

	}
	//fmt.Println(g.vblank_cycle_count)        

	//fmt.Println(g.vblank_cycle_count)        
	if g.LY > 153 {
		g.LY = 0
		g.line_done = 0
		g.cycle_count += 456
		//fmt.Println(g.cycle_count)        
		//	time.Sleep(time.Duration(5) * time.Millisecond)
	}

}

func (g *GPU) Update(m *MMU, clocks uint16) {

	if g.LCDC&0x80 == 0x80 {
		//	fmt.Printf("STAT:0x%04u\n",g.LY)

		if g.LY >= 144 {
			g.vblank(m, clocks)

		} else if g.cycle_count >= 456-OAM_CYCLES {
			g.STAT &= 0xfc
			g.STAT |= 2

		} else if g.cycle_count >= 456-OAM_CYCLES-RAM_CYCLES {
			g.STAT |= 3

		} else if g.cycle_count >= 456-HBLANK_CYCLES-OAM_CYCLES-RAM_CYCLES {
			g.STAT &= 0xfc

			if g.line_done == 0 {
				g.hblank(m, clocks)
				g.check_stat_int_hblank(m)

			}

		}
		if g.cycle_count <= 0 {
			//fmt.Println(g.cycle_count,g.LY)
			g.cycle_count += 456
			g.line_done = 0
			g.LY++
//			g.check_stat_int(m)

		}
		g.check_stat_int(m)

		g.cycle_count -= int16(clocks)
	} else {
		g.STAT &= 0xfc
		g.cycle_count = 456
		g.LY = 0
	}

}
