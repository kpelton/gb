package gpu

import (
	"banthar/sdl"
	"constants"
	"fmt"
	"ic"
	"time"
)

type Screen struct {
	screen *sdl.Surface
	rect   sdl.Rect
	scale  int16
}

var SCALE uint8 = 6

const (
	DARKEST  = 0x000000
	DARK     = 0xff555555
	LIGHT    = 0xffaaaaaa
	LIGHTEST = 0xffffffff

	DARKEST_SEL   = 3
	DARK_SEL      = 2
	LIGHT_SEL     = 1
	LIGHTEST_SEL  = 0
	HBLANK_CYCLES = 204
	OAM_CYCLES    = 80
	RAM_CYCLES    = 172
	fullspeed     = false
)

func newScreen(scale int16) *Screen {
	s := new(Screen)
	s.scale = scale

	s.initSDL()
	return s
}

func (s *Screen) initSDL() {

	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		fmt.Println(sdl.GetError())
	}

	s.screen = sdl.SetVideoMode(160*int(s.scale), 144*int(s.scale), 32, sdl.HWSURFACE|sdl.DOUBLEBUF|sdl.ASYNCBLIT)
	if s.screen == nil {
		fmt.Println(sdl.GetError())
	}

	sdl.WM_SetCaption("Monko's Gameboy", "")

}
func (s *Screen) PutPixel(x int16, y int16, color uint32) {
	s.rect.H = uint16(s.scale)
	s.rect.W = uint16(s.scale)
	s.rect.X = x * s.scale
	s.rect.Y = y * s.scale
	s.screen.FillRect(&s.rect, color)
}



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
	BC_index           uint8
	BCPS               uint8
	BCPD               uint8
	OC_index           uint8
	OCPS               uint8
	OCPD               uint8
	Vram               *VRAM



	mem_written        bool
	window_line        uint8
	oam_cycle_count    uint16
	hblank_cycle_count uint16
	vblank_cycle_count uint16
	cycle_count        int16
	last_update        time.Time
	frame_time         time.Time
	rect               sdl.Rect
	Cache              TileCache
	tile_vals          TileVals
	currline           uint8
	bg_tmap            [2]TileMap
	w_tmap             [2]TileMap
	bg_attr_map            TileAttr
	w_attr_map            TileAttr

	line_done          uint8
	frames             uint16
	//palettes
	bg_palette  GBPalette
	win_palette GBPalette
	gbc_palette [8]GBPalette
	gbc_oc_palette [8]GBPalette


	obp0_palette GBPalette
	obp1_palette GBPalette
	scale        int16
	last_lcdc    uint8
	lyc_int      uint8
	Gbc_mode     bool
	Oam          [0xA0]uint8
	ic           *ic.IC
	Pal_mem      [0x40] uint8
	Pal_oc_mem      [0x40] uint8

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

type sprite_gbc struct {
	y         uint8
	x         uint8
	num       uint8
	fl_x_flip uint8
	fl_y_flip uint8
	fl_pri    uint8
	bank      uint8
	pal       uint8
}

type bg_attr struct {
	pri  uint8
	h_flip  uint8
	v_flip uint8
	bank   uint8
	pal    uint8


}

type Tile [8][8]uint8
type Tile16 [8][16]uint8
type TileMap [32][32]Tile
type TileVals [0x180]Tile
type TileCache [0x180]*Tile
type TileAttr  [32][32] bg_attr
type GBPalette [4]uint32

type Line [160]uint8

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
func (g *GPU) UpdatePalette(p *GBPalette, val uint8) {

	p[0] = g.get_palette_color(val & 0x03)
	p[1] = g.get_palette_color((val & 0x0C) >> 2)
	p[2] = g.get_palette_color((val & 0x30) >> 4)
	p[3] = g.get_palette_color((val & 0xC0) >> 6)
}
func (g *GPU) UpdatePaletteBg(val uint8) {
	g.UpdatePalette(&g.bg_palette, val)

}
func (g *GPU) UpdatePaletteObp0(val uint8) {
	g.UpdatePalette(&g.obp0_palette, val)

}
func (g *GPU) UpdatePaletteObp1(val uint8) {
	g.UpdatePalette(&g.obp1_palette, val)

}

