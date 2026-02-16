package nes

import (
	"log"
)

type nrom struct {
	c      *cartridge
	is16Kb bool
}

func createNromMapper(c *cartridge) *nrom {
	return &nrom{
		is16Kb: c.prgBanks == 1,
		c:      c,
	}
}

func (n *nrom) cpuRead(address uint16) byte {
	if address < 0x8000 {
		log.Println("Invalid cpu read to address", address)
		return 0
	}

	address -= 0x8000
	if n.is16Kb {
		address %= 0x4000
	}

	return n.c.prgRom[address]
}

func (n *nrom) cpuWrite(address uint16, value byte) {
	log.Println("invalid cpu write for NROM")
}

func (n *nrom) ppuRead(address uint16) byte {
	if address >= 0x3F00 {
		log.Println("Invalid ppu read to address", address)
		return 0
	}

	if address < kb8 {
		return n.c.chr[address]
	}

	return n.c.readVram(address)
}

func (n *nrom) ppuWrite(address uint16, value byte) {
	if address >= 0x3F00 {
		log.Println("Invalid ppu write to address", address)
		return
	}

	if address < 0x2000 {
		n.c.chr[address] = value
		return
	}

	n.c.writeVram(address, value)
}
