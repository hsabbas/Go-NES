package nes

import "log"

type cnrom struct {
	c       *cartridge
	chrBank byte
}

func createCnromMapper(c *cartridge) *cnrom {
	return &cnrom{c: c}
}

func (c *cnrom) cpuRead(address uint16) byte {
	if address < 0x6000 {
		log.Println("Invalid cpu read to address", address)
		return 0
	}

	if address < 0x8000 {
		return c.c.prgRam[address-0x6000]
	}

	return c.c.prgRom[address-0x8000]
}

func (c *cnrom) cpuWrite(address uint16, value byte) {
	if address < 0x6000 {
		log.Println("Invalid cpu write to address", address)
		return
	}

	if address < 0x8000 {
		c.c.prgRam[address-0x6000] = value
	}

	c.chrBank = (value & 3) /*& c.c.prgRom[address-0x8000] Bus conflict?*/
}

func (c *cnrom) ppuRead(address uint16) byte {
	if address >= 0x3F00 {
		log.Println("Invalid ppu read to address", address)
		return 0
	}

	if address < 0x2000 {
		return c.c.chr[address+uint16(c.chrBank)*kb8]
	}

	return c.c.readVram(address)
}

func (c *cnrom) ppuWrite(address uint16, value byte) {
	if address >= 0x3F00 {
		log.Println("Invalid ppu write to address", address)
		return
	}

	if address < 0x2000 {
		c.c.chr[address+uint16(c.chrBank)*kb8] = value
	}

	c.c.writeVram(address, value)
}