func (g *GPU) Read_mmio(addr uint16) uint8 {
	var val uint8
	switch addr {
	case 0xff40:
		val = g.LCDC
	case 0xff41:
		val = g.STAT
		//fmt.Printf("<-STAT:%04X\n", val)
	case 0xff42:
		val = g.SCY
	case 0xff43:
		val = g.SCX
	case 0xff44:
		val = g.LY
	case 0xff45:
		val = g.LYC
	case 0xff46:
		val = 0xff
	case 0xff47:
		val = g.BGP
	case 0xff48:
		val = g.OBP0
	case 0xff49:
		val = g.OBP1
	case 0xff4A:
		val = g.WY
	case 0xff4B:
		val = g.WX
	case VBANK_MMIO:
        val=g.Vram.Read_mmio(addr)
	case 0xff68:
		val = g.BCPS
		fmt.Printf("<-BCPS:%04X\n", val &0x1)
	case 0xff69:
		val = g.Pal_mem[g.BC_index]
		fmt.Printf("<-BCPD:%04X\n", val &0x1)
	case 0xff6A:
		fmt.Printf("<-OCPS:%04X\n", val &0x1)
	case 0xff6B:
		val = g.BCPS
	default:
		panic("unhandled read addr")
	}
	return val
}

func (g *GPU) Write_mmio(addr uint16,val uint8) {
	switch addr {

	case 0xff40:
		g.LCDC = val
	case 0xff41:
		g.STAT |= val & 0xf8
	case 0xff42:
		g.SCY = val
	case 0xff43:
		g.SCX = val
	case 0xff44:
		g.LY = 0		
	case 0xff45:
		g.LYC = val		
	case 0xff47:
		if val != g.BGP {
			g.BGP = val
            g.UpdatePaletteBg( val)
		}
	case 0xff48:
		if val != g.OBP0 {
			g.OBP0 = val
			g.UpdatePaletteObp0( val)
		}
	case 0xff49:
		if val != g.OBP1 {
			g.OBP1 = val
			g.UpdatePaletteObp1( val)
		}
	case 0xff4A:
		g.WY = val
	case 0xff4B:
		g.WX = val
	case VBANK_MMIO:
		g.Gbc_mode = true
		g.Vram.Write_mmio(addr,val)

	case 0xff68:
		g.BCPS = val
	//	fmt.Printf("->BCPS:%04X\n", val)
		g.BC_index = val & 0x3f
		

	case 0xff69:
		g.BCPD = val
	//	fmt.Printf("->BCPDIN:%04X %X  %d \n", val,g.STAT,g.BC_index,)
		g.Pal_mem[g.BC_index] = val
		if g.BCPS  & 0x80 == 0x80  {
			g.BC_index = (g.BC_index +1) %0x40
			g.BCPS  = 0x80 | 	g.BC_index 

		}
	case 0xff6A:
		g.OCPS = val
		fmt.Printf("->OCPS:%04X\n", val)
		g.OC_index = val & 0x3f
		

	case 0xff6B:
		g.OCPD = val
		fmt.Printf("->OCPDIN:%04X %X  %d \n", val,g.STAT,g.OC_index,)
		g.Pal_oc_mem[g.OC_index] = val
		if g.OCPS  & 0x80 == 0x80  {
			g.OC_index = (g.OC_index +1) %0x40
			g.OCPS  = 0x80 | 	g.BC_index 

		}
	default: 
		panic("Unhandled GPU mmio write")
	}
}



func (g *GPU) Update_paletteGBC(pal_mem  *[0x40]uint8, pal *[8]GBPalette) {
	offset:= 0
	var val uint32
	for i:=0; i<8; i++ {
		for j:=0; j<4; j++ {
			val = 0
			bcpd :=  (uint16(pal_mem[offset+1])<<8) | uint16(pal_mem[offset])
			red := uint32(bcpd &0x1f) *8
			green := uint32((bcpd & 0x3e0) >> 5) *8
			blue := uint32((bcpd & 0x7c00) >> 10) *8
		//	fmt.Printf("R:%02x G:%02x B:%02x BCBD:%04x PAL%x NUM%x\n",red,green,blue,bcpd,i,j) 
			val = 0xff << 24 | red <<16 | green <<8 | blue 
			pal[i][j] = val
			offset+=2
	
		}

	}
}


