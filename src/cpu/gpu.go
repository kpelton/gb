package cpu

import (
	"fmt"
    "github.com/banthar/Go-SDL/sdl"
)


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
    s.screen = sdl.SetVideoMode(160, 144, 32, sdl.HWSURFACE|sdl.DOUBLEBUF|sdl.ASYNCBLIT)    


	if s.screen == nil {
		fmt.Println(sdl.GetError())
	}

	sdl.WM_SetCaption("Monko's Gameboy", "")
    

 
}

func (s *Screen) PutPixel(x int16,y int16,color uint32) {
    //Old Method
	s.screen.FillRect(&sdl.Rect{x,y,1,1},color)
	//s.screen.Set(int(x),int(y),color)
	//pix := s.pixPtr(x, y)
	//pix.SetUint(color)
}


type GPU struct {
    screen *Screen
    t_screen *Screen
    LCDC uint8
    STAT uint8
    SCY uint8
    SCX uint8
    LYC uint8
    LY uint8
    WX uint8
    WY uint8
    BGP uint8

    currline uint8
    bg_tmap TileMap
	w_tmap TileMap
    
}

type sprite struct {
    y uint8
	x uint8
	num uint8
	fl_pal uint8
	fl_x_flip uint8 
	fl_y_flip uint8
	fl_pri uint8
}


func NewGPU() *GPU {
    g := new(GPU)
    g.screen = newScreen()
	g.t_screen = newScreen()
 

    return g
}

type Tile [8][8]uint8
type TileMap [32][32]Tile

func (g *GPU) get_tile_val(m *MMU,addr uint16) (Tile) {
        
    var k int16
    var j uint16
	var i uint8
    var tile Tile
 
    for k=0; k<8; k++ {
        var off = uint16(k)
        bl:=m.read_b(addr+(off*2))
		bh:=m.read_b(addr+1+(off*2))
		i=7
        for j=0; j<8; j++ {
            val:= uint8((bh>>(j) & 0x1) <<1)   | uint8((bl>>j & 0x1)) 
            tile[i][k] = val
			i--
        }
    }  
   return tile
}


func (g *GPU) output_pixel(val uint8, x uint16, y uint16) {

	switch (val) {
               case 1:
                    g.screen.PutPixel(int16(x),int16(y),uint32(0xaaaaaa))
                case 2:
                    g.screen.PutPixel(int16(x),int16(y),uint32(0x555555))
                case 3:
                    g.screen.PutPixel(int16(x),int16(y),uint32(0x0000000))
                case 0:
                    g.screen.PutPixel(int16(x),int16(y),uint32(0xffffff))

            }
}

func (g *GPU) print_tile(m *MMU,addr uint16,xoff uint16, yoff uint16,ytoff uint16,xflip bool) {
    var i int16
	var j uint16
    tile := g.get_tile_val(m,addr)

	if xflip  {
		j=0
		for i=7; i>=0; i-- {
			g.output_pixel(tile[i][ytoff],uint16(j)+xoff,yoff) 
			j++
		}
	}else{
		for i=0; i<8; i++ { 
			g.output_pixel(tile[i][ytoff],uint16(i)+xoff,yoff) 
		}
	}
}


func (g *GPU) get_tile_map(m *MMU)  {

    var map_base uint16
    var map_limit uint16
	//var w_map_base uint16
	//var w_map_limit uint16
	var i int
	var j int
    var tile_base uint16
    //var tile_limit uint16

    var tile uint16
    
    //Bit3 Tile map base
    if (g.LCDC & 0x08 == 0x08) {
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
	   // fmt.Println("WARNING!!!!!!!!!!!!!!!\n")
		//fmt.Println("!!!!!!!!!!!!!NOT IMPLEMENTED!!!!!!!!!!!!!!!\n")
        //tile_limit = 0x97FF
    }
/*
    //Bit3 Tile map base
    if (g.LCDC & 0x40 == 0x40) {
        w_map_base = 0x9c00
        w_map_limit = 0x9fff
    } else {
        w_map_base = 0x9800
        w_map_limit = 0x9Bff
    }
*/
  //b:=0
		for offset:=map_base; offset<=map_limit; offset++ {
  	    	b:=m.read_b(offset)
    		

    	    if tile_base == 0x8800 { 
    			//signed case
    		    if int8(b) >= 0 {
          	    	tile = 0x9000+ uint16(int(int8(b))*16)
                    //fmt.Println(int8(b))
            	    //fmt.Printf("%x,%x,\n",b,tile)
                 }else{
           		    tile = tile_base+ uint16((128+int(int8(b))) * 16)
                    //fmt.Println(int8(b))
                    //fmt.Println((128+int(int8(b))) )
                    //fmt.Printf("%x,%x,\n",b,tile)
                 }    
      
		    }else {
    			
    			//unsigned

    			tile = tile_base+(uint16(b)*16)
    		}			
    			g.bg_tmap[i][j] =g.get_tile_val(m,tile)

    			i++
    			if i == 32 {
    				i=0
    				j++
    			}
    //.b++	
    }
		
	i= 0
	j=0
/*
	if (g.LCDC & 0x10 == 0x10){

	
	for offset:=w_map_base; offset<=w_map_limit; offset++ {
		b:=m.read_b(offset)
			 if tile_base == 0x8800 { 
			//signed case
			if b > 127 {
				b -= 128
			} else {
				
			}
			
			tile = tile_base+(uint16((b)*16))
		}else {
			
			//unsigned
			tile = tile_base+(uint16(b)*16)
		}		
		g.w_tmap[i][j] =g.get_tile_val(m,tile)
		i++
		if i == 32 {
			i=0
			j++
		}
	}	
}
*/


}

