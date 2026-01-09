package nes

import "log"

type axrom struct {
	prgRom          []byte
	prgBanks        byte
	prgSelect       byte
	chr             [kb8]byte
	nametableSelect byte

	vram [0x800]byte
}

func createAxromMapper(prgRom []byte, prg16kbBanks int) *axrom {
	return &axrom{
		prgRom:          prgRom,
		prgBanks:        byte(prg16kbBanks / 2),
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
	return a.prgRom[newAddress]
}

func (a *axrom) cpuWrite(address uint16, value byte) {
	if address < 0x8000 {
		log.Println("Invalid cpu write to address", address)
		return
	}

	a.prgSelect = value & 7
	a.nametableSelect = (value & bit4) >> 4
}

func (a *axrom) ppuRead(address uint16) byte {
	if address >= 0x3000 {
		log.Println("Invalid ppu read to address", address)
		return 0
	}

	if address < 0x2000 {
		return a.chr[address]
	}

	address %= 0x400
	address += uint16(a.nametableSelect) * 0x400
	return a.vram[address]
}

func (a *axrom) ppuWrite(address uint16, value byte) {
	if address >= 0x3000 {
		log.Println("Invalid ppu write to address", address)
		return
	}

	if address < 0x2000 {
		a.chr[address] = value
		return
	}

	address %= 0x400
	address += uint16(a.nametableSelect) * 0x400
	a.vram[address] = value
}