func NewGPU(ic *ic.IC, scale int16) *GPU {
	g := new(GPU)
	g.screen = newScreen(scale)
	g.Vram = newVRAM()
	g.last_update = time.Now()
	g.frame_time = time.Now()
	g.scale = scale
	g.bg_palette[0] = LIGHTEST
	g.bg_palette[1] = LIGHT
	g.bg_palette[2] = DARK
	g.bg_palette[3] = DARKEST

	g.win_palette = g.bg_palette

	g.obp0_palette = g.bg_palette
	g.obp1_palette = g.bg_palette
	g.cycle_count = 456
	g.ic = ic
	return g
}



func (g *GPU) get_tile_val(addr uint16, bank uint16,xflip uint8,yflip uint8) Tile {

	var k int16
	var j uint16
	var i uint8
	offset := (addr & 0x1ff0) >> 4

	//fmt.Printf("0x%x->0x%x\n",addr,((addr) &(0x1ff0))/16)
	//if g.Cache[offset] != nil {
	//	/	return *g.Cache[offset]
	//	}
	if yflip == 0{
		for k = 0; k < 8; k++ {
			var off uint16 = addr + uint16(k*2)
			bl := g.Vram.Vm[off&0x1fff + bank *0x2000]
			bh := g.Vram.Vm[(off+1)&0x1fff + bank * 0x2000]
			if xflip ==1 {
				i=0
				for j = 0; j< 8; j++{
					val := uint8((bh>>(j)&0x1)<<1) | uint8((bl >> j & 0x1))
					g.tile_vals[offset][i][k] = val
					i++
				}
			}else {
				i=7
				for j = 0; j < 8; j++ {
					val := uint8((bh>>(j)&0x1)<<1) | uint8((bl >> j & 0x1))
					g.tile_vals[offset][i][k] = val
					i--
				}
			}
		}
	
	}else {
		z:=7
		for k = 0; k <8 ; k++ {
			var off uint16 = addr + uint16(z*2)
			bl := g.Vram.Vm[off&0x1fff + bank *0x2000]
			bh := g.Vram.Vm[(off+1)&0x1fff + bank * 0x2000]
			if xflip ==1 {
				i=0
				for j = 0; j < 8; j++ {
					val := uint8((bh>>(j)&0x1)<<1) | uint8((bl >> j & 0x1))
					g.tile_vals[offset][i][k] = val
					i++
				}
			}else {
				i=7
				for j = 0; j < 8; j++ {
					val := uint8((bh>>(j)&0x1)<<1) | uint8((bl >> j & 0x1))
					g.tile_vals[offset][i][k] = val
					i--
				}
			}
			z--
		}
	}

	//g.Cache[offset] = &g.tile_vals[offset]
	return g.tile_vals[offset]
}

