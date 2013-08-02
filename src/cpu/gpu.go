package cpu

import (
	"fmt"
	"github.com/banthar/Go-SDL/sdl"
	"time"
)

type Screen struct {
	screen *sdl.Surface
    rect sdl.Rect
}

const (
	DARKEST  = 0x000000
	DARK     = 0x55555555
	LIGHT    = 0xaaaaaaaa
	LIGHTEST = 0xffffffff

	DARKEST_SEL  = 3
	DARK_SEL     = 2
	LIGHT_SEL    = 1
	LIGHTEST_SEL = 0
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
	//s.screen = sdl.SetVideoMode(320, 288, 32, sdl.HWSURFACE|sdl.DOUBLEBUF|sdl.ASYNCBLIT)
	s.screen = sdl.SetVideoMode(640, 576, 32, sdl.HWSURFACE|sdl.DOUBLEBUF|sdl.ASYNCBLIT)
	//s.screen = sdl.SetVideoMode(1280, 1152, 32, sdl.HWSURFACE|sdl.DOUBLEBUF|sdl.ASYNCBLIT)

	if s.screen == nil {
		fmt.Println(sdl.GetError())
	}

	sdl.WM_SetCaption("Monko's Gameboy", "")

}
func (s *Screen) PutPixel(x int16, y int16, color uint32) {
    s.rect.H = 4
    s.rect.W = 4
    s.rect.X = x *4
    s.rect.Y = y *4
	s.screen.FillRect(&s.rect, color)
}

type Palette [4]uint32

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
	cycle_count        uint16
	last_update        time.Time
	currline           uint8
	bg_tmap            TileMap
	w_tmap             TileMap
	line_done          uint8
	frames             uint16
	//palettes
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

	g.bg_palette[0] = LIGHTEST
	g.bg_palette[1] = LIGHT
	g.bg_palette[2] = DARK
	g.bg_palette[3] = DARKEST

	g.obp0_palette[0] = LIGHTEST
	g.obp0_palette[1] = LIGHT
	g.obp0_palette[2] = DARK
	g.obp0_palette[3] = DARKEST

	g.obp1_palette[0] = LIGHTEST
	g.obp1_palette[1] = LIGHT
	g.obp1_palette[2] = DARK
	g.obp1_palette[3] = DARKEST

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
		var off = uint16(k)
		bl := m.read_b(addr + (off * 2))
		bh := m.read_b(addr + 1 + (off * 2))
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
		var off = uint16(k)
		bl := m.read_b(addr + (off * 2))
		bh := m.read_b(addr + 1 + (off * 2))
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

func (g *GPU) print_tile_sprite16(m *MMU, addr uint16, xoff uint16, yoff uint16, ytoff uint16, xflip bool, pal *Palette) {
	var i int16
	var j uint8
	tile := g.get_tile_val16(m, addr)
	if xoff < 7 {

	}
	if xflip {
		j = 0
		for i = 7; i >= 0; i-- {

			g.output_pixel_sprite(tile[i][ytoff], uint16(uint8(j)+uint8(xoff)), yoff, pal)
			j++
		}
	} else {
		for i = 0; i < 8; i++ {

			g.output_pixel_sprite(tile[i][ytoff], uint16(uint8(i)+uint8(xoff)), yoff, pal)
		}
	}
}

func (g *GPU) print_tile_sprite(m *MMU, addr uint16, xoff uint16, yoff uint16, ytoff uint16, xflip bool, pal *Palette) {
	var i int16
	var j uint16

	tile := g.get_tile_val(m, addr)

	if xflip {
		j = 0
		for i = 7; i >= 0; i-- {

			g.output_pixel_sprite(tile[i][ytoff], uint16(uint8(j)+uint8(xoff)), yoff, pal)
			j++
		}
	} else {
		for i = 0; i < 8; i++ {
			g.output_pixel_sprite(tile[i][ytoff], uint16(uint8(i)+uint8(xoff)), yoff, pal)
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
		b := m.read_b(offset)

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
			b := m.read_b(offset)

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
			g.w_tmap[i][j] = g.get_tile_val(m, tile)

			i++
			if i == 32 {
				i = 0
				j++
			}
		}
	}

}

