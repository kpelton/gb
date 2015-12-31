package carts

import (
	"fmt"
	"os"
)

////MBC1///////

type ROM_MBC1 struct {
	cart        [0x200000]uint8
	bank        uint16
	ram_enabled bool
	ram_bank    uint8
	ram         [0x20000]uint8
	memory_mode uint8
	file        os.File
	name        string
	has_battery bool
	dirty       bool
	count       uint32
	GenCart
}

func NewROM_MBC1(name string, cart_data []uint8, size int, has_battery bool) *ROM_MBC1 {
	m := new(ROM_MBC1)
	fmt.Println(size)
	copy(m.cart[:], cart_data)
	m.memory_mode = SIXTEEN_MB
	m.name = name
	m.has_battery = has_battery
	if has_battery == true {
		m.Load_ram()
	}
	return m
}
func (m *ROM_MBC1) Load_ram() {
	save_name := m.name + ".data"

	file, err := os.OpenFile(save_name, os.O_RDWR, 666) // For read access.
	if err != nil {
		//fmt.Println(err)
		file, err = os.Create(save_name) // For read access.
	} else {
		fmt.Println("Read save data")
		file.Read(m.ram[0:])
	}

	m.file = *file

}

func (m *ROM_MBC1) Save_ram() {
	if m.dirty == true {
		m.file.WriteAt(m.ram[0:], 0)
		m.dirty = false
	}
}

func (m *ROM_MBC1) Read(addr uint16) uint8 {
	var retval uint8

	if addr < 0x4000 {
		retval = m.cart[addr]
	} else if addr < 0x8000 {
		retval = m.cart[uint32(addr)+(uint32(m.bank)*0x4000)]
	} else {
		if (m.memory_mode == FOUR_MB) || (m.memory_mode == SIXTEEN_MB) {
			bank_offset := uint16(uint32(m.ram_bank) * 0x2000)
			fixed_addr := uint16(addr-0xa000) + bank_offset
			//fmt.Printf("RAM  BANK READ:%v  %04X->%04X:%x\n", m.ram_bank, addr, fixed_addr, retval)

			retval = m.ram[fixed_addr]
		} else {
			panic("Tried to read from ram that wasn't enabled!")
		}
	}
	m.count++
	if m.count >= 10000000 && m.has_battery && m.dirty {
		m.Save_ram()
		fmt.Println("Saving Ram")
		m.count = 0
	}
	return retval
}

func (m *ROM_MBC1) Write(addr uint16, val uint8) {
	if addr >= 0x2000 && addr < 0x4000 {
		if val > 1 {
			//fmt.Println("ROM Bank from",m.bank,val-1)
			m.bank = uint16((val) - 1)
		} else {
			m.bank = uint16(0)
		}
	} else if addr < 0x2000 {
		if m.memory_mode == FOUR_MB {
			if val == 0x0A {
				m.ram_enabled = true
				fmt.Println("RAM enabled", val)

			} else {
				fmt.Println("RAM Disabled", val)
				m.ram_enabled = false

			}
		}

	} else if addr >= 0x4000 && addr < 0x6000 {

		fmt.Println("RAM bank", "from", m.ram_bank, "to", val&0xf)
		m.ram_bank = (val & 0xf)
	} else if addr >= 0x6000 && addr < 0x7000 {
		if val > 0 {
			m.memory_mode = SIXTEEN_MB
			fmt.Println("RAM Memory mode", val, "selected")
		} else {
			m.memory_mode = FOUR_MB
		}

	} else if addr >= 0xA000 && addr < 0xc000 {

		if (m.memory_mode == FOUR_MB && m.ram_enabled == true) || (m.memory_mode == SIXTEEN_MB) {

			bank_offset := uint16(uint32(m.ram_bank) * 0x2000)
			fixed_addr := uint16(addr-0xa000) + bank_offset
			//fmt.Printf("RAM  BANK WRITE:%v  %04X->%04X:%x\n", m.ram_bank, addr, fixed_addr, val)

			m.ram[fixed_addr] = val
			if m.has_battery {
				m.dirty = true
			}
		} else {
			panic("Tried to read from ram that wasn't enabled!")
		}
	}

}
