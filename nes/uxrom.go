package nes

import (
	"log"
)

type uxrom struct {
	c            *cartridge
	selectedBank byte
}

func createUxromMapper(c *cartridge) *uxrom {
	return &uxrom{
		c:            c,
		selectedBank: 0,
	}
}

func (u *uxrom) cpuRead(address uint16) byte {
	if address < 0x8000 {
		log.Println("Invalid cpu read to address", address)
		return 0
	}

	newAddress := int(address)
	if address >= 0xC000 {
		newAddress -= 0xC000
		newAddress += int(u.c.prgBanks-1) * kb16
	} else if address >= 0x8000 {
		newAddress -= 0x8000
		newAddress += int(u.selectedBank) * kb16
	}
	return u.c.prgRom[newAddress]
}

func (u *uxrom) cpuWrite(address uint16, value byte) {
	if address < 0x8000 {
		log.Println("Invalid cpu write to address", address)
		return
	}

	if address >= 0x8000 {
		u.selectedBank = value % byte(u.c.prgBanks)
	}
}

func (u *uxrom) ppuRead(address uint16) byte {
	if address >= 0x3F00 {
		log.Println("Invalid ppu read to address", address)
		return 0
	}

	if address < 0x2000 {
		return u.c.chr[address]
	}

	return u.c.readVram(address)
}

func (u *uxrom) ppuWrite(address uint16, value byte) {
	if address >= 0x3F00 {
		log.Println("Invalid ppu write to address", address)
		return
	}

	if address < 0x2000 {
		u.c.chr[address] = value
		return
	}

	u.c.writeVram(address, value)
}

func (u *uxrom) step() {}
