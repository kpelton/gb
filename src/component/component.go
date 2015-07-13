package component

type Register struct {
	Name string
	Addr uint16
}

type Range struct {
	Name string
	Addr_lo uint16
	Addr_hi uint16
}

type RegList []Register
type RangeList []Range

type MMIOComponent interface {
	Read_mmio(uint16) uint8
	Write_mmio(uint16, uint8)
	Get_reg_list() RegList
    Reset()
}


type MemComponent interface {
	Read(uint16) uint8
	Write(uint16, uint8)
	Get_range_list() RangeList
    Reset()
}


