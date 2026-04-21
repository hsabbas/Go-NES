package nes

import "log"

type axrom struct {
	c               *cartridge
	prgSelect       byte
	nametableSelect byte
}

func createAxromMapper(c *cartridge) *axrom {
	return &axrom{
		c:               c,
		prgSelect:       0,
		nametableSelect: 0,
	}
}

func (a *axrom) cpuRead(address uint16) byte {
	if address < 0x8000 {
		log.Println("Invalid cpu read to address", address)
		return 0
	}
	newAddress := int(address) - 0x8000
	newAddress += int(a.prgSelect) * kb32
	return a.c.prgRom[newAddress]
}

func (a *axrom) cpuWrite(address uint16, value byte) {
	if address < 0x8000 {
		log.Println("Invalid cpu write to address", address)
		return
	}

	a.prgSelect = value & 7
	if hasBit4(value) {
		a.c.mirroring = oneScreenUpper
	} else {
		a.c.mirroring = oneScreenLower
	}
}

func (a *axrom) ppuRead(address uint16) byte {
	if address >= 0x3F00 {
		log.Println("Invalid ppu read to address", address)
		return 0
	}

	if address < 0x2000 {
		return a.c.chr[address]
	}

	return a.c.readVram(address)
}

func (a *axrom) ppuWrite(address uint16, value byte) {
	if address >= 0x3F00 {
		log.Println("Invalid ppu write to address", address)
		return
	}

	if address < 0x2000 {
		a.c.chr[address] = value
		return
	}

	a.c.writeVram(address, value)
}

func (a *axrom) step() {}