func (g *GPU) get_tile_val16(addr uint16,bank uint16) Tile16 {

	var k int16
	var j uint16
	var i uint8
	var tile Tile16

	for k = 0; k < 16; k++ {
		var off uint16 = addr + uint16(k*2)
		bl := g.Vram.Vm[off&0x1fff + bank *0x2000]
		bh := g.Vram.Vm[(off+1)&0x1fff + bank * 0x2000]
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

func (g *GPU) output_pixel_sprite(val uint8, x uint16, y uint16, pal *GBPalette) {
	if val != 0 {
		g.screen.PutPixel(int16(x), int16(y), pal[val])
	}
}


func (g *GPU) print_tile_sprite16(addr uint16, bank uint8,xoff uint16, yoff uint16, ytoff uint16, xflip bool, pri uint8, pal *GBPalette, line *Line) {
	var i int16
	var j uint8
	tile := g.get_tile_val16(addr,uint16(bank))
	var x uint16
	if xflip {
		j = 0
		for i = 7; i >= 0; i-- {
			x = uint16(uint8(j) + uint8(xoff))
			if x < 160 && (pri == 0 || line[x] == 0) {
				g.output_pixel_sprite(tile[i][ytoff], x, yoff, pal)
			}
			j++
		}
	} else {
		for i = 0; i < 8; i++ {
			x = uint16(uint8(i) + uint8(xoff))
			if x < 160 && (pri == 0 || line[x] == 0) {
				g.output_pixel_sprite(tile[i][ytoff], x, yoff, pal)
			}

		}
	}
}

func (g *GPU) print_tile_sprite(addr uint16, bank uint8, xoff uint16, yoff uint16, ytoff uint16, xflip bool, pri uint8, pal *GBPalette, line *Line) {
	var i int16
	var j uint16
	var x uint16
	tile := g.get_tile_val(addr,uint16(bank),0,0)

	if xflip {
		j = 0
		for i = 7; i >= 0; i-- {
			x = uint16(uint8(j) + uint8(xoff))
			if x < 160 && (pri == 0 || line[x] == 0) {
				g.output_pixel_sprite(tile[i][ytoff], uint16(uint8(j)+uint8(xoff)), yoff, pal)
			}
			j++
		}
	} else {
		for i = 0; i < 8; i++ {
			x = uint16(uint8(i) + uint8(xoff))
			if x < 160 && (pri == 0 || line[x] == 0) {
				g.output_pixel_sprite(tile[i][ytoff], x, yoff, pal)
			}
		}
	}
}

func (g *GPU) get_tile_map() {

	var map_base uint16 = 0x9800
	var map_limit uint16 = 0x9bff
	var w_map_base uint16 = 0x9900
	var w_map_limit uint16 = 0x9bff
	var tile_base uint16 = 0x8800

	//Bit3 Tile map base
	if g.LCDC&0x08 == 0x08 {
		map_base = 0x9c00
		map_limit = 0x9fff
	} 
	//Bit4 Tile data select
	if g.LCDC&0x10 == 0x10 {
		tile_base = 0x8000
	} 
	//Bit3 Tile map base
	if g.LCDC&0x40 == 0x40 {
		w_map_base = 0x9c00
		w_map_limit = 0x9fff
	} 


	//get attribute map
	if g.Gbc_mode == true {
		g.get_attr_map(map_base,map_limit,&g.bg_attr_map)
		g.get_tmap_gbc(map_base,map_limit,tile_base,0,&g.bg_tmap[0],&g.bg_attr_map)
		g.get_tmap_gbc(map_base,map_limit,tile_base,1,&g.bg_tmap[1],&g.bg_attr_map)
		if g.LCDC&0x20 == 0x20 {
			g.get_attr_map(w_map_base,w_map_limit,&g.w_attr_map)
			g.get_tmap_gbc(w_map_base,w_map_limit,tile_base,0,&g.w_tmap[0],&g.bg_attr_map)
			g.get_tmap_gbc(w_map_base,w_map_limit,tile_base,1,&g.w_tmap[1],&g.bg_attr_map)

		}
		g.Update_paletteGBC(&g.Pal_mem,&g.gbc_palette)
		g.Update_paletteGBC(&g.Pal_oc_mem,&g.gbc_oc_palette)

	}else {

		//get background tmap
		g.get_tmap(map_base,map_limit,tile_base,0,&g.bg_tmap[0])
		//if window is enabled get 
		if g.LCDC&0x20 == 0x20 {
			g.get_tmap(w_map_base,w_map_limit,tile_base,0,&g.w_tmap[0])
		}
	}

}
func (g* GPU) Up() {
	g.Update_paletteGBC(&g.Pal_mem,&g.gbc_palette)
}

func (g* GPU) get_tmap_gbc(map_base uint16, map_limit uint16,tile_base uint16,bank uint16, tmap *TileMap,amap *TileAttr) {
	var i int
	var j int
	var tile uint16

	for offset := map_base; offset <= map_limit; offset++ {
		b := g.Vram.Vm[offset&0x1fff]

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
		attr:=amap[i][j]

		tmap[i][j] = g.get_tile_val(tile,bank,attr.h_flip,attr.v_flip)

		i++
		if i == 32 {
			i = 0
			j++
		}
	}
}


func (g* GPU) get_tmap(map_base uint16, map_limit uint16,tile_base uint16,bank uint16, tmap *TileMap) {
	var i int
	var j int
	var tile uint16

	for offset := map_base; offset <= map_limit; offset++ {
			b := g.Vram.Vm[offset&0x1fff]

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
		tmap[i][j] = g.get_tile_val(tile,bank,0,0)

		i++
			if i == 32 {
				i = 0
				j++
			}
		}
}

func (g *GPU) get_attr_map(map_base uint16, map_limit uint16, attr * TileAttr)  {

	var j int
	var i int
	var abs_offset uint16
	var val uint8

	for offset := map_base; offset <= map_limit; offset++ {
		abs_offset = (offset & 0x1fff) + 0x2000 
		val = g.Vram.Vm[abs_offset]

		attr[i][j].pal = val &0x7
		attr[i][j].bank = (val &0x8) >>3
		attr[i][j].h_flip = (val &0x20)  >>5
		attr[i][j].v_flip = (val &0x40) >>6
		attr[i][j].pri  = val >>7
		
		if attr[i][j].pal != 0 {
			//fmt.Println("XXXX",g.bg_attr_map[i][j].pal)
		}
		i++
		if i == 32 {
			i = 0
			j++
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

			val := (g.bg_tmap[0][i][map_line][j][tile_line])
			//g.screen.PutPixel(int16(x), int16(line), g.bg_palette[val])
			if x < 160 {
				scanline[x] = val
			}
			j++
			x++
		}
		j = 0
		i = (1 + i) & 31

	}

}
func (g *GPU) print_tile_line_gbc(line uint, scanline *Line) {
	tile_line := (uint8(line) + g.SCY) & 7
	map_line := (uint8(line) + g.SCY) >> 3
	//tile_line := (uint8(line)) & 7
	//map_line := (uint8(line)) >>3 

	j := g.SCX & 7
	i := g.SCX >> 3
		g.Update_paletteGBC(&g.Pal_mem,&g.gbc_palette)

	for x := 0; x < 160; {
		for j < 8 {
			//fmt.Println(i,map_line,j,tile_line)
			attr:=g.bg_attr_map[i][map_line]

			val := (g.bg_tmap[attr.bank][i][map_line][j][tile_line])

			if x < 160 {
				//scanline[x] = val
				g.screen.PutPixel(int16(x), int16(line), g.gbc_palette[attr.pal][val])

			}
			j++
			x++
		}
		j = 0
		i = (1 + i) & 31

	}

}


func (g *GPU) print_tile_line_w(line uint, scanline *Line) {
	var x uint8
	tile_line := (g.window_line - g.WY) & 7
	map_line := (g.window_line - g.WY) >> 3
	j := 0
	i := 0
	if g.WX < 7 {
		x = 0
	} else {
		x = g.WX - 7
	}
	for x < 166 {
		for j < 8 {
				val := g.w_tmap[0][i][map_line][j][tile_line]

			if x < 160 {
				scanline[x] = val

			}				

			j++
			x++
		}
		j = 0
		i = (1 + i) & 31

	}
}

func (g *GPU) print_tile_line_w_gbc(line uint, scanline *Line) {
	var x uint8
	tile_line := (g.window_line - g.WY) & 7
	map_line := (g.window_line - g.WY) >> 3
	j := 0
	i := 0
	if g.WX < 7 {
		x = 0
	} else {
		x = g.WX - 7
	}
	for x < 166 {
		for j < 8 {
			attr:=g.w_attr_map[i][map_line]

			val := (g.w_tmap[attr.bank][i][map_line][j][tile_line])
			if x < 160 {
				g.screen.PutPixel(int16(x), int16(line), g.gbc_palette[attr.pal][val])

			}
			j++
			x++
		}
		j = 0
		i = (1 + i) & 31

	}
}


func (g *GPU) print_sprites(line *Line) {

	var sp sprite
	var j uint8
	var yoff uint8
	var ytoff uint8
	var xflip bool = false
	var mask uint8 = 0xff
	var yflip_mask uint8 = 0x7
	var pal *GBPalette = &g.obp0_palette

	var size uint8 = 8

	if g.LCDC&0x04 == 0x04 {
		mask = 0xfe
		yflip_mask = 0xf
		//on 8x16 least sig bit of num is 0

	}

	for i := 0x9f; i > 0; i -= 4 {
		//Main attributes
		sp.y = g.Oam[i-3]
		pal = &g.obp0_palette

		if sp.y > 155 {
			continue
		}
		sp.x = g.Oam[i-2]

		//tile number is uint

		if g.LCDC&0x04 == 0x04 {
			mask = 0xfe
			size = 16
		}

		sp.num = g.Oam[i-1] & mask

		//Flags
		sp.fl_pri = g.Oam[i] >> 7
		sp.fl_y_flip = (g.Oam[i] & 0x40) >> 6
		sp.fl_x_flip = (g.Oam[i] & 0x20) >> 5
		sp.fl_pal = (g.Oam[i] & 0x10) >> 4

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

				g.print_tile_sprite16(0x8000+(uint16(sp.num)*16),0,uint16((sp.x-8)+j), uint16(g.LY), uint16(ytoff), xflip, sp.fl_pri, pal, line)
			} else {
				g.print_tile_sprite(0x8000+(uint16(sp.num)*16),0,uint16((sp.x-8)+j), uint16(g.LY), uint16(ytoff), xflip, sp.fl_pri, pal, line)

			}
		}

	}

}

