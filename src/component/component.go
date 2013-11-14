package component

type Register struct {
	Name string
	Addr uint16
}
type RegList []Register


type MMIOComponent interface {
	Read_mmio(uint16) uint8
	Write_mmio(uint16, uint8)
	Get_reg_list() RegList
}



type MemComponent interface {
	Read(uint16) uint8
	Write(uint16, uint8)
}


