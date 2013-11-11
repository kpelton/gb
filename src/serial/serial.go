package serial

const (
	SB_ADDR   = 0xff01
	SC_ADDR   = 0xff02
	HZ_8192_t = 512
)

type Serial interface {
	Update(cycles uint16) uint8
	Read_mmio(addr uint16) uint8
	Write_mmio(addr uint16, val uint8)
}