func (g *GPU) print_sprites_gbc(line *Line) {

	var sp sprite_gbc
	var j uint8
	var yoff uint8
	var ytoff uint8
	var xflip bool = false
	var mask uint8 = 0xff
	var yflip_mask uint8 = 0x7

	var size uint8 = 8

	if g.LCDC&0x04 == 0x04 {
		mask = 0xfe
		yflip_mask = 0xf
		//on 8x16 least sig bit of num is 0

	}

	for i := 0x9f; i > 0; i -= 4 {
		//Main attributes
		sp.y = g.Oam[i-3]

		if sp.y > 155 {
			continue
		}
		sp.x = g.Oam[i-2]

		//tile number is uint

		if g.LCDC&0x04 == 0x04 {
			mask = 0xfe
			size = 16
		}

		sp.num = g.Oam[i-1] & mask

		//Flags
		sp.fl_pri = g.Oam[i] >> 7
		sp.fl_y_flip = (g.Oam[i] & 0x40) >> 6
		sp.fl_x_flip = (g.Oam[i] & 0x20) >> 5
		sp.bank = (g.Oam[i] & 0x8) >> 3
		sp.pal = (g.Oam[i] & 0x07) 
		yoff = sp.y - 16
		ytoff = (g.LY - yoff)
		//fmt.Println(sp,yoff,g.LY,ytoff)	

		xflip = false
		if ytoff < size {
			//fmt.Println(sp,yoff,g.LY,ytoff)	

			if sp.fl_y_flip == 1 {
				ytoff = (^ytoff & yflip_mask)
			}


			if sp.fl_x_flip == 1 {
				xflip = true
			}
			if g.LCDC&0x04 == 0x04 {

				g.print_tile_sprite16(0x8000+(uint16(sp.num)*16),sp.bank,uint16((sp.x-8)+j), uint16(g.LY), uint16(ytoff), xflip, sp.fl_pri, &g.gbc_oc_palette[sp.pal], line)
			} else {
				g.print_tile_sprite(0x8000+(uint16(sp.num)*16), sp.bank,uint16((sp.x-8)+j), uint16(g.LY), uint16(ytoff), xflip, sp.fl_pri, &g.gbc_oc_palette[sp.pal], line)

			}
		}

	}

}