func (g *GPU) print_tile_line(line uint,) {
    tile_line := (uint8(line)+g.SCY) & 7
    map_line := (uint8(line)+g.SCY) >>3 
    //tile_line := (uint8(line)) & 7
    //map_line := (uint8(line)) >>3 

	j:=g.SCX &7
    i:=g.SCX >>3
    for x:=0; x<160; {
        for j<8 {
            //fmt.Println(i,map_line,j,tile_line)
            switch (g.bg_tmap[i][map_line][j][tile_line]) {
                case 0: 
    	            g.screen.PutPixel(int16(x),int16(line),uint32(0xffffff))
                case 1:
                    g.screen.PutPixel(int16(x),int16(line),uint32(0xaaaaaa))
                case 2:
                    g.screen.PutPixel(int16(x),int16(line),uint32(0x555555))
                case 3:
                    g.screen.PutPixel(int16(x),int16(line),uint32(0x0000000))
    				
            }
            j++
			x++
        }
        j=0
        i=(1+i) &31
        
    }    
	
}
func (g *GPU) print_tile_line_w(line uint,) {
    tile_line := (uint8(line)+g.WY) & 7
    map_line := (uint8(line)+g.WX) >>3 
    j:=g.WX &7
    i:= g.WX >>3
    for x:=0; x<160; {
        
        for j<8 {
            switch (g.w_tmap[i][map_line][j][tile_line]) {

                case 1:
                    g.screen.PutPixel(int16(x),int16(line),uint32(0xc0c0c0))
                case 2:
                    g.screen.PutPixel(int16(x),int16(line),uint32(0x606060))
                case 3:
                    g.screen.PutPixel(int16(x),int16(line),uint32(0x00ff000))
            
            }
            j++
           x++
        }
         j=0
         i=(1+i) &31
        
    }    

}

func (g *GPU) print_sprites(m *MMU) {

	var sp sprite
	var j uint8
	var yoff uint8
	var ytoff uint8
	var xflip bool = false
	for i:=0; i<0xA0; i+=4 {
		//Main attributes
		sp.y = m.oam[i]
		if sp.y >155 {
			continue
		}
		sp.x = m.oam[i+1]
		//tile number is uint
		sp.num = m.oam[i+2]
		//Flags
		sp.fl_pri = m.oam[i+3] >>7
		sp.fl_y_flip = (m.oam[i+3] &0x40) >>6
		sp.fl_x_flip = (m.oam[i+3] &0x20) >>5
		sp.fl_pal = (m.oam[i+3] &0x10) >>4
		yoff = sp.y - 16

		
		if yoff >  g.LY-8 && yoff <=g.LY  {
			ytoff = (g.LY - yoff)
			if sp.fl_y_flip == 1 { ytoff = (^ytoff) & 0x07 }
			if sp.fl_x_flip == 1 { xflip = true }
			//fmt.Println(sp)
			g.print_tile(m,0x8000+(uint16(sp.num)*16),uint16((sp.x-8)+j),uint16(g.LY),uint16(ytoff),xflip)
		}
	}


}


func (g *GPU) print_tile_map(m *MMU) {
  

    if (g.LY==0) {g.get_tile_map(m)}
        //fmt.Println(g.tmap)
	if (g.LCDC & 0x81 == 0x81){
       g.print_tile_line(uint(g.LY))
		
	}
//	}

	if (g.LCDC & 0x10 == 0x10){
	//	g.LY++
	  //  g.print_tile_line_w(uint(g.LY))
		//H-BLANK
	//	g.STAT= 0x00
		
	}

	if (g.LCDC & 0x02 == 0x02){
	    g.print_sprites(m)
	}

	g.LY++
    g.STAT = 0x00

	//if g.LY == g.LYC {
	//	//set coincidenct flag
	//	g.STAT |= 0x2
	//	m.write_b(0xff0f,m.read_b(0xff0f)|0x02)  

//	}else{
//		//reset the flag
//		g.STAT &= 0xfd

//	}
        
    if (g.LY==153) {
		//V-BLANK
		g.STAT = 0x01
        //ASSERT vblank int
		m.cpu.ic.Assert(V_BLANK) 
		g.screen.screen.Flip()
		g.LY=0
	


	}

	}
   