func (g *GPU) print_tile_line(line uint) {
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
			g.screen.PutPixel(int16(x), int16(line), g.bg_palette[val])

			j++
			x++
		}
		j = 0
		i = (1 + i) & 31

	}

}
func (g *GPU) print_tile_line_w(line uint) {
	if g.WY <= 0 || g.WY >= 144 || uint8(line) < g.WY {

		return
	}
	tile_line := (uint8(line) - g.WY) & 7
	map_line := (uint8(line) - g.WY) >> 3
	j := 0
	i := 0
	fmt.Println(g.WX, g.WY, tile_line, map_line)
	for x := g.WX - 7; x < 166; {

		for j < 8 {

			val := g.w_tmap[i][map_line][j][tile_line]

			g.screen.PutPixel(int16(x), int16(line), g.bg_palette[val])

			j++
			x++
		}
		j = 0
		i = (1 + i) & 31

	}
}
func (g *GPU) print_sprites(m *MMU) {

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

	for i := 0; i < 0xA0; i += 4 {
		//Main attributes
		sp.y = m.oam[i]
		pal = &g.obp0_palette

		if sp.y > 155 {
			continue
		}
		sp.x = m.oam[i+1]

		//tile number is uint

		if g.LCDC&0x04 == 0x04 {
			mask = 0xfe
			size = 16
		}

		sp.num = m.oam[i+2] & mask

		//Flags
		sp.fl_pri = m.oam[i+3] >> 7
		sp.fl_y_flip = (m.oam[i+3] & 0x40) >> 6
		sp.fl_x_flip = (m.oam[i+3] & 0x20) >> 5
		sp.fl_pal = (m.oam[i+3] & 0x10) >> 4

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

				g.print_tile_sprite16(m, 0x8000+(uint16(sp.num)*16), uint16((sp.x-8)+j), uint16(g.LY), uint16(ytoff), xflip, pal)
			} else {
				g.print_tile_sprite(m, 0x8000+(uint16(sp.num)*16), uint16((sp.x-8)+j), uint16(g.LY), uint16(ytoff), xflip, pal)

			}
		}

	}

}

func (g *GPU) hblank(m *MMU, clocks uint16) {
	if g.LY == 0 {
		g.get_tile_map(m)

	} else {
		if g.last_lcdc&0x58 != g.LCDC&0x58 {
			g.get_tile_map(m)
			fmt.Println("REFRESH")
		}

	}

	if g.LCDC&0x81 == 0x81 {

		g.print_tile_line(uint(g.LY))

		if g.LCDC&0x20 == 0x20 {

			g.print_tile_line_w(uint(g.LY))
		}

		if g.LCDC&0x82 == 0x82 {
			g.print_sprites(m)

		}

	}
	g.LY++
	g.line_done = 1
	g.last_lcdc = g.LCDC
}

func (g *GPU) check_stat_int(m *MMU) {

	//Check LYC FLAg
	if g.LY == g.LYC {
		g.STAT |= 0x04
		if g.STAT&0x40 == 0x40 {
			m.cpu.ic.Assert(LCDC)

		}

	} else {
		g.STAT &= ^uint8(0x4)
	}
}
func (g *GPU) vblank(m *MMU, clocks uint16) {

	if g.LY == 144 && g.STAT&0x1 == 0 {
		//V-BLANK

		g.STAT = (g.STAT & 0xfc) | 0x01
		//ASSERT vblank int
		m.cpu.ic.Assert(V_BLANK)
		g.screen.screen.Flip()
		g.frames += 1
		if time.Since(g.last_update) > time.Second {
			fmt.Println("FPS", int(g.frames))
			g.frames = 0
			g.last_update = time.Now()
		}

	}
	g.vblank_cycle_count += clocks
	//fmt.Println(g.vblank_cycle_count)        

	if g.vblank_cycle_count >= 456 && g.LY <= 154 {
		g.vblank_cycle_count = 0
		g.LY += 1
		//fmt.Println(g.vblank_cycle_count)        
	} else if g.LY > 154 {
		g.vblank_cycle_count = 0
		g.cycle_count = 0
		g.LY = 0
		g.line_done = 0
		g.STAT &= 0xfc

	}

}

func (g *GPU) Update(m *MMU, clocks uint16) {
 
	if g.LCDC&0x80 == 0x80 {

		g.cycle_count += clocks 
		if g.LY >= 144 {
			g.vblank(m, clocks)
			g.check_stat_int(m)

		} else if g.cycle_count < 204 && g.line_done == 0 {
			g.STAT &= 0xfc
			g.hblank(m, clocks)
			g.check_stat_int(m)

		} else if g.STAT&0x2 != 0x2 && g.cycle_count >= 204 && g.cycle_count < 204+80 {
			g.STAT |= 2
		} else if g.STAT&0x3 != 0x3 && g.cycle_count >= 204+80 && g.cycle_count < 204+80+172 {
			g.STAT |= 3
		} else if g.cycle_count >= 204+80+172 {
			g.cycle_count = 0
			g.line_done = 0
		}

	} else {
		g.STAT &= 0xfc

		g.LY = 0
		g.cycle_count = 0
		g.vblank_cycle_count = 0

	}

}