func (g *GPU) display_line(y int16, line *Line, pal *GBPalette) {
    //minimize draws to lines that have more than one color.
	g.rect.H = uint16(g.scale)
	g.rect.Y = y * g.scale
	var x int16
	for x = 0; x < 160; x++ {
		g.rect.X = x * g.scale
		g.rect.W = uint16(g.scale)
		col := pal[line[x]]
		for j := x + 1; j < 160; j++ {
			col2 := pal[line[j]]
			if col2 != col {
				break
			}
			g.rect.W += uint16(g.scale)
			x++

		}

		g.screen.screen.FillRect(&g.rect, pal[line[x]])
	}
}
func (g *GPU) hblank(clocks uint16) {
	var line Line

	if g.LY == 0 {
		g.get_tile_map()
 
	}
		//g.get_attr_map(0x9800,0x9bff,&g.bg_attr_map)

	if g.LCDC&0x81 == 0x81 {
		if g.last_lcdc&0x58 != g.LCDC&0x58 { //&& g.lyc_int != g.LY {
			g.get_tile_map()
			//fmt.Println("Refresh")
		}
		g.last_lcdc = g.LCDC
		if g.Gbc_mode == true {		
			//g.get_tile_map()

			g.print_tile_line_gbc(uint(g.LY), &line)
		//g.screen.screen.Flip()

		} else {
			g.print_tile_line(uint(g.LY), &line)

		}
		
		if g.LCDC&0x20 == 0x20 {

			if g.WX < 166 {
				if g.LY >= g.WY {
					if g.Gbc_mode == true {		
						g.print_tile_line_w_gbc(uint(g.LY), &line)
					} else {
						g.print_tile_line_w(uint(g.LY), &line)
					}
				
				}
				g.window_line++

			}
		}

		if g.Gbc_mode != true {	
			g.display_line(int16(g.LY), &line, &g.bg_palette)
		}
		if g.LCDC&0x82 == 0x82 {
			if g.Gbc_mode == true {	
				g.print_sprites_gbc(&line)
			}else{
				g.print_sprites(&line)	
			}
		}

	}

	g.line_done = 1
}

