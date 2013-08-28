package serial

const (
	SB_ADDR = 0xff01
    SC_ADDR =0xff02
	HZ_8192_t    = 512

)

type Serial interface {
	Update () uint8
    Read   (addr uint16) uint8
    Write   (addr uint16,val uint8) 

}

