package cpu

import (
	"fmt"
    "github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"
    //"time"
)

//GPU registers



type Screen struct {
    screen *sdl.Surface 
}

func newScreen() *Screen {
    s := new(Screen)
    s.initSDL()
    return s
}
    
func (s *Screen) initSDL() () {

    if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		fmt.Println(sdl.GetError())
	}
    s.screen = sdl.SetVideoMode(160, 144, 32, sdl.RESIZABLE)
    

			println("Couldn't open Joystick!")


	if s.screen == nil {
		fmt.Println(sdl.GetError())
	}


	var video_info = sdl.GetVideoInfo()

	fmt.Println("HW_available = ", video_info.HW_available)
	fmt.Println("WM_available = ", video_info.WM_available)
	fmt.Println("Video_mem = ", video_info.Video_mem, "kb")

	sdl.EnableUNICODE(1)

	sdl.WM_SetCaption("Monko's Gameboy", "")
    

 
}

func (s *Screen) PutPixel(x int16,y int16,color uint32) {
    s.screen.FillRect(&sdl.Rect{x,y,1,1},color)

}

func (s*Screen) DrawLoop() {
      for {
        s.screen.Flip()
    }


}

type GPU struct {
    screen *Screen
    LCDC uint8
    STAT uint8
    SCY uint8
    SCX uint8
    LYC uint8
    LY uint8
    currline uint8
    tmap TileMap
    
}

func newGPU() *GPU {
    g := new(GPU)
    g.screen = newScreen()
    return g
}

type Tile [8][8]uint8
type TileMap [32][32]Tile

func (g *GPU) get_tile_val(m *MMU,addr uint16) (Tile) {
        
    var k int16
    var jindex int16
    var j uint16
    var tile Tile
 
    for k=0; k<8; k++ {
        var off = uint16(k)
        bl:=m.read_b(addr+(off*2))
        br:=m.read_b(addr+(off*2)+1)
        for jindex=7; jindex>=0; jindex-- {
            j=uint16(jindex)    
            val:= (bl>>j & 0x1) |(br>>j& 0x1) >>1
            j=7-j 
            tile[j][k] = val
        }
    }  
   return tile
}




func (g *GPU) print_tile(m *MMU,addr uint16,xoff uint16, yoff uint16) {
    var i uint16
    var j uint16

    tile := g.get_tile_val(m,addr)
    for i=0; i<8; i++ {
        for j=0; j<8; j++ {
 
            switch (tile[j][i]) {
                case 1:
                    g.screen.PutPixel(int16(j+xoff),int16(i+yoff),uint32(0xc0c0c0))
                case 2:
                    g.screen.PutPixel(int16(j+xoff),int16(i+yoff),uint32(0x606060))
                case 3:
                    g.screen.PutPixel(int16(j+xoff),int16(i+yoff),uint32(0xff00000))
                default:
                    g.screen.PutPixel(int16(j+xoff),int16(i+yoff),uint32(0xff00000))

            }
        }
    }
}


func (g *GPU) get_tile_map(m *MMU)  {

    var map_base uint16
    var map_limit uint16

    var tile_base uint16
    //var tile_limit uint16

    var tile uint16
    
    //Bit3 Tile map base
    if (g.LCDC & 0x8 == 0x10) {
        map_base = 0x9c00
        map_limit = 0x9fff
    } else {
        map_base = 0x9800
        map_limit = 0x9Bff
    }
    //Bit4 Tile data select
    if (g.LCDC & 0x10 == 0x10) {
        tile_base = 0x8000
        //tile_limit = 0x8FFF
    } else {
        tile_base = 0x8800
        //tile_limit = 0x97FF
    }
    fmt.Printf("\nmap_base:0x%X\n",map_base)
    fmt.Printf("tile_base:0x%X\n",tile_base)
   
    i:=0
    j:=0
    for offset:=map_base; offset<=map_limit; offset++ {
        b:=m.read_b(offset)
        //only handles simple case
        tile = tile_base+(uint16(b)*16)
        g.tmap[i][j] =g.get_tile_val(m,tile)
        i++
        if i == 32 {
            i=0
            j++
        }
    }

}

func (g *GPU) print_tile_line(line uint,) {
    tile_line := (uint8(line)+g.SCY) & 7
    map_line := (uint8(line)+g.SCY) >>3 
    j:=g.SCX &7
    i:= g.SCX >>3
    for x:=0; x<160; {
        
        for j<8 {
            //fmt.Println(i,map_line,j,tile_line)
            switch (g.tmap[i][map_line][j][tile_line]) {
                case 0:
                    g.screen.PutPixel(int16(x),int16(line),uint32(0))

                case 1:
                    g.screen.PutPixel(int16(x),int16(line),uint32(0xc0c0c0))
                case 2:
                    g.screen.PutPixel(int16(x),int16(line),uint32(0x606060))
                case 3:
                    g.screen.PutPixel(int16(x),int16(line),uint32(0xff00000))
            
            }
            j++
           x++
        }
         j=0
         i=(1+i) &31
        
    }    

}
func (g *GPU) print_tile_map(m *MMU) {
  
   
    
    m.gpu.get_tile_map(m)
    //fmt.Println(g.tmap)
    for i:=0; i<144; i++ {
        g.print_tile_line(uint(i))
        g.LY++

    }
            g.screen.screen.Flip()

}
   