func (g *GPU) check_stat_int() {

	//Check LYC FLAg
	if g.LY == g.LYC {
		if g.STAT&0x04 != 0x04 && g.STAT&0x40 == 0x40 {
			g.ic.Assert(constants.LCDC)
			g.STAT |= 0x04
			g.lyc_int = g.LY
			//fmt.Printf("Asserted lyc 0x%x 0x%x",g.LY,g.LYC)
		}

	} else {
		g.STAT &= ^uint8(0x4)
	}

}
func (g *GPU) check_stat_int_hblank() {

	if g.STAT&0x8 == 0x8 {
		g.ic.Assert(constants.LCDC)
		//	fmt.Println("Asserted hblank")

	}
}

func (g *GPU) vblank(clocks uint16) {

	if g.LY == 144 && g.STAT&0x1 == 0 {
		//V-BLANK
		g.window_line = 0
		g.STAT = (g.STAT & 0xfc) | 0x01
		//ASSERT vblank int
		g.ic.Assert(constants.V_BLANK)
		g.screen.screen.Flip()
		g.frames += 1
		
		if !fullspeed {
			if time.Since(g.frame_time) < time.Duration(17)*time.Millisecond {
				time.Sleep((time.Duration(16700) * time.Microsecond) - time.Since(g.frame_time))
				//		 time.Sleep((time.Duration(1) * time.Microsecond) - time.Since(g.frame_time))

			}
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
//		fmt.Println(g.bg_attr_map)
		//fmt.Println(g.cycle_count)        
		//	time.Sleep(time.Duration(5) * time.Millisecond)
	}

}

func (g *GPU) Update(clocks uint16) {

	g.check_stat_int()
	if g.LCDC&0x80 == 0x80 {
		//	fmt.Printf("STAT:0x%04u\n",g.LY)

		if g.LY >= 144 {
			g.vblank(clocks)

		} else if g.cycle_count >= 456-OAM_CYCLES {
			g.STAT &= 0xfc
			g.STAT |= 2

		} else if g.cycle_count >= 456-OAM_CYCLES-RAM_CYCLES {
			g.STAT |= 3

		} else if g.cycle_count >= 456-HBLANK_CYCLES-OAM_CYCLES-RAM_CYCLES {
			g.STAT &= 0xfc

			if g.line_done == 0 {
				g.hblank(clocks)
				g.check_stat_int_hblank()

			}

		}
		if g.cycle_count <= 0 {
			//fmt.Println(g.cycle_count,g.LY)
			g.cycle_count += 456
			g.line_done = 0
			g.LY++

		}

		g.cycle_count -= int16(clocks)
	} else {
		g.STAT &= 0xfc
		g.cycle_count = 456
		g.LY = 0
	}

}
